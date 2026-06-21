package httpapi

import (
	"context"
	"net/http"

	"config-manager/internal/cache"
)

func writeCachedJSON(w http.ResponseWriter, status int, payload []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func invalidateLatestConfigCache(ctx context.Context, cacheSvc *cache.Service, namespace, path string) {
	if cacheSvc == nil || !cacheSvc.Enabled() {
		return
	}
	_ = cacheSvc.InvalidateLatestConfig(ctx, namespace, path)
}
