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

type SearchService interface {
	SearchDocuments(ctx context.Context, userID uint, q dto.SearchDocumentsQuery) (*dto.SearchDocumentsResponse, error)
	ListMyHistory(ctx context.Context, userID uint, limit int) (*dto.ListSearchHistoryResponse, error)
	ClearMyHistory(ctx context.Context, userID uint) (int64, error)
}

type SearchHandler struct {
	searchService SearchService
}

func NewSearchHandler(searchService SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) SearchDocuments(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var q dto.SearchDocumentsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.searchService.SearchDocuments(c.Request.Context(), userID, q)
	if err != nil {
		if errors.Is(err, app_error.ErrNotMember) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.search_documents.not_lab_member")))
			return
		}
		slog.Error("SearchDocuments service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.search_documents.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SearchHandler) ListMyHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var q dto.ListSearchHistoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.searchService.ListMyHistory(c.Request.Context(), userID, q.Limit)
	if err != nil {
		slog.Error("ListMyHistory service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.search_history.list_failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SearchHandler) ClearMyHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	deleted, err := h.searchService.ClearMyHistory(c.Request.Context(), userID)
	if err != nil {
		slog.Error("ClearMyHistory service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.search_history.clear_failed")))
		return
	}
	c.JSON(http.StatusOK, dto.DeleteSearchHistoryResponse{Deleted: deleted})
}
