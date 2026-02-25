package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	pbhealth "gateway/internal/pb/health"
	"gateway/internal/producer"
)

// HealthProtoHandler holds the Kafka producer used by the endpoint.
type HealthProtoHandler struct {
	producer *producer.Producer
}

// NewHealthProtoHandler creates a handler backed by the given producer.
func NewHealthProtoHandler(p *producer.Producer) *HealthProtoHandler {
	return &HealthProtoHandler{producer: p}
}

// Publish godoc
//
//	@Summary	Health-protobuf probe
//	@Tags		health
//	@Produce	json
//	@Success	200
//	@Router		/health-protobuf [get]
func (h *HealthProtoHandler) Publish(c *gin.Context) {
	event := &pbhealth.HealthEvent{
		Status:    "ok",
		Service:   "svc-gateway",
		Timestamp: time.Now().UnixMilli(),
	}

	if err := h.producer.Publish(c.Request.Context(), "health", event); err != nil {
		slog.Error("health-protobuf: failed to publish", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "svc-gateway",
		"message": "protobuf event published",
	})
}
