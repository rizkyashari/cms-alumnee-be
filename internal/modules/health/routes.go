package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
)

type ReadinessChecker interface {
	PingContext(context.Context) error
}

type response struct {
	Message string `json:"message"`
	Content struct {
		Status string `json:"status"`
	} `json:"content"`
}

func RegisterRoutes(mux interface {
	Get(string, http.HandlerFunc)
}, db ReadinessChecker) {
	mux.Get("/api/v1/health/live", live)
	mux.Get("/api/v1/health/ready", ready(db))
}

func live(w http.ResponseWriter, _ *http.Request) {
	writeHealth(w, http.StatusOK, "success", "ok")
}

func ready(db ReadinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil || db.PingContext(r.Context()) != nil {
			writeHealth(w, http.StatusServiceUnavailable, "service unavailable", "not_ready")
			return
		}
		writeHealth(w, http.StatusOK, "success", "ready")
	}
}

func writeHealth(w http.ResponseWriter, status int, message, state string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	result := response{Message: message}
	result.Content.Status = state
	_ = json.NewEncoder(w).Encode(result)
}

var _ ReadinessChecker = (*sql.DB)(nil)
