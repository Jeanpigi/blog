package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Jeanpigi/blog/db"
)

// HealthHandler — GET /healthz. Para monitores de uptime / systemd / Docker:
// 200 si la app y la DB responden, 503 si la DB no está disponible.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	w.Header().Set("Cache-Control", "no-store")
	if err := db.Db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "db": "down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
