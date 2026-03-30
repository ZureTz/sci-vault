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
	"gateway/pkg/grpc_client"
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
	db                *gorm.DB
	cacheConn         *cache.CacheConnector
	storageClient     *storage.Client
	recommenderClient *grpc_client.RecommenderClient
	mailer            *mailer.Mailer
}

// New initializes all project dependencies, completes DI (Dependency Injection), and returns a ready-to-run App.
func New(configPath string) (*App, error) {
	// 0. Set up the global slog logger with defaults so errors before config load are visible
	logger.Setup("info", "text")

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Reconfigure logger with values from config
	logger.Setup(cfg.Log.Level, cfg.Log.Format)

	// 1. Initialize low-level dependencies (Databases, Redis, gRPC Clients)
	recommenderClient, err := grpc_client.NewRecommenderClient(cfg.RecommenderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommender gRPC client: %w", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	if err := database.Setup(db, &model.User{}, &model.UserProfile{}, &model.Document{}); err != nil {
		return nil, err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_embedding_hnsw ON documents USING hnsw (embedding vector_cosine_ops)`).Error; err != nil {
		return nil, fmt.Errorf("failed to create embedding hnsw index: %w", err)
	}

	storageClient := storage.NewClient(cfg.Storage.Endpoint, cfg.Storage.PresignEndpoint, cfg.Storage.AccessKey, cfg.Storage.SecretKey, cfg.Storage.PrivateBucket, cfg.Storage.PublicBucket, cfg.Storage.PublicProxyPath, cfg.Storage.PrivateProxyPath, cfg.Storage.UseSSL)
	if err := storageClient.EnsureBuckets(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure storage buckets: %w", err)
	}

	jwtGenerator := jwt.NewJWTGenerator(&cfg.JWT)

	cacheConn := cache.NewCacheConnector(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	mailSrv := mailer.NewMailer(cfg.Mailer.Host, cfg.Mailer.Port, cfg.Mailer.User, cfg.Mailer.Password)
	go mailSrv.Start()

	// 2. Initialize repositories layer (data access)
	userRepo := repo.NewUserRepo(db)
	userAvatarRepo := repo.NewUserProfileRepo(db)
	documentRepo := repo.NewDocumentRepo(db)

	// 3. Initialize services layer (business logic)
	userService := service.NewUserService(userRepo, userAvatarRepo, jwtGenerator, mailSrv, cacheConn, storageClient)
	documentService := service.NewDocumentService(documentRepo, storageClient, recommenderClient, cacheConn)

	// 4. Initialize handlers layer (HTTP/API)
	healthHandler := handler.NewHealthHandler(recommenderClient)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler()
	documentHandler := handler.NewDocumentHandler(documentService)

	// 5. Initialize router layer (routing and middleware mapping)
	r := router.NewRouter(&router.RouterDeps{
		HealthHandler:   healthHandler,
		UserHandler:     userHandler,
		AuthHandler:     authHandler,
		DocumentHandler: documentHandler,

		Config: cfg,
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
		db:                db,
		cacheConn:         cacheConn,
		storageClient:     storageClient,
		recommenderClient: recommenderClient,
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
