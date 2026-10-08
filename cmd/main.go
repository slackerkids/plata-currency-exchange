package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/slackerkids/plata-currency-exchange.git/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	application, err := app.New(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "app initialization", "error", err)
		os.Exit(1)
	}

	go func() {
		<-signalCh

		timeOutCtx, timeOutCancel := context.WithTimeout(ctx, time.Second*15)
		defer timeOutCancel()

		if err := application.Close(timeOutCtx); err != nil {
			slog.ErrorContext(timeOutCtx, "closing program", "error", err)
		}
		cancel()
	}()

	slog.InfoContext(ctx, "starting HTTP server")
	if err := application.Start(ctx); err != nil {
		slog.ErrorContext(ctx, "starting application", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
}
