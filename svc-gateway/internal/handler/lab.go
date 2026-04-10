package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/dto"
	"gateway/pkg/app_error"
	"gateway/pkg/utils"
)

type LabService interface {
	CreateLab(ctx context.Context, ownerID uint, req dto.CreateLabRequest) (*dto.JoinLabResponse, error)
	JoinLabByCode(ctx context.Context, userID uint, req dto.JoinLabByCodeRequest) (*dto.JoinLabResponse, error)
	GetMyLabs(ctx context.Context, userID uint) ([]dto.LabListItem, error)
	GetLab(ctx context.Context, labID, userID uint) (*dto.LabDetailResponse, error)
	GetMembers(ctx context.Context, labID, userID uint) ([]dto.LabMemberInfo, error)
	LeaveLab(ctx context.Context, labID, userID uint) error
}

type LabHandler struct {
	labService LabService
}

func NewLabHandler(labService LabService) *LabHandler {
	return &LabHandler{labService: labService}
}

func (h *LabHandler) CreateLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}

	var req dto.CreateLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.labService.CreateLab(c.Request.Context(), userID, req)
	if err != nil {
		slog.Error("CreateLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.create_lab.failed")))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *LabHandler) JoinLabByCode(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}

	var req dto.JoinLabByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.labService.JoinLabByCode(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, app_error.ErrInvalidInviteCode) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.join_lab.invalid_code")))
			return
		}
		if errors.Is(err, app_error.ErrAlreadyMember) {
			c.JSON(http.StatusConflict, utils.ErrorResponse(fmt.Errorf("service.join_lab.already_member")))
			return
		}
		slog.Error("JoinLabByCode service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.join_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *LabHandler) GetMyLabs(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}

	labs, err := h.labService.GetMyLabs(c.Request.Context(), userID)
	if err != nil {
		slog.Error("GetMyLabs service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_my_labs.failed")))
		return
	}
	c.JSON(http.StatusOK, labs)
}

func (h *LabHandler) GetLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	resp, err := h.labService.GetLab(c.Request.Context(), labID, userID)
	if err != nil {
		if errors.Is(err, app_error.ErrNotMember) || errors.Is(err, app_error.ErrLabNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.get_lab.not_found")))
			return
		}
		slog.Error("GetLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *LabHandler) GetMembers(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	members, err := h.labService.GetMembers(c.Request.Context(), labID, userID)
	if err != nil {
		if errors.Is(err, app_error.ErrNotMember) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.get_members.forbidden")))
			return
		}
		slog.Error("GetMembers service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.get_members.failed")))
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *LabHandler) LeaveLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	if err := h.labService.LeaveLab(c.Request.Context(), labID, userID); err != nil {
		if errors.Is(err, app_error.ErrNotMember) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.leave_lab.not_member")))
			return
		}
		if errors.Is(err, app_error.ErrOwnerCannotLeave) {
			c.JSON(http.StatusConflict, utils.ErrorResponse(fmt.Errorf("service.leave_lab.owner_cannot_leave")))
			return
		}
		slog.Error("LeaveLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.leave_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("left lab successfully"))
}

func (h *LabHandler) KickMember(c *gin.Context) {
	// TODO: Remove member (owner only)
}

func (h *LabHandler) TransferOwnership(c *gin.Context) {
	// TODO: Transfer lab ownership to another member (owner only)
}

// DANGEROUS: This will delete the lab and all associated data
// Only lab owner can do this.
func (h *LabHandler) DeleteLab(c *gin.Context) {
	// TODO: Delete lab (admin/owner only)
	// ADD CONFIRMATION STEP (e.g. type lab name to confirm), with email confirmation for extra safety
}

func (h *LabHandler) ResetInviteCode(c *gin.Context) {
	// TODO: Reset invite code (owner only)
}
