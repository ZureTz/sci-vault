package handler

import (
	"gateway/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Test(c *gin.Context) {
	c.JSON(200, utils.MessageResponse("Authenticated route accessed successfully!"))
}
