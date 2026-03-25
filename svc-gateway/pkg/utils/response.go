package utils

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// fieldToKey converts a Go struct field name to its i18n locale key segment.
func fieldToKey(field string) string {
	switch field {
	case "Username":
		return "username"
	case "Password":
		return "password"
	case "Email":
		return "email"
	case "EmailCode":
		return "email_code"
	case "ConfirmedPassword":
		return "confirmed_password"
	case "Avatar":
		return "avatar"
	case "Nickname":
		return "nickname"
	case "Bio":
		return "bio"
	case "Website":
		return "website"
	case "Location":
		return "location"
	case "UserID":
		return "user_id"
	case "File":
		return "file"
	case "Title":
		return "title"
	case "Year":
		return "year"
	case "DOI":
		return "doi"
	case "DocID":
		return "doc_id"
	default:
		return field
	}
}

// formatValidationError returns a dot-notation i18n locale key for the frontend to translate,
// e.g. "validation.username.required", "validation.email.invalid".
func formatValidationError(f validator.FieldError) string {
	field := fieldToKey(f.Field())
	switch f.Tag() {
	case "required_without":
		// Username and Email are mutually required_without each other → unified identifier message
		return "validation.identifier.required"
	case "email", "http_url",
		"len", "numeric",
		"custom_username_validator", "custom_password_validator":
		return fmt.Sprintf("validation.%s.invalid", field)
	case "eqfield":
		return fmt.Sprintf("validation.%s.mismatch", field)
	}
	// covers: required, min, max, and any other standard tags
	return fmt.Sprintf("validation.%s.%s", field, f.Tag())
}

func ErrorResponse(err error) map[string]any {
	if validationErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		errMsgs := make([]string, 0, len(validationErrs))
		for _, f := range validationErrs {
			errMsgs = append(errMsgs, formatValidationError(f))
		}
		return map[string]any{"errors": errMsgs}
	}
	return map[string]any{"errors": []string{err.Error()}}
}

func MessageResponse(message string) map[string]any {
	return map[string]any{"message": message}
}
