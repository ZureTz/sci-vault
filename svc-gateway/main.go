package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"gateway/internal/config"
	"gateway/internal/grpcclient"
	"gateway/internal/handler"
	"gateway/internal/router"
)

// App holds all application dependencies and the HTTP server.
type App struct {
	// Configuration
	cfg    *config.Config
	engine *gin.Engine
	server *http.Server

	// gRPC client
	recommenderClient *grpcclient.RecommenderClient
}

// NewApp initializes all dependencies and returns a ready-to-run App.
func NewApp() (*App, error) {
	cfg := config.Load()

	recommenderClient, err := grpcclient.NewRecommenderClient(cfg.RecommenderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommender gRPC client: %w", err)
	}

	// TODO: Pass actual user service implementation to the handler
	userHandler := handler.NewUserHandler(nil)
	r := router.NewRouter(router.RouterDeps{
		UserHandler:       userHandler,
		RecommenderClient: recommenderClient,
	})

	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: r,
	}

	return &App{
		cfg:               cfg,
		engine:            r,
		server:            srv,
		recommenderClient: recommenderClient,
	}, nil
}

// Run starts the HTTP server. Blocks until the server stops.
func (a *App) Run() error {
	slog.Info("starting gateway", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// Close releases underlying resources (e.g. gRPC connections).
func (a *App) Close() {
	if err := a.recommenderClient.Close(); err != nil {
		slog.Warn("error closing recommender gRPC client", "err", err)
	}
}

func main() {
	app, err := NewApp()
	if err != nil {
		slog.Error("failed to initialize app", "err", err)
		os.Exit(1)
	}
	defer app.Close()

	go func() {
		if err := app.Run(); err != nil {
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

	if err := app.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}

	slog.Info("gateway stopped")
}
