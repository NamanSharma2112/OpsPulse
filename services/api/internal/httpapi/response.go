// Package httpapi exposes the service over HTTP.
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// errorBody is the single error shape every endpoint returns.
type errorBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	var body errorBody
	body.Error.Message = message
	writeJSON(w, status, body)
}

// writeDomainError maps a service error onto the right status code, keeping
// internal detail out of the response.
func writeDomainError(w http.ResponseWriter, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "already exists")
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Error("unhandled error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// decodeJSON reads a JSON request body, rejecting unknown fields so typos in
// client payloads surface immediately.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}
