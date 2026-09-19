package httpapi

import (
	"net/http"
	"strconv"
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
		OrganizationID string `json:"organization_id"`
		Name           string `json:"name"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	project, err := s.projects.Create(r.Context(), currentUser(r).ID, in.OrganizationID, in.Name)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) handleListRepositories(w http.ResponseWriter, r *http.Request) {
	list, err := s.projects.ListRepositories(r.Context(), currentUser(r).ID, r.PathValue("projectID"))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repositories": list})
}

// handleConnectRepository connects a GitHub repository to a project. The
// response carries the webhook secret exactly once, so it can be pasted into
// the repository's webhook settings; it is never served again.
func (s *Server) handleConnectRepository(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Repo          string `json:"repo"`
		DefaultBranch string `json:"default_branch"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	result, err := s.projects.ConnectGitHub(r.Context(), currentUser(r),
		r.PathValue("projectID"), in.Repo, in.DefaultBranch)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// handleListGitHubRepositories lists the repositories the caller can
// administer, which is what the connect flow picks from.
func (s *Server) handleListGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := s.projects.ListGitHubRepositories(r.Context(), currentUser(r))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repositories": repos})
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
