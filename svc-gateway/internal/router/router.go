package router

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"

	"gateway/internal/grpcclient"
	"gateway/internal/handler"
)

// New creates and configures a gin Engine with all routes registered.
func New(recommenderClient *grpcclient.RecommenderClient) *gin.Engine {
	r := gin.New()

	r.Use(logger.SetLogger())
	r.Use(gin.Recovery())

	r.GET("/health", handler.HealthCheck(recommenderClient))

	v1 := r.Group("/api/v1")
	_ = v1

	return r
}
