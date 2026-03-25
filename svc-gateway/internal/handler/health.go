package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/pkg/grpc_client"
)

type HealthHandler struct {
	recommenderClient *grpc_client.RecommenderClient
}

func NewHealthHandler(recommenderClient *grpc_client.RecommenderClient) *HealthHandler {
	return &HealthHandler{
		recommenderClient: recommenderClient,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	services := gin.H{
		"status":  "ok",
		"service": "svc-gateway",
	}

	resp, err := h.recommenderClient.Health(c.Request.Context())
	if err != nil {
		services["svc-recommender"] = gin.H{"status": "unreachable", "error": err.Error()}
		c.JSON(http.StatusServiceUnavailable, services)
		return
	}

	services["svc-recommender"] = gin.H{
		"status":  resp.GetStatus(),
		"service": resp.GetService(),
	}
	c.JSON(http.StatusOK, services)
}
