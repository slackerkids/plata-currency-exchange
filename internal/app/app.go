package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/slackerkids/plata-currency-exchange.git/internal/config"
	"github.com/slackerkids/plata-currency-exchange.git/internal/service"
	"github.com/slackerkids/plata-currency-exchange.git/internal/worker"
)

const (
	maxConcurrentExchangeApiRequests = 100
)

type Application struct {
	conf           *config.Configuration
	postgresClient *pgxpool.Pool

	// Worker
	exchangeRateClient worker.ExchangeRateClient
	quoteRepository    worker.QuoteRepository

	// Service
	jobQueue service.JobQueue

	httpServer *http.Server
}

func New(ctx context.Context) (*Application, error) {
	app := &Application{}

	if err := app.setConfig(); err != nil {
		return nil, fmt.Errorf("set config: %w", err)
	}

	if err := app.setServer(); err != nil {
		return nil, fmt.Errorf("set server: %w", err)
	}

	if err := app.setRepositories(ctx); err != nil {
		return nil, fmt.Errorf("set repositories: %w", err)
	}

	if err := app.setMigrations(ctx); err != nil {
		return nil, fmt.Errorf("set migrations: %w", err)
	}

	if err := app.setWorkers(ctx); err != nil {
		return nil, fmt.Errorf("set workers: %w", err)
	}

	return app, nil
}

func (a *Application) setConfig() error {
	conf, err := config.New()
	if err != nil {
		return fmt.Errorf("new config: %w", err)
	}

	a.conf = conf
	return nil
}

func (a *Application) setServer() error {
	mux := newRouter()

	srv := &http.Server{
		Addr:    a.conf.HTTPServerAddress,
		Handler: mux,
	}

	a.httpServer = srv
	return nil
}

// mux router for http server
func newRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// 1. update quote
	// trigger background job goroutine to update prices of
	// base currency and quote currency and calculate exchange rate
	// 202: return id of job
	// 400: user send incorrect input (wrong currency and so on)
	// 405: handled by builtin router
	// 500: server problems (gateway, database connection...)
	mux.HandleFunc("POST /quote", http.NotFound)

	// 2. get quote by id
	// status of worker, if success return the result of the job
	// 200: status of the worker or exchange rate result and time
	// 404: id not exist in db
	// 405: handled by builtin router
	// 500: same as p1
	mux.HandleFunc("GET /quote/{id}", http.NotFound)

	// 3. get last value of quote
	// check the last entry from db
	// 200: exchange rate result and last updated time
	// 400: user send incorrect input (wrong currency and so on)
	// 404: exchange rate for given quote not found
	// 405: handled by builtin router
	// 500: same as p1
	mux.HandleFunc("GET /currency/{code}", http.NotFound)

	return mux
}

func (a *Application) setRepositories(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, a.conf.PostgresConnString)
	if err != nil {
		return fmt.Errorf("database connection: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}

	// repositories here...

	a.postgresClient = pool
	return nil
}

func (a *Application) setMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(a.postgresClient)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	if err := goose.Up(db, a.conf.MigrationsPath); err != nil {
		return fmt.Errorf("goose up migrations: %w", err)
	}

	slog.InfoContext(ctx, "migrations up successfully")
	return nil
}

func (a *Application) setWorkers(ctx context.Context) error {
	w := worker.NewPool(
		ctx,
		maxConcurrentExchangeApiRequests,
		a.quoteRepository,
		a.exchangeRateClient,
	)

	a.jobQueue = w
	return nil
}

func (a *Application) Start(ctx context.Context) error {
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server start: %w", err)
	}

	return nil
}

func (a *Application) Close(ctx context.Context) error {
	// shut down http server gracefully
	if a.httpServer != nil {
		if err := a.httpServer.Close(); err != nil {
			return fmt.Errorf("HTTP server close: %w", err)
		}
		slog.InfoContext(ctx, "server shutdown gracefully")
	}

	// shut down database connection
	if a.postgresClient != nil {
		a.postgresClient.Close()
	}

	return nil
}
