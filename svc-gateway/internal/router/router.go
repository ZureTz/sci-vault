package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/pkg/cache"
	"gateway/pkg/logger"
	customValidator "gateway/pkg/validator"
)

type RouterDeps struct {
	// Handlers
	HealthHandler    *handler.HealthHandler
	UserHandler      *handler.UserHandler
	AuthHandler      *handler.AuthHandler
	DocumentHandler  *handler.DocumentHandler
	StatsHandler     *handler.StatsHandler
	TranslateHandler *handler.TranslateHandler

	// Cache connector
	CacheConn *cache.CacheConnector

	// Config
	Config *config.Config
}

// New creates and configures a gin Engine with all routes registered.
func NewRouter(deps *RouterDeps) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.LoggerWithFormatter(logger.GinLogger))
	engine.Use(gin.Recovery())

	// Register custom validators for username and password
	deps.registerCustomValidators()

	// API versioning: all routes will be prefixed with /api/v1
	v1 := engine.Group("/api/v1")

	// Health check endpoint
	v1.GET("/health", deps.HealthHandler.HealthCheck)

	// Register user routes
	deps.registerUserRoutes(v1.Group("/user"))

	// Protected routes (require JWT authentication)
	protected := v1.Group("")
	protected.Use(middleware.CheckJWT(&deps.Config.JWT))
	{
		deps.registerAuthenticatedRoutes(protected.Group("/auth"))
		deps.registerDocumentRoutes(protected.Group("/docs"))
		deps.registerStatsRoutes(protected.Group("/stats"))
		deps.registerTranslateRoutes(protected.Group("/translate"))
	}

	// Assign the configured engine to the router struct
	return engine
}

// Register custom validation functions for username and password
func (deps *RouterDeps) registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("custom_username_validator", customValidator.CustomUsernameValidator)
		v.RegisterValidation("custom_password_validator", customValidator.CustomPasswordValidator)
	}
}

// User login and registration routes (/api/v1/user)
func (deps *RouterDeps) registerUserRoutes(group *gin.RouterGroup) {
	// Send email verification code (rate limited: 1 req per min per email)
	group.POST("/send_email_code", middleware.StrictRateLimit(deps.CacheConn, "send_email_code", 1, time.Minute), deps.UserHandler.SendEmailCode)

	// For login and registration
	group.POST("/login", deps.UserHandler.Login)
	group.POST("/register", deps.UserHandler.Register)
	group.POST("/reset_password", deps.UserHandler.ResetPassword)

	// Protected user routes
	protected := group.Group("")
	protected.Use(middleware.CheckJWT(&deps.Config.JWT))
	{
		protected.POST("/upload_avatar", deps.UserHandler.UploadAvatar)
		protected.PUT("/profile", deps.UserHandler.UpdateProfile)
		protected.GET("/avatar/:user_id", deps.UserHandler.GetAvatar)
		protected.GET("/profile/:user_id", deps.UserHandler.GetProfile)
	}
}

// Authenticated routes (example: /api/v1/auth/...)
func (deps *RouterDeps) registerAuthenticatedRoutes(group *gin.RouterGroup) {
	group.GET("/test", deps.AuthHandler.Test)
}

// Document routes (/api/v1/docs/...)
func (deps *RouterDeps) registerDocumentRoutes(group *gin.RouterGroup) {
	group.POST("/upload", deps.DocumentHandler.UploadDocument)
	group.GET("/mine", deps.DocumentHandler.ListMyDocuments)
	group.GET("/pending", deps.DocumentHandler.ListPendingDocuments)
	group.GET("/:doc_id", deps.DocumentHandler.GetDocument)
	group.GET("/:doc_id/enrich_status", deps.DocumentHandler.GetEnrichStatus)
	group.POST("/:doc_id/restart_enrichment", deps.DocumentHandler.RestartEnrichment)
}

// Stats routes (/api/v1/stats/...)
func (deps *RouterDeps) registerStatsRoutes(group *gin.RouterGroup) {
	group.GET("/dashboard", deps.StatsHandler.GetDashboardStats)
}

// Translate routes (/api/v1/translate/...)
func (deps *RouterDeps) registerTranslateRoutes(group *gin.RouterGroup) {
	group.POST("/summary", deps.TranslateHandler.TranslateSummary)
}
