package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/slackerkids/plata-currency-exchange.git/internal/config"
)

type Application struct {
	conf *config.Configuration

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
	srv := &http.Server{
		Addr:    a.conf.HTTPServerAddress,
		Handler: nil,
	}

	a.httpServer = srv
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
	// ...

	return nil
}
