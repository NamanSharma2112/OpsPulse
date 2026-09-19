package httpapi

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	gh "github.com/NamanSharma2112/OpsPulse/services/api/internal/github"
)

// maxWebhookBody caps a delivery at GitHub's own 25 MB limit.
const maxWebhookBody = 25 << 20

// handleGitHubWebhook ingests one delivery. The request is authenticated by
// its HMAC signature against the project's webhook secret, so no bearer token
// is involved.
func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-GitHub-Event")
	deliveryID := r.Header.Get("X-GitHub-Delivery")
	if eventType == "" || deliveryID == "" {
		writeError(w, http.StatusBadRequest, "missing X-GitHub-Event or X-GitHub-Delivery header")
		return
	}
	if eventType == "ping" {
		writeJSON(w, http.StatusOK, map[string]any{"status": "pong"})
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read request body")
		return
	}

	payload, err := gh.ParsePayload(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	owner, name := repoFromPayload(payload)
	if owner == "" || name == "" {
		writeError(w, http.StatusBadRequest, "payload does not identify a repository")
		return
	}

	project, err := s.projects.GetByRepo(r.Context(), owner, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// The repository is not registered. Say so plainly rather than
			// failing: GitHub should stop retrying.
			writeError(w, http.StatusNotFound, "no project registered for this repository")
			return
		}
		writeDomainError(w, s.log, err)
		return
	}

	if err := gh.VerifySignature(project.WebhookSecret, body, r.Header.Get("X-Hub-Signature-256")); err != nil {
		s.log.Warn("rejected webhook delivery",
			"error", err, "project_id", project.ID, "delivery_id", deliveryID)
		writeError(w, http.StatusUnauthorized, "invalid delivery signature")
		return
	}

	result, err := s.ingest.Handle(r.Context(), project, gh.Delivery{
		ID:    deliveryID,
		Event: eventType,
		Body:  body,
	}, payload)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

// repoFromPayload pulls "owner/name" out of the delivery, preferring the
// explicit owner login and falling back to splitting full_name.
func repoFromPayload(p *gh.Payload) (string, string) {
	owner, name := p.Repository.Owner.Login, p.Repository.Name
	if owner != "" && name != "" {
		return owner, name
	}
	if o, n, ok := strings.Cut(p.Repository.FullName, "/"); ok {
		return o, n
	}
	return "", ""
}
