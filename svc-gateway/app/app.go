package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/internal/router"
	"gateway/internal/service"
	"gateway/pkg/cache"
	"gateway/pkg/database"
	"gateway/pkg/grpcclient"
	"gateway/pkg/jwt"
	"gateway/pkg/logger"
	"gateway/pkg/mailer"
	"gateway/pkg/storage"
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
	db                *gorm.DB
	cacheConn         *cache.CacheConnector
	storageClient     *storage.Client
	mailer            *mailer.Mailer
}

// New initializes all project dependencies, completes DI (Dependency Injection), and returns a ready-to-run App.
func New() (*App, error) {
	cfg := config.Load()

	// 0. Set up the global slog logger as early as possible
	logger.Setup(cfg.Log.Level, cfg.Log.Format)

	// 1. Initialize low-level dependencies (Databases, Redis, gRPC Clients)
	recommenderClient, err := grpcclient.NewRecommenderClient(cfg.RecommenderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommender gRPC client: %w", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	if err := db.AutoMigrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	storageClient := storage.NewClient(cfg.Storage.Endpoint, cfg.Storage.AccessKey, cfg.Storage.SecretKey, cfg.Storage.Bucket, cfg.Storage.UseSSL)
	if err := storageClient.EnsureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure storage bucket: %w", err)
	}

	jwtGenerator := jwt.NewJWTGenerator(&cfg.JWT)

	cacheConn := cache.NewCacheConnector(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	mailSrv := mailer.NewMailer(cfg.Mailer.Host, cfg.Mailer.Port, cfg.Mailer.User, cfg.Mailer.Password)
	go mailSrv.Start()

	// 2. Initialize repositories layer (data access)
	userRepo := repo.NewUserRepo(db)

	// 3. Initialize services layer (business logic)
	userService := service.NewUserService(userRepo, jwtGenerator, mailSrv, cacheConn)

	// 4. Initialize handlers layer (HTTP/API)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler()

	// 5. Initialize router layer (routing and middleware mapping)
	r := router.NewRouter(&router.RouterDeps{
		UserHandler: userHandler,
		AuthHandler: authHandler,

		Config: cfg,

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
		db:                db,
		cacheConn:         cacheConn,
		storageClient:     storageClient,
		mailer:            mailSrv,
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
	if sqlDB, err := a.db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			slog.Warn("error closing database connection", "err", err)
		}
	}
	if err := a.cacheConn.Close(); err != nil {
		slog.Warn("error closing cache connection", "err", err)
	}
}
