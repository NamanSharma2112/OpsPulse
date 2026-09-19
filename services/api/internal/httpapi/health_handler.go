package httpapi

import (
	"context"
	"net/http"
	"time"
)

// handleLiveness answers as long as the process is running.
func (s *Server) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"env":    s.cfg.Env,
		"time":   time.Now().UTC(),
	})
}

// handleReadiness additionally checks the database, so a load balancer stops
// sending traffic when Postgres is unreachable.
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.pinger.Ping(ctx); err != nil {
		s.log.Warn("readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":   "degraded",
			"database": "unreachable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "database": "ok"})
}
