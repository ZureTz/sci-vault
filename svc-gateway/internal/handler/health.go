package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthService interface {
	CheckRecommender(ctx context.Context) (status string, service string, err error)
}

type HealthHandler struct {
	healthService HealthService
}

func NewHealthHandler(healthService HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	services := gin.H{
		"status":  "ok",
		"service": "svc-gateway",
	}

	status, srvName, err := h.healthService.CheckRecommender(c.Request.Context())
	if err != nil {
		services["svc-recommender"] = gin.H{"status": "unreachable", "error": err.Error()}
		c.JSON(http.StatusServiceUnavailable, services)
		return
	}

	services["svc-recommender"] = gin.H{
		"status":  status,
		"service": srvName,
	}
	c.JSON(http.StatusOK, services)
}
