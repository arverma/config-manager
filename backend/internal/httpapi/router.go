package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// Minimal CORS for local UI dev (private-network assumptions).
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
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
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid json body", nil)
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
