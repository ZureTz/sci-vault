package jwt

import (
	"gateway/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// Issuer is the token issuer.
	issuer = "sci-vault"
	// Audience is the token audience.
	audience = []string{"sci-vault-service"}
)

type JWTGenerator struct {
	secret       string
	timeoutHours int
}

func NewJWTGenerator(cfg *config.JWTConfig) *JWTGenerator {
	return &JWTGenerator{
		secret:       cfg.Secret,
		timeoutHours: cfg.Timeout,
	}
}

// GenerateJWT creates a new JWT string given the user ID and username.
func (j *JWTGenerator) GenerateJWT(userID uint, username string) (string, error) {
	claims := struct {
		jwt.RegisteredClaims
		CustomClaims
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  audience,
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(j.timeoutHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		},
		CustomClaims: CustomClaims{
			UserID:       userID,
			Username:     username,
			ShouldReject: false,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}
