package httpapi

import (
	"net/http"
	"strconv"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/projects"
)

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	list, err := s.projects.ListForUser(r.Context(), currentUser(r).ID)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": list})
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		OrgID         string `json:"org_id"`
		Name          string `json:"name"`
		Repo          string `json:"repo"`
		DefaultBranch string `json:"default_branch"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	project, secret, err := s.projects.Create(r.Context(), currentUser(r).ID, projects.CreateInput{
		OrgID:         in.OrgID,
		Name:          in.Name,
		Repo:          in.Repo,
		DefaultBranch: in.DefaultBranch,
	})
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	// webhook_secret is shown exactly once, so it can be pasted into the
	// repository's webhook settings.
	writeJSON(w, http.StatusCreated, map[string]any{
		"project":        project,
		"webhook_secret": secret,
	})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	project, err := s.projects.Get(r.Context(), currentUser(r).ID, r.PathValue("projectID"))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// queryLimit reads a caller-supplied page size, falling back to the default.
func queryLimit(r *http.Request, fallback int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
