package httpapi

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Logger)

	r.Use(corsMiddleware(parseAllowedOriginsEnv()))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "not_ready", "database not reachable", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	// Namespaces
	r.Get("/namespaces", func(w http.ResponseWriter, req *http.Request) {
		handleListNamespaces(w, req, db)
	})
	r.Post("/namespaces", func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Name string `json:"name"`
		}
		if err := decodeJSONBody(w, req, &body, 1<<20); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
			return
		}
		handleCreateNamespace(w, req, db, body.Name)
	})
	r.Delete("/namespaces/{namespace}", func(w http.ResponseWriter, req *http.Request) {
		ns := chi.URLParam(req, "namespace")
		handleDeleteNamespace(w, req, db, ns)
	})
	r.Get("/namespaces/{namespace}/browse", func(w http.ResponseWriter, req *http.Request) {
		ns := chi.URLParam(req, "namespace")
		handleBrowseNamespace(w, req, db, ns)
	})

	// Browse
	r.Get("/configs", func(w http.ResponseWriter, req *http.Request) {
		handleListConfigs(w, req, db)
	})

	// Greedy path routing: /configs/{namespace}/{path...}
	r.Route("/configs/{namespace}", func(r chi.Router) {
		r.Route("/{path:.*}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				handleGetLatestConfig(w, req, db)
			})

			r.Post("/", func(w http.ResponseWriter, req *http.Request) {
				handleCreateConfig(w, req, db)
			})

			r.Put("/", func(w http.ResponseWriter, req *http.Request) {
				handleUpdateConfig(w, req, db)
			})

			r.Delete("/", func(w http.ResponseWriter, req *http.Request) {
				handleDeleteConfig(w, req, db)
			})

			r.Get("/versions", func(w http.ResponseWriter, req *http.Request) {
				handleListConfigVersions(w, req, db)
			})

			r.Get("/versions/{version}", func(w http.ResponseWriter, req *http.Request) {
				if _, err := strconv.Atoi(chi.URLParam(req, "version")); err != nil {
					writeError(w, http.StatusBadRequest, "bad_request", "version must be an integer >= 1", nil)
					return
				}
				handleGetConfigVersion(w, req, db)
			})

			r.Delete("/versions/{version}", func(w http.ResponseWriter, req *http.Request) {
				if _, err := strconv.Atoi(chi.URLParam(req, "version")); err != nil {
					writeError(w, http.StatusBadRequest, "bad_request", "version must be an integer >= 1", nil)
					return
				}
				handleDeleteConfigVersion(w, req, db)
			})
		})
	})

	return r
}

func parseAllowedOriginsEnv() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		return []string{"http://localhost:3000"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"http://localhost:3000"}
	}
	return out
}

func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAny := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAny = true
			continue
		}
		allowed[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")
			if origin != "" {
				if allowAny {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					if _, ok := allowed[origin]; ok {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Add("Vary", "Origin")
					}
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			}

			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
