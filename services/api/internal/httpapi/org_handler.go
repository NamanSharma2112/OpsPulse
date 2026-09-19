package httpapi

import (
	"net/http"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

func (s *Server) handleListOrgs(w http.ResponseWriter, r *http.Request) {
	list, err := s.orgs.ListForUser(r.Context(), currentUser(r).ID)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"organizations": list})
}

func (s *Server) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		GitHubLogin string `json:"github_login"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	org, err := s.orgs.Create(r.Context(), currentUser(r).ID, in.Name, in.GitHubLogin)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, org)
}

func (s *Server) handleGetOrg(w http.ResponseWriter, r *http.Request) {
	org, err := s.orgs.Get(r.Context(), currentUser(r).ID, r.PathValue("organizationID"))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (s *Server) handleAddOrgMember(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Role == "" {
		in.Role = domain.RoleMember
	}
	err := s.orgs.AddMember(r.Context(), currentUser(r).ID, r.PathValue("organizationID"), in.UserID, in.Role)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) handleListOrgProjects(w http.ResponseWriter, r *http.Request) {
	list, err := s.projects.ListForOrganization(r.Context(), currentUser(r).ID, r.PathValue("organizationID"))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": list})
}
