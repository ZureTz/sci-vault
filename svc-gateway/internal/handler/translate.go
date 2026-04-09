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

type TranslateService interface {
	TranslateStream(ctx context.Context, text, targetLang string, onChunk func(chunk string) error) error
}

type TranslateHandler struct {
	translateService TranslateService
}

func NewTranslateHandler(translateService TranslateService) *TranslateHandler {
	return &TranslateHandler{translateService: translateService}
}

// TranslateSummary streams translated text back to the client as SSE events.
func (h *TranslateHandler) TranslateSummary(c *gin.Context) {
	var req dto.TranslateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	// We set SSE headers before attempting to call the service since we use a callback
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeaderNow()

	flusher, _ := c.Writer.(http.Flusher)

	err := h.translateService.TranslateStream(c.Request.Context(), req.Text, req.TargetLanguage, func(chunk string) error {
		_, err := fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
		if flusher != nil {
			flusher.Flush()
		}
		return err
	})

	if err != nil {
		slog.Error("TranslateSummary: translation stream error", "err", err)
		fmt.Fprintf(c.Writer, "event: error\ndata: translation failed\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		return
	}

	// Signal completion
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	if flusher != nil {
		flusher.Flush()
	}
}
