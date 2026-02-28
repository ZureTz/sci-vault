package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/config"
	"gateway/internal/grpcclient"
	"gateway/internal/handler"
	"gateway/internal/router"
)

// App holds all dependencies and the HTTP server for the gateway application.
// As the project grows, you can add database connections, Redis clients, Kafka producers, etc. here for centralized lifecycle management.
type App struct {
	// Basic configuration
	cfg    *config.Config
	engine *gin.Engine
	server *http.Server

	// External dependencies (Clients / Connectors)
	recommenderClient *grpcclient.RecommenderClient

	// Internal dependencies (Repos & Services)
	// userRepo repo.UserRepo
	// userService service.UserService
}

// New initializes all project dependencies, completes DI (Dependency Injection), and returns a ready-to-run App.
func New() (*App, error) {
	cfg := config.Load()

	// 1. Initialize low-level dependencies (Databases, Redis, gRPC Clients)
	recommenderClient, err := grpcclient.NewRecommenderClient(cfg.RecommenderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommender gRPC client: %w", err)
	}

	// 2. Initialize repositories layer (data access)

	// 3. Initialize services layer (business logic)

	// 4. Initialize handlers layer (HTTP/API)
	userHandler := handler.NewUserHandler(nil)

	// 5. Initialize router layer (routing and middleware mapping)
	r := router.NewRouter(router.RouterDeps{
		UserHandler:       userHandler,
		RecommenderClient: recommenderClient,
	})

	// 6. Build the HTTP server
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

// Run starts the HTTP server (blocking operation)
func (a *App) Run() error {
	slog.Info("starting gateway", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown performs graceful shutdown (stop accepting new requests, finish old requests in this context before exit)
func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// Close releases all basic resources held by App, such as database connections, gRPC connections, cache connections, etc.
func (a *App) Close() {
	if err := a.recommenderClient.Close(); err != nil {
		slog.Warn("error closing recommender gRPC client", "err", err)
	}
}
