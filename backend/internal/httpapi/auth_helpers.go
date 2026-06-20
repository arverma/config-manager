package httpapi

import (
	"net/http"

	"config-manager/internal/auth"
)

func resolveCreatedBy(req *http.Request, clientProvided *string, authEnabled bool) (*string, error) {
	if clientProvided != nil && authEnabled {
		return nil, errClientCreatedByNotAllowed
	}
	if !authEnabled {
		return clientProvided, nil
	}
	actor, ok := auth.ActorFromContext(req.Context())
	if !ok {
		return nil, errUnauthenticated
	}
	label := actor.CreatedByLabel()
	return &label, nil
}

var (
	errClientCreatedByNotAllowed = &createdByError{code: "client_created_by_not_allowed", message: "created_by must not be supplied when authentication is enabled"}
	errUnauthenticated           = &createdByError{code: "unauthorized", message: "authentication required"}
)

type createdByError struct {
	code    string
	message string
}

func (e *createdByError) Error() string {
	return e.message
}

func writeCreatedByError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if ce, ok := err.(*createdByError); ok {
		status := http.StatusBadRequest
		if ce.code == "unauthorized" {
			status = http.StatusUnauthorized
		}
		writeError(w, status, ce.code, ce.message, nil)
		return true
	}
	writeError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
	return true
}

func authEnabled() bool {
	svc := globalAuthService
	return svc != nil && svc.Enabled()
}

var globalAuthService *auth.Service

func SetAuthService(svc *auth.Service) {
	globalAuthService = svc
}

func AuthService() *auth.Service {
	return globalAuthService
}
