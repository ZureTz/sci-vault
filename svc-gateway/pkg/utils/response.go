package utils

func ErrorResponse(err error) map[string]any {
	return map[string]any{"error": err.Error()}
}

func MessageResponse(message string) map[string]any {
	return map[string]any{"message": message}
}
