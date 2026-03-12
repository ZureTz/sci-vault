package password

import "golang.org/x/crypto/bcrypt"

// Hash generates a bcrypt hash from a plaintext password.
func Hash(plaintext string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	return string(bytes), err
}

// Verify returns true if the plaintext password matches the bcrypt hash.
func Verify(hash, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}
