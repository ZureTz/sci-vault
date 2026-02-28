package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// /^[a-zA-Z0-9_]+$/: matches alphanumeric characters and underscores
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// /^[a-zA-Z0-9_!@#$%^&*]+$/: matches alphanumeric characters, underscores, and special characters
var passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9_!@#$%^&*]+$`)

// Check if the username is valid
func CustomUsernameValidator(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return usernameRegex.MatchString(username)
}

// Check if the password is valid
func CustomPasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return passwordRegex.MatchString(password)
}
