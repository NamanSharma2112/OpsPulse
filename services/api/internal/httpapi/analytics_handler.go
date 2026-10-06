package httpapi

import (
	"net/http"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/health"
)

// handleProjectAnalytics serves the M2 metric set: pull request cycle time,
// deployment frequency, open issues, commit frequency and merge rate.
func (s *Server) handleProjectAnalytics(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}

	name := r.URL.Query().Get("window")
	window, err := domain.ParseWindow(name)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	if name == "" {
		name = domain.DefaultWindow
	}

	summary, err := s.analytics.Summary(r.Context(), project.ID, window)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	summary.Window = name

	writeJSON(w, http.StatusOK, summary)
}

// handleListIssues serves the issue projection.
func (s *Server) handleListIssues(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}

	state := r.URL.Query().Get("state")
	switch state {
	case "", domain.IssueOpen, domain.IssueClosed:
	default:
		writeError(w, http.StatusBadRequest, "state must be open or closed")
		return
	}

	issues, err := s.issues.ListForProject(r.Context(), project.ID, state, queryLimit(r, 25))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"issues": issues})
}

// handleListCommits serves the commit projection.
func (s *Server) handleListCommits(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}

	commits, err := s.commits.ListForProject(r.Context(), project.ID, queryLimit(r, 25))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"commits": commits})
}

// handleProjectScore serves the M3 engineering health score.
//
// Distinct from handleProjectHealth, which is the operational status panel
// (what is on fire right now). This is the composite score over a window:
// one headline number over Reliability, Delivery and Performance.
func (s *Server) handleProjectScore(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}

	name := r.URL.Query().Get("window")
	window, err := domain.ParseWindow(name)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	if name == "" {
		name = domain.DefaultWindow
	}

	summary, err := s.analytics.Summary(r.Context(), project.ID, window)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	summary.Window = name

	writeJSON(w, http.StatusOK, health.Score(health.Inputs{
		Analytics:      summary,
		OpenIncidents:  summary.OpenIncidents,
		MedianResolveH: summary.MedianResolveHours,
	}, s.cfg.HealthWeights()))
}

// handleListAnomalies serves the anomalies detected for a project.
func (s *Server) handleListAnomalies(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}

	list, err := s.anomalies.ListForProject(r.Context(), project.ID, queryLimit(r, 25))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"anomalies": list})
}
