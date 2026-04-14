// Package app_error defines sentinel errors shared across layers (service, handler).
// Handler code uses errors.Is against these values; service code wraps them.
package app_error

import "errors"

var (
	ErrEmailCodeExpired  = errors.New("email code expired or invalid")
	ErrEmailCodeMismatch = errors.New("email code mismatch")
	ErrAvatarTooLarge    = errors.New("avatar file too large")
	ErrAvatarInvalidType = errors.New("unsupported avatar image type")

	ErrDocumentTooLarge    = errors.New("document file too large")
	ErrDocumentInvalidType = errors.New("unsupported document type; only PDF is accepted")
	ErrDocumentNotFound    = errors.New("document not found")

	ErrLabNotFound       = errors.New("lab not found")
	ErrInvalidInviteCode = errors.New("invalid invite code")
	ErrAlreadyMember     = errors.New("already a member of this lab")
	ErrNotMember         = errors.New("user is not a member of this lab")
	ErrOwnerCannotLeave  = errors.New("owner cannot leave; transfer ownership first")
	ErrNotOwner          = errors.New("only the lab owner can perform this action")
	ErrCannotKickOwner   = errors.New("cannot remove the lab owner")
	ErrCannotKickSelf    = errors.New("cannot kick yourself; use leave instead")
	ErrTargetNotMember   = errors.New("target user is not a member of this lab")
	ErrLabNameMismatch   = errors.New("lab name does not match; deletion not confirmed")
)
