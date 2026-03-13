package handler

import "github.com/gin-gonic/gin"

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Test(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Authenticated route accessed successfully!",
	})
}
