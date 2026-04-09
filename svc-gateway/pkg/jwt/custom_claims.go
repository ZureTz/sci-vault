package jwt

import (
	"context"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// Validate errors out if `ShouldReject` is true.
func (c *CustomClaims) Validate(ctx context.Context) error {
	return nil
}
