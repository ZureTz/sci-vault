// Handler for user authentication (login and registration)
package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/jwt"
	"gateway/pkg/utils"
)

type UserService interface {
	SendEmailCode(ctx context.Context, req dto.SendEmailCodeRequest) error
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) // returns JWT token
	Register(ctx context.Context, req dto.RegisterRequest) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
	UploadAvatar(ctx context.Context, userID uint, file io.Reader, contentType, filename string, size int64) (*dto.UploadAvatarResponse, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) SendEmailCode(c *gin.Context) {
	// Send email verification code for registration
	var req dto.SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	// Call the userService to send the email code
	if err := h.userService.SendEmailCode(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("verification code sent successfully"))
}

// For login, registration, and password reset (without JWT authentication)

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	// Call the userService to perform login
	response, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.userService.Register(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(err))
		return
	}
	c.JSON(http.StatusCreated, utils.MessageResponse("user registered successfully"))
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("password reset successfully"))
}

// UploadAvatar is a protected route that requires JWT authentication
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	claims, err := jwt.GetClaims(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized: %w", err)))
		return
	}

	var form dto.UploadAvatarForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	file, err := form.Avatar.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("failed to read uploaded file")))
		return
	}
	defer file.Close()

	resp, err := h.userService.UploadAvatar(c.Request.Context(), claims.UserID,
		file, form.Avatar.Header.Get("Content-Type"), form.Avatar.Filename, form.Avatar.Size,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}
