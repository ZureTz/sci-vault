package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"

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
	RequestLeaveLab(ctx context.Context, labID, userID uint) error
	LeaveLab(ctx context.Context, labID, userID uint, emailCode string) error
	KickMember(ctx context.Context, labID, requesterID, targetUserID uint) error
	TransferOwnership(ctx context.Context, labID, requesterID, targetUserID uint) error
	RequestDeleteLab(ctx context.Context, labID, requesterID uint) error
	DeleteLab(ctx context.Context, labID, requesterID uint, confirmName, emailCode string) error
	ResetInviteCode(ctx context.Context, labID, requesterID uint) (string, error)
	UpdateLabInfo(ctx context.Context, labID, requesterID uint, req dto.UpdateLabInfoRequest) (*dto.LabDetailResponse, error)
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

func (h *LabHandler) RequestLeaveLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	if err := h.labService.RequestLeaveLab(c.Request.Context(), labID, userID); err != nil {
		if errors.Is(err, app_error.ErrNotMember) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.leave_lab.not_member")))
			return
		}
		if errors.Is(err, app_error.ErrOwnerCannotLeave) {
			c.JSON(http.StatusConflict, utils.ErrorResponse(fmt.Errorf("service.leave_lab.owner_cannot_leave")))
			return
		}
		slog.Error("RequestLeaveLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.request_leave_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("confirmation code sent to your email"))
}

func (h *LabHandler) LeaveLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	var req dto.LeaveLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.labService.LeaveLab(c.Request.Context(), labID, userID, req.EmailCode); err != nil {
		if errors.Is(err, app_error.ErrNotMember) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.leave_lab.not_member")))
			return
		}
		if errors.Is(err, app_error.ErrOwnerCannotLeave) {
			c.JSON(http.StatusConflict, utils.ErrorResponse(fmt.Errorf("service.leave_lab.owner_cannot_leave")))
			return
		}
		if errors.Is(err, app_error.ErrEmailCodeExpired) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.leave_lab.code_expired")))
			return
		}
		if errors.Is(err, app_error.ErrEmailCodeMismatch) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.leave_lab.code_mismatch")))
			return
		}
		slog.Error("LeaveLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.leave_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("left lab successfully"))
}

func (h *LabHandler) KickMember(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	targetIDStr := c.Param("user_id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 64)
	if err != nil || targetID == 0 || targetID > math.MaxUint {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("invalid_user_id")))
		return
	}

	if err := h.labService.KickMember(c.Request.Context(), labID, userID, uint(targetID)); err != nil {
		if errors.Is(err, app_error.ErrNotOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.kick_member.forbidden")))
			return
		}
		if errors.Is(err, app_error.ErrCannotKickSelf) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.kick_member.cannot_kick_self")))
			return
		}
		if errors.Is(err, app_error.ErrCannotKickOwner) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.kick_member.cannot_kick_owner")))
			return
		}
		if errors.Is(err, app_error.ErrTargetNotMember) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.kick_member.target_not_member")))
			return
		}
		slog.Error("KickMember service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.kick_member.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("member removed successfully"))
}

func (h *LabHandler) TransferOwnership(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	var req dto.TransferOwnershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.labService.TransferOwnership(c.Request.Context(), labID, userID, req.TargetUserID); err != nil {
		if errors.Is(err, app_error.ErrNotOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.transfer_ownership.forbidden")))
			return
		}
		if errors.Is(err, app_error.ErrTargetNotMember) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.transfer_ownership.target_not_member")))
			return
		}
		slog.Error("TransferOwnership service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.transfer_ownership.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("ownership transferred successfully"))
}

func (h *LabHandler) RequestDeleteLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	if err := h.labService.RequestDeleteLab(c.Request.Context(), labID, userID); err != nil {
		if errors.Is(err, app_error.ErrNotOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.request_delete_lab.forbidden")))
			return
		}
		if errors.Is(err, app_error.ErrLabNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.request_delete_lab.not_found")))
			return
		}
		slog.Error("RequestDeleteLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.request_delete_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("confirmation code sent to your email"))
}

func (h *LabHandler) DeleteLab(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	var req dto.DeleteLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	if err := h.labService.DeleteLab(c.Request.Context(), labID, userID, req.ConfirmName, req.EmailCode); err != nil {
		if errors.Is(err, app_error.ErrNotOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.delete_lab.forbidden")))
			return
		}
		if errors.Is(err, app_error.ErrLabNameMismatch) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.delete_lab.name_mismatch")))
			return
		}
		if errors.Is(err, app_error.ErrEmailCodeExpired) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.delete_lab.code_expired")))
			return
		}
		if errors.Is(err, app_error.ErrEmailCodeMismatch) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("service.delete_lab.code_mismatch")))
			return
		}
		if errors.Is(err, app_error.ErrLabNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.delete_lab.not_found")))
			return
		}
		slog.Error("DeleteLab service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.delete_lab.failed")))
		return
	}
	c.JSON(http.StatusOK, utils.MessageResponse("lab deleted successfully"))
}

func (h *LabHandler) UpdateLabInfo(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	var req dto.UpdateLabInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err))
		return
	}

	resp, err := h.labService.UpdateLabInfo(c.Request.Context(), labID, userID, req)
	if err != nil {
		if errors.Is(err, app_error.ErrNotOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.update_lab_info.forbidden")))
			return
		}
		if errors.Is(err, app_error.ErrNotMember) || errors.Is(err, app_error.ErrLabNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(fmt.Errorf("service.update_lab_info.not_found")))
			return
		}
		slog.Error("UpdateLabInfo service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.update_lab_info.failed")))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *LabHandler) ResetInviteCode(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse(fmt.Errorf("unauthorized")))
		return
	}
	labID := c.GetUint("lab_id")

	newCode, err := h.labService.ResetInviteCode(c.Request.Context(), labID, userID)
	if err != nil {
		if errors.Is(err, app_error.ErrNotOwner) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(fmt.Errorf("service.reset_invite_code.forbidden")))
			return
		}
		slog.Error("ResetInviteCode service error", "err", err)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(fmt.Errorf("service.reset_invite_code.failed")))
		return
	}
	c.JSON(http.StatusOK, dto.ResetInviteCodeResponse{InviteCode: newCode})
}
