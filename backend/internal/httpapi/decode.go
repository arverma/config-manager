package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func decodeJSONBody(w http.ResponseWriter, req *http.Request, dst any, maxBytes int64) error {
	req.Body = http.MaxBytesReader(w, req.Body, maxBytes)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return errors.New("invalid json body")
	}
	// Ensure there are no trailing tokens.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("invalid json body")
	}
	return nil
}
