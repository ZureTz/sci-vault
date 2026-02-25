package router

import (
	"github.com/gin-gonic/gin"

	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/internal/producer"
)

// New creates and configures a gin Engine with all routes registered.
func New(p *producer.Producer) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handler.HealthCheck)

	healthProto := handler.NewHealthProtoHandler(p)
	r.GET("/health-protobuf", healthProto.Publish)

	v1 := r.Group("/api/v1")
	_ = v1

	return r
}
