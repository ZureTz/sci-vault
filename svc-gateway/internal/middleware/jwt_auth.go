package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	jwt_middleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/gin-gonic/gin"

	"gateway/internal/config"
	"gateway/pkg/jwt"
)

var (
	// The issuer of our token.
	issuer = "sci-vault"

	// The audience of our token.
	audience = []string{"sci-vault-service"}
)

// checkJWT is a gin.HandlerFunc middleware
// that will check the validity of our JWT.
func CheckJWT(cfg *config.JWTConfig) gin.HandlerFunc {
	// Set up the validator.
	jwtValidator, err := validator.New(
		validator.WithKeyFunc(func(ctx context.Context) (any, error) {
			return []byte(cfg.Secret), nil
		}),
		validator.WithAlgorithm(validator.HS256),
		validator.WithIssuer(issuer),
		validator.WithAudiences(audience),
		// WithCustomClaims now uses generics - no need to return interface type
		validator.WithCustomClaims(func() *jwt.CustomClaims {
			return &jwt.CustomClaims{}
		}),
		validator.WithAllowedClockSkew(30*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to set up the validator: %v", err)
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)
	}

	// Set up the jwtMiddleware using pure options pattern
	jwtMiddleware, err := jwt_middleware.New(
		jwt_middleware.WithValidator(jwtValidator),
		jwt_middleware.WithErrorHandler(errorHandler),
	)
	if err != nil {
		log.Fatalf("failed to set up the middleware: %v", err)
	}

	return func(ctx *gin.Context) {
		encounteredError := true
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			encounteredError = false
			ctx.Request = r
			ctx.Next()
		}

		jwtMiddleware.CheckJWT(handler).ServeHTTP(ctx.Writer, ctx.Request)

		if encounteredError {
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				map[string]string{"message": "JWT is invalid."},
			)
		}
	}
}
