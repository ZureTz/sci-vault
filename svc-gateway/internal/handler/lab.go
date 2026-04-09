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

func (h *LabHandler) JoinLabByCode(c *gin.Context) {
	// TODO: Join via invite code {"invite_code": "xxx"}
}

func (h *LabHandler) GetLab(c *gin.Context) {
	// TODO: Get Lab info
}

func (h *LabHandler) GetMembers(c *gin.Context) {
	// TODO: Get member list (must be a member)
}

func (h *LabHandler) LeaveLab(c *gin.Context) {
	// TODO: Leave lab proactively
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
