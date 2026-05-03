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

type DocumentInteractionService interface {
	Like(ctx context.Context, userID, docID uint) (*dto.LikeStateResponse, error)
	Unlike(ctx context.Context, userID, docID uint) (*dto.LikeStateResponse, error)
	ListViewHistory(ctx context.Context, userID uint, q dto.ListHistoryQuery) (*dto.ListHistoryResponse, error)
	ListLikeHistory(ctx context.Context, userID uint, q dto.ListHistoryQuery) (*dto.ListHistoryResponse, error)
}

type DocumentInteractionHandler struct {
	service DocumentInteractionService
}

func NewDocumentInteractionHandler(service DocumentInteractionService) *DocumentInteractionHandler {
	return &DocumentInteractionHandler{service: service}
}

func (h *DocumentInteractionHandler) Like(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}
	docID := c.GetUint("doc_id")

	resp, err := h.service.Like(c.Request.Context(), userID, docID)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrInteractionDocNotFound):
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.like_document.not_found")))
		default:
			slog.Error("Like service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.like_document.failed")))
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentInteractionHandler) Unlike(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}
	docID := c.GetUint("doc_id")

	resp, err := h.service.Unlike(c.Request.Context(), userID, docID)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrInteractionDocNotFound):
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.like_document.not_found")))
		default:
			slog.Error("Unlike service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.like_document.failed")))
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentInteractionHandler) ListViewHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}
	var q dto.ListHistoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}
	resp, err := h.service.ListViewHistory(c.Request.Context(), userID, q)
	if err != nil {
		slog.Error("ListViewHistory service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.history.list_failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentInteractionHandler) ListLikeHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}
	var q dto.ListHistoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}
	resp, err := h.service.ListLikeHistory(c.Request.Context(), userID, q)
	if err != nil {
		slog.Error("ListLikeHistory service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.history.list_failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}
