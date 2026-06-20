package auth

import (
	"encoding/json"
	"net/http"
)

type jsonWriter interface {
	WriteJSON(w http.ResponseWriter, status int, v any)
	WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any)
}

// Handlers exposes HTTP handlers for auth routes.
type Handlers struct {
	svc *Service
	w   jsonWriter
}

func NewHandlers(svc *Service, w jsonWriter) *Handlers {
	return &Handlers{svc: svc, w: w}
}

func (h *Handlers) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	returnTo := r.URL.Query().Get("returnTo")
	url, err := h.svc.GoogleLoginURL(r.Context(), returnTo)
	if err != nil {
		h.w.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to start google login", nil)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handlers) CallbackGoogle(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		h.w.WriteError(w, http.StatusBadRequest, "bad_request", "missing oauth code", nil)
		return
	}

	session, returnTo, err := h.svc.GoogleCallback(r.Context(), code, state)
	if err != nil {
		status := http.StatusUnauthorized
		codeName := "unauthorized"
		if err.Error() == "email domain is not allowed" {
			status = http.StatusForbidden
			codeName = "forbidden"
		}
		h.w.WriteError(w, status, codeName, err.Error(), nil)
		return
	}

	h.svc.SetSessionCookie(w, session.Token)
	http.Redirect(w, r, returnTo, http.StatusFound)
}

func (h *Handlers) Session(w http.ResponseWriter, r *http.Request) {
	if !h.svc.Enabled() {
		h.w.WriteJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"auth_enabled":  false,
		})
		return
	}

	actor, err := h.svc.ResolveRequest(r.Context(), r)
	if err != nil {
		h.w.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated", map[string]any{
			"auth_enabled": true,
		})
		return
	}

	resp := map[string]any{
		"authenticated": true,
		"auth_enabled":  true,
		"actor_type":    actor.Type,
		"actor_id":      actor.ID,
	}
	if actor.Email != "" {
		resp["email"] = actor.Email
	}
	h.w.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.svc.RevokeSession(r.Context(), r)
	h.svc.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Middleware enforces authentication when enabled.
func (h *Handlers) Middleware(basePath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !h.svc.Enabled() {
				next.ServeHTTP(w, r)
				return
			}

			routePath := NormalizeRoutePath(r, basePath)
			if IsPublicPath(routePath) {
				next.ServeHTTP(w, r)
				return
			}

			actor, err := h.svc.ResolveRequest(r.Context(), r)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithActor(r.Context(), actor)))
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"message": message,
	})
}
