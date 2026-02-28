// Handler for user authentication (login and registration)
package handler

import (
	// "context"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/model"
)

type UserService interface {
	// Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) // returns JWT token
	// Register(ctx context.Context, req model.RegisterRequest) error
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	// response, err := h.userService.Login(c.Request.Context(), req)
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
	// 	return
	// }
	c.JSON(http.StatusOK, gin.H{
		"user_id":  123, // response.UserID,
		"username": "1123", // response.Username,
		"token":	 "mock-jwt-token", // response.JWTToken,
	})
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// if err := h.userService.Register(c.Request.Context(), req); err != nil {
	// 	c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
	// 	return
	// }
	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}
