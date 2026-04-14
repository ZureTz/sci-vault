package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/utils"
)

type StatsService interface {
	GetMyDashboardStats(ctx context.Context, userID uint) (*dto.MyDashboardStatsResponse, error)
}

type StatsHandler struct {
	statsService StatsService
}

func NewStatsHandler(statsService StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetMyDashboardStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		slog.Warn("GetMyDashboardStats: missing user ID in context")
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	resp, err := h.statsService.GetMyDashboardStats(c.Request.Context(), userID)
	if err != nil {
		slog.Error("GetMyDashboardStats service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_my_dashboard_stats.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}
