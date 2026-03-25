package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/app_error"
	"gateway/pkg/jwt"
	"gateway/pkg/utils"
)

type DocumentService interface {
	UploadDocument(ctx context.Context, userID uint, file io.Reader, form dto.UploadDocumentForm) (*dto.DocumentResponse, error)
	GetDocument(ctx context.Context, docID uint) (*dto.DocumentResponse, error)
	GetEnrichStatus(ctx context.Context, docID uint) (string, error)
}

type DocumentHandler struct {
	documentService DocumentService
}

func NewDocumentHandler(documentService DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	claims, err := jwt.GetClaims(c.Request.Context())
	if err != nil {
		slog.Warn("UploadDocument: missing JWT claims", "err", err)
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var form dto.UploadDocumentForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	file, err := form.File.Open()
	if err != nil {
		slog.Error("UploadDocument: failed to open uploaded file", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.upload_document.read_failed")))
		return
	}
	defer file.Close()

	resp, err := h.documentService.UploadDocument(c.Request.Context(), claims.UserID, file, form)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrDocumentTooLarge):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.too_large")))
		case errors.Is(err, app_error.ErrDocumentInvalidType):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.unsupported_type")))
		default:
			slog.Error("UploadDocument service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.upload_document.failed")))
		}
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *DocumentHandler) GetEnrichStatus(c *gin.Context) {
	var uri dto.DocumentIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	status, err := h.documentService.GetEnrichStatus(c.Request.Context(), uri.DocID)
	if err != nil {
		slog.Error("GetEnrichStatus service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_enrich_status.failed")))
		return
	}
	c.JSON(http.StatusOK, dto.EnrichStatusResponse{DocID: uri.DocID, Status: status})
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	var uri dto.DocumentIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.documentService.GetDocument(c.Request.Context(), uri.DocID)
	if err != nil {
		if errors.Is(err, app_error.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.get_document.not_found")))
			return
		}
		slog.Error("GetDocument service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_document.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}
