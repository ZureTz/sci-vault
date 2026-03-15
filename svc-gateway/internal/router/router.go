package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/pkg/grpcclient"
	"gateway/pkg/logger"
	customValidator "gateway/pkg/validator"
)

type RouterDeps struct {
	// Handlers
	UserHandler *handler.UserHandler
	AuthHandler *handler.AuthHandler

	// Config
	Config *config.Config

	// gRPC clients
	RecommenderClient *grpcclient.RecommenderClient
}

// New creates and configures a gin Engine with all routes registered.
func NewRouter(deps *RouterDeps) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.LoggerWithFormatter(logger.GinLogger))
	engine.Use(gin.Recovery())

	// Register custom validators for username and password
	registerCustomValidators()

	// API versioning: all routes will be prefixed with /api/v1
	v1 := engine.Group("/api/v1")

	// Health check endpoint
	v1.GET("/health", handler.HealthCheck(deps.RecommenderClient))

	// Register user routes
	registerUserRoutes(v1.Group("/user"), deps.UserHandler)

	// Protected routes (require JWT authentication)
	auth := v1.Group("/auth")
	auth.Use(middleware.CheckJWT(&deps.Config.JWT))
	registerAuthenticatedRoutes(auth, deps.AuthHandler)

	// Assign the configured engine to the router struct
	return engine
}

// Register custom validation functions for username and password
func registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("custom_username_validator", customValidator.CustomUsernameValidator)
		v.RegisterValidation("custom_password_validator", customValidator.CustomPasswordValidator)
	}
}

// User login and registration routes (/api/v1/user)
func registerUserRoutes(group *gin.RouterGroup, userHandler *handler.UserHandler) {
	// Send email verification code
	group.POST("/send_email_code", userHandler.SendEmailCode)

	// For login and registration
	group.POST("/login", userHandler.Login)
	group.POST("/register", userHandler.Register)
	group.POST("/reset_password", userHandler.ResetPassword)
}

// Authenticated routes (example: /api/v1/auth/...)
func registerAuthenticatedRoutes(group *gin.RouterGroup, authHandler *handler.AuthHandler) {
	group.GET("/test", authHandler.Test)
}
