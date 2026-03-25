package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/pkg/logger"
	customValidator "gateway/pkg/validator"
)

type RouterDeps struct {
	// Handlers
	HealthHandler   *handler.HealthHandler
	UserHandler     *handler.UserHandler
	AuthHandler     *handler.AuthHandler
	DocumentHandler *handler.DocumentHandler

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
	deps.registerUserRoutes(v1.Group("/user"), deps.UserHandler)

	// Protected routes (require JWT authentication)
	auth := v1.Group("/auth")
	auth.Use(middleware.CheckJWT(&deps.Config.JWT))
	deps.registerAuthenticatedRoutes(auth, deps.AuthHandler)

	docs := v1.Group("/docs")
	docs.Use(middleware.CheckJWT(&deps.Config.JWT))
	deps.registerDocumentRoutes(docs)

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
func (deps *RouterDeps) registerUserRoutes(group *gin.RouterGroup, userHandler *handler.UserHandler) {
	// Send email verification code
	group.POST("/send_email_code", userHandler.SendEmailCode)

	// For login and registration
	group.POST("/login", userHandler.Login)
	group.POST("/register", userHandler.Register)
	group.POST("/reset_password", userHandler.ResetPassword)

	// Protected user routes
	protected := group.Group("/")
	protected.Use(middleware.CheckJWT(&deps.Config.JWT))
	protected.POST("/upload_avatar", userHandler.UploadAvatar)
	protected.PUT("/profile", userHandler.UpdateProfile)
	protected.GET("/avatar/:user_id", userHandler.GetAvatar)
	protected.GET("/profile/:user_id", userHandler.GetProfile)
}

// Authenticated routes (example: /api/v1/auth/...)
func (deps *RouterDeps) registerAuthenticatedRoutes(group *gin.RouterGroup, authHandler *handler.AuthHandler) {
	group.GET("/test", authHandler.Test)
}

// Document routes (/api/v1/docs/...)
func (deps *RouterDeps) registerDocumentRoutes(group *gin.RouterGroup) {
	group.POST("/upload", deps.DocumentHandler.UploadDocument)
	group.GET("/:doc_id", deps.DocumentHandler.GetDocument)
	group.GET("/:doc_id/enrich_status", deps.DocumentHandler.GetEnrichStatus)
}
