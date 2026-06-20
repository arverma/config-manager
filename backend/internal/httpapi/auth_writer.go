package httpapi

import "net/http"

type authJSONWriter struct{}

func (authJSONWriter) WriteJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, v)
}

func (authJSONWriter) WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	writeError(w, status, code, message, details)
}

func AuthJSONWriter() authJSONWriter {
	return authJSONWriter{}
}
