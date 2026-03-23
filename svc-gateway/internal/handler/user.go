// Handler for user authentication (login and registration)
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

type UserService interface {
	SendEmailCode(ctx context.Context, req dto.SendEmailCodeRequest) error
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
	UploadAvatar(ctx context.Context, userID uint, file io.Reader, contentType, filename string, size int64) (*dto.UploadAvatarResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req dto.UpdateProfileRequest) error
	GetAvatar(ctx context.Context, userID uint) (*dto.AvatarResponse, error)
	GetProfile(ctx context.Context, userID uint) (*dto.ProfileResponse, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) SendEmailCode(c *gin.Context) {
	var req dto.SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.userService.SendEmailCode(c.Request.Context(), req); err != nil {
		slog.Error("SendEmailCode service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.send_email_code.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("verification code sent successfully"))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	response, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		slog.Warn("Login failed", "err", err)
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("service.login.failed")))
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
		if errors.Is(err, app_error.ErrEmailCodeExpired) || errors.Is(err, app_error.ErrEmailCodeMismatch) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.email_code.invalid")))
			return
		}
		slog.Error("Register service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.register.failed")))
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
		if errors.Is(err, app_error.ErrEmailCodeExpired) || errors.Is(err, app_error.ErrEmailCodeMismatch) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.email_code.invalid")))
			return
		}
		slog.Error("ResetPassword service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.reset_password.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("password reset successfully"))
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	claims, err := jwt.GetClaims(c.Request.Context())
	if err != nil {
		slog.Warn("UploadAvatar: missing JWT claims", "err", err)
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var form dto.UploadAvatarForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	file, err := form.Avatar.Open()
	if err != nil {
		slog.Error("UploadAvatar: failed to open uploaded file", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.upload_avatar.read_failed")))
		return
	}
	defer file.Close()

	resp, err := h.userService.UploadAvatar(c.Request.Context(), claims.UserID,
		file, form.Avatar.Header.Get("Content-Type"), form.Avatar.Filename, form.Avatar.Size,
	)
	if err != nil {
		switch {
		case errors.Is(err, app_error.ErrAvatarTooLarge):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_avatar.too_large")))
		case errors.Is(err, app_error.ErrAvatarInvalidType):
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.upload_avatar.unsupported_type")))
		default:
			slog.Error("UploadAvatar service error", "err", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.upload_avatar.failed")))
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, err := jwt.GetClaims(c.Request.Context())
	if err != nil {
		slog.Warn("UpdateProfile: missing JWT claims", "err", err)
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("common.unauthorized")))
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), claims.UserID, req); err != nil {
		slog.Error("UpdateProfile service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.update_profile.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("profile updated successfully"))
}

func (h *UserHandler) GetAvatar(c *gin.Context) {
	var uri dto.UserIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.userService.GetAvatar(c.Request.Context(), uri.UserID)
	if err != nil {
		slog.Error("GetAvatar service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_avatar.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	var uri dto.UserIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.userService.GetProfile(c.Request.Context(), uri.UserID)
	if err != nil {
		slog.Error("GetProfile service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_profile.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}
