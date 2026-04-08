package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/jwt"
	"gateway/pkg/utils"
)

type StatsServiceInterface interface {
	GetDashboardStats(ctx context.Context, userID uint) (*dto.DashboardStatsResponse, error)
}

type StatsHandler struct {
	statsService StatsServiceInterface
}

func NewStatsHandler(statsService StatsServiceInterface) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	claims, err := jwt.GetClaims(c.Request.Context())
	if err != nil {
		slog.Warn("GetDashboardStats: missing JWT claims", "err", err)
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	resp, err := h.statsService.GetDashboardStats(c.Request.Context(), claims.UserID)
	if err != nil {
		slog.Error("GetDashboardStats service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_dashboard_stats.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}
