package router

import (
	"github.com/gin-gonic/gin"

	"gateway/internal/grpcclient"
	"gateway/internal/handler"
	"gateway/internal/middleware"
)

// New creates and configures a gin Engine with all routes registered.
func New(rc *grpcclient.RecommenderClient) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handler.HealthCheck(rc))

	v1 := r.Group("/api/v1")
	_ = v1

	return r
}
