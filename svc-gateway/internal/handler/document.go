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
	"gateway/pkg/utils"
)

type DocumentService interface {
	UploadDocument(ctx context.Context, userID uint, file io.Reader, form dto.UploadDocumentForm) (*dto.DocumentResponse, error)
	BatchUploadDocuments(ctx context.Context, userID uint, form dto.BatchUploadDocumentForm) (*dto.BatchUploadDocumentResponse, error)
	GetDocument(ctx context.Context, userID, docID uint) (*dto.DocumentResponse, error)
	GetEnrichStatus(ctx context.Context, userID, docID uint) (string, error)
	ListMyDocuments(ctx context.Context, userID uint, page, pageSize int) (*dto.ListDocumentsResponse, error)
	ListPendingDocuments(ctx context.Context, userID uint) (*dto.ListDocumentsResponse, error)
	RestartEnrichment(ctx context.Context, userID, docID uint) error
	UpdateVisibility(ctx context.Context, docID, userID uint, req dto.UpdateVisibilityRequest) error
	BatchUpdateVisibility(ctx context.Context, userID uint, req dto.BatchUpdateVisibilityRequest) (int64, error)
	SearchDocuments(ctx context.Context, userID uint, q dto.SearchDocumentsQuery) (*dto.SearchDocumentsResponse, error)
}

type DocumentHandler struct {
	documentService DocumentService
}

func NewDocumentHandler(documentService DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		slog.Warn("UploadDocument: missing user ID in context")
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

	resp, err := h.documentService.UploadDocument(c.Request.Context(), userID, file, form)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrDocumentTooLarge):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.too_large")))
		case errors.Is(err, app_error.ErrDocumentInvalidType):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.unsupported_type")))
		case errors.Is(err, app_error.ErrDocumentDuplicate):
			c.JSON(http.StatusConflict, utils.ErrorResponse(fmt.Errorf("service.upload_document.duplicate")))
		case errors.Is(err, app_error.ErrLabRequiredForLabVis):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.lab_required")))
		case errors.Is(err, app_error.ErrInvalidVisibility):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.invalid_visibility")))
		case errors.Is(err, app_error.ErrNotMember):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.upload_document.not_lab_member")))
		default:
			slog.Error("UploadDocument service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.upload_document.failed")))
		}
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// batchUploadItemErrorCode maps a per-file sentinel error message (set by
// DocumentService.BatchUploadDocuments) to the stable i18n code the frontend
// renders. Unknown errors fall back to the generic failure code.
func batchUploadItemErrorCode(msg string) string {
	switch msg {
	case app_error.ErrDocumentTooLarge.Error():
		return "service.upload_document.too_large"
	case app_error.ErrDocumentInvalidType.Error():
		return "service.upload_document.unsupported_type"
	case app_error.ErrDocumentDuplicate.Error():
		return "service.upload_document.duplicate"
	default:
		return "service.upload_document.failed"
	}
}

func (h *DocumentHandler) BatchUploadDocuments(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var form dto.BatchUploadDocumentForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.documentService.BatchUploadDocuments(c.Request.Context(), userID, form)
	if err != nil {
		// Whole-batch failures share the single-upload error shape (visibility/lab resolution).
		switch {
		case errors.Is(err, app_error.ErrLabRequiredForLabVis):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.lab_required")))
		case errors.Is(err, app_error.ErrInvalidVisibility):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_document.invalid_visibility")))
		case errors.Is(err, app_error.ErrNotMember):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.upload_document.not_lab_member")))
		default:
			slog.Error("BatchUploadDocuments service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.upload_document.failed")))
		}
		return
	}

	// Translate per-file raw sentinel messages into i18n codes.
	for i := range resp.Results {
		if resp.Results[i].Error != "" {
			resp.Results[i].Error = batchUploadItemErrorCode(resp.Results[i].Error)
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentHandler) UpdateVisibility(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var uri dto.DocumentIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	var req dto.UpdateVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.documentService.UpdateVisibility(c.Request.Context(), uri.DocID, userID, req); err != nil {
		switch {
		case errors.Is(err, app_error.ErrLabRequiredForLabVis):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.update_visibility.lab_required")))
		case errors.Is(err, app_error.ErrInvalidVisibility):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.update_visibility.invalid_visibility")))
		case errors.Is(err, app_error.ErrNotMember):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.update_visibility.not_lab_member")))
		case errors.Is(err, app_error.ErrNotDocumentOwner):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.update_visibility.forbidden")))
		default:
			slog.Error("UpdateVisibility service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.update_visibility.failed")))
		}
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("service.update_visibility.success"))
}

func (h *DocumentHandler) BatchUpdateVisibility(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var req dto.BatchUpdateVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	updated, err := h.documentService.BatchUpdateVisibility(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrLabRequiredForLabVis):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.update_visibility.lab_required")))
		case errors.Is(err, app_error.ErrInvalidVisibility):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.update_visibility.invalid_visibility")))
		case errors.Is(err, app_error.ErrNotMember):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.update_visibility.not_lab_member")))
		case errors.Is(err, app_error.ErrSomeDocsNotAccessible):
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.update_visibility.partial_forbidden")))
		default:
			slog.Error("BatchUpdateVisibility service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.update_visibility.failed")))
		}
		return
	}
	c.JSON(http.StatusOK, dto.BatchUpdateVisibilityResponse{Updated: updated})
}

func (h *DocumentHandler) GetEnrichStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var uri dto.DocumentIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	status, err := h.documentService.GetEnrichStatus(c.Request.Context(), userID, uri.DocID)
	if err != nil {
		if errors.Is(err, app_error.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.get_enrich_status.not_found")))
			return
		}
		slog.Error("GetEnrichStatus service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_enrich_status.failed")))
		return
	}
	c.JSON(http.StatusOK, dto.EnrichStatusResponse{DocID: uri.DocID, Status: status})
}

func (h *DocumentHandler) ListMyDocuments(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		slog.Warn("ListMyDocuments: missing user ID in context")
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var query dto.ListMyDocumentsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	resp, err := h.documentService.ListMyDocuments(c.Request.Context(), userID, query.Page, query.PageSize)
	if err != nil {
		slog.Error("ListMyDocuments service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.list_documents.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentHandler) ListPendingDocuments(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	resp, err := h.documentService.ListPendingDocuments(c.Request.Context(), userID)
	if err != nil {
		slog.Error("ListPendingDocuments service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.list_documents.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var uri dto.DocumentIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.documentService.GetDocument(c.Request.Context(), userID, uri.DocID)
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

func (h *DocumentHandler) RestartEnrichment(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var uri dto.DocumentIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	err := h.documentService.RestartEnrichment(c.Request.Context(), userID, uri.DocID)
	if err != nil {
		if errors.Is(err, app_error.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.restart_enrichment.not_found")))
			return
		}
		if errors.Is(err, app_error.ErrNotDocumentOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.restart_enrichment.forbidden")))
			return
		}
		slog.Error("RestartEnrichment service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.restart_enrichment.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("service.restart_enrichment.success"))
}

func (h *DocumentHandler) SearchDocuments(c *gin.Context) {
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

	resp, err := h.documentService.SearchDocuments(c.Request.Context(), userID, q)
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
