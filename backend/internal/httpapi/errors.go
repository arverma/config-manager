package httpapi

import "net/http"

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	writeJSON(w, status, apiError{
		Code:    code,
		Message: message,
		Details: details,
	})
}
