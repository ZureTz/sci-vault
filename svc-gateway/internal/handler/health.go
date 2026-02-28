package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/grpcclient"
)

func HealthCheck(recommenderClient *grpcclient.RecommenderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		services := gin.H{
			"status":  "ok",
			"service": "svc-gateway",
		}

		resp, err := recommenderClient.Health(c.Request.Context())
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
}

