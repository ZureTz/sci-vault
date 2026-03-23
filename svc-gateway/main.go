package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"

	"gateway/app"
)

var args struct {
	Config string `arg:"-c,--config" help:"path to config file" default:""`
}

func main() {
	arg.MustParse(&args)

	application, err := app.New(args.Config)
	if err != nil {
		slog.Error("failed to initialize app", "err", err)
		os.Exit(1)
	}
	defer application.Close()

	go func() {
		if err := application.Run(); err != nil {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}

	slog.Info("gateway stopped")
}
