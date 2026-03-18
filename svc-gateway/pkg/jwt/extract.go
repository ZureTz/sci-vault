package jwt

import (
	"context"
	"errors"
	"fmt"

	jwt_middleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/validator"
)

// GetClaims extracts the validated CustomClaims from the request context.
// It should be called inside handlers protected by the CheckJWT middleware.
func GetClaims(ctx context.Context) (*CustomClaims, error) {
	token, err := jwt_middleware.GetClaims[*validator.ValidatedClaims](ctx)
	if err != nil {
		return nil, fmt.Errorf("token claims not found in context: %w", err)
	}
	customClaims, ok := token.CustomClaims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid custom claims type")
	}
	return customClaims, nil
}
