// Package app_error defines sentinel errors shared across layers (service, handler).
// Handler code uses errors.Is against these values; service code wraps them.
package app_error

import "errors"

var (
	ErrEmailCodeExpired  = errors.New("email code expired or invalid")
	ErrEmailCodeMismatch = errors.New("email code mismatch")
	ErrAvatarTooLarge    = errors.New("avatar file too large")
	ErrAvatarInvalidType = errors.New("unsupported avatar image type")
)
