package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/grpc_client"
	"gateway/pkg/jwt"
	"gateway/pkg/utils"
)

type TranslateHandler struct {
	recommenderClient *grpc_client.RecommenderClient
}

func NewTranslateHandler(recommenderClient *grpc_client.RecommenderClient) *TranslateHandler {
	return &TranslateHandler{recommenderClient: recommenderClient}
}

// TranslateSummary streams translated text back to the client as SSE events.
func (h *TranslateHandler) TranslateSummary(c *gin.Context) {
	if _, err := jwt.GetClaims(c.Request.Context()); err != nil {
		slog.Warn("TranslateSummary: missing JWT claims", "err", err)
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var req dto.TranslateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	stream, err := h.recommenderClient.TranslateTextStream(c.Request.Context(), req.Text, req.TargetLanguage)
	if err != nil {
		slog.Error("TranslateSummary: failed to open gRPC stream", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.translate.failed")))
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeaderNow()

	flusher, _ := c.Writer.(http.Flusher)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			// Signal completion
			fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			return
		}
		if err != nil {
			slog.Error("TranslateSummary: stream recv error", "err", err)
			fmt.Fprintf(c.Writer, "event: error\ndata: translation failed\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			return
		}

		fmt.Fprintf(c.Writer, "data: %s\n\n", resp.Chunk)
		if flusher != nil {
			flusher.Flush()
		}
	}
}
