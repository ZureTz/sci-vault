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

type RecommendService interface {
	RecommendSimilar(ctx context.Context, userID, docID uint, q dto.RecommendSimilarQuery) (*dto.RecommendSimilarResponse, error)
}

type RecommendHandler struct {
	recommendService RecommendService
}

func NewRecommendHandler(recommendService RecommendService) *RecommendHandler {
	return &RecommendHandler{recommendService: recommendService}
}

func (h *RecommendHandler) RecommendSimilar(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	docID := c.GetUint("doc_id")

	var q dto.RecommendSimilarQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.recommendService.RecommendSimilar(c.Request.Context(), userID, docID, q)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrNotMember):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.recommend_similar.not_lab_member")))
		default:
			slog.Error("RecommendSimilar service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.recommend_similar.failed")))
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}
