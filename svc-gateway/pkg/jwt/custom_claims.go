package jwt

import (
	"context"
	"errors"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	ShouldReject bool   `json:"should_reject"`
}

// Validate errors out if `ShouldReject` is true.
func (c *CustomClaims) Validate(ctx context.Context) error {
	if c.ShouldReject {
		return errors.New("should reject was set to true")
	}
	return nil
}
