package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/app_error"
	"gateway/pkg/utils"
)

type StatsService interface {
	GetMyDashboardStats(ctx context.Context, userID uint) (*dto.MyDashboardStatsResponse, error)
	GetLabDashboardStats(ctx context.Context, userID, labID uint) (*dto.LabDashboardStatsResponse, error)
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

func (h *StatsHandler) GetLabDashboardStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		slog.Warn("GetLabDashboardStats: missing user ID in context")
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")
	if labID == 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("invalid_lab_id")))
		return
	}

	resp, err := h.statsService.GetLabDashboardStats(c.Request.Context(), userID, labID)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrNotMember):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.get_lab_dashboard_stats.forbidden")))
			return
		}
		slog.Error("GetLabDashboardStats service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_lab_dashboard_stats.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}
