package handler

import (
	"github.com/gin-gonic/gin"
)

type LabHandler struct {
	// Add your LabService dependency here
}

func NewLabHandler() *LabHandler {
	return &LabHandler{}
}

func (h *LabHandler) CreateLab(c *gin.Context) {
	// TODO: Create Lab, user becomes owner
}

func (h *LabHandler) GetLab(c *gin.Context) {
	// TODO: Get Lab info
}

func (h *LabHandler) GetLabMembers(c *gin.Context) {
	// TODO: Get member list (must be a member)
}

func (h *LabHandler) DeleteLab(c *gin.Context) {
	// TODO: Delete lab (admin/owner only)
}

func (h *LabHandler) JoinLab(c *gin.Context) {
	// TODO: Join via invite code {"invite_code": "xxx"}
}

func (h *LabHandler) LeaveLab(c *gin.Context) {
	// TODO: Leave lab proactively
}

func (h *LabHandler) ResetInviteCode(c *gin.Context) {
	// TODO: Reset invite code (owner only)
}

func (h *LabHandler) RemoveMember(c *gin.Context) {
	// TODO: Remove member (owner only)
}

func (h *LabHandler) TransferOwnership(c *gin.Context) {
	// TODO: Transfer lab ownership to another member (owner only)
}
