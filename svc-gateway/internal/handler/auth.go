package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/pkg/utils"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Test(c *gin.Context) {
	c.JSON(http.StatusOK, utils.MessageResponse("Authenticated route accessed successfully!"))
}
