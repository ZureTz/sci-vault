package codegen

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"strings"
)

// VerificationCode generates a cryptographically secure random 6-digit code (100000–999999).
func VerificationCode() (string, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// InviteCode generates a cryptographically secure 8-character uppercase alphanumeric code.
func InviteCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate invite code: %w", err)
	}
	// base32 of 5 bytes = exactly 8 characters, no padding needed
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)), nil
}
