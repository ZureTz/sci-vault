package dto

type CreateLabRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

type JoinLabByCodeRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

type JoinLabResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	InviteCode  string  `json:"invite_code"`
	OwnerID     uint    `json:"owner_id"`
	MemberCount int64   `json:"member_count"`
}

type LabListItem struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	OwnerID     uint    `json:"owner_id"`
	MemberCount int64   `json:"member_count"`
	Role        string  `json:"role"`
}

type LabDetailResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	InviteCode  string  `json:"invite_code"`
	OwnerID     uint    `json:"owner_id"`
	MemberCount int64   `json:"member_count"`
	MyRole      string  `json:"my_role"`
}

type LabMemberInfo struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}
