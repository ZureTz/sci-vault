package router

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"gateway/internal/grpcclient"
	"gateway/internal/handler"
	customValidator "gateway/pkg/validator"
)

type RouterDeps struct {
	// Handlers
	UserHandler *handler.UserHandler

	// gRPC clients
	RecommenderClient *grpcclient.RecommenderClient
}

// New creates and configures a gin Engine with all routes registered.
func NewRouter(deps RouterDeps) *gin.Engine {
	engine := gin.New()

	engine.Use(logger.SetLogger())
	engine.Use(gin.Recovery())

	// Register custom validators for username and password
	registerCustomValidators()

	// API versioning: all routes will be prefixed with /api/v1
	v1 := engine.Group("/api/v1")

	// Health check endpoint
	v1.GET("/health", handler.HealthCheck(deps.RecommenderClient))

	// Register auth routes
	registerAuthRoutes(v1.Group("/user"), deps.UserHandler)

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
func registerAuthRoutes(group *gin.RouterGroup, userHandler *handler.UserHandler) {
	group.POST("/login", userHandler.Login)
	group.POST("/register", userHandler.Register)
}
