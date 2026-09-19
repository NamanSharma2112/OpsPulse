package httpapi

import (
	"net/http"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// authorizeProject resolves the project in the path and confirms the caller
// may read it, writing the error response itself when they may not.
func (s *Server) authorizeProject(w http.ResponseWriter, r *http.Request) (*domain.Project, bool) {
	project, err := s.projects.Get(r.Context(), currentUser(r).ID, r.PathValue("projectID"))
	if err != nil {
		writeDomainError(w, s.log, err)
		return nil, false
	}
	return project, true
}

// healthResponse is the summary the dashboard's Health panel renders.
type healthResponse struct {
	ProjectID        string             `json:"project_id"`
	Window           string             `json:"window"`
	Status           string             `json:"status"`
	OpenIncidents    int                `json:"open_incidents"`
	OpenPullRequests int                `json:"open_pull_requests"`
	Deployments      healthDeployments  `json:"deployments"`
	LastDeployment   *domain.Deployment `json:"last_deployment,omitempty"`
}

type healthDeployments struct {
	Succeeded   int     `json:"succeeded"`
	Failed      int     `json:"failed"`
	SuccessRate float64 `json:"success_rate"`
}

func (s *Server) handleProjectHealth(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)

	succeeded, err := s.metrics.CountSince(ctx, project.ID, domain.MetricDeploymentSucceeded, since)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	failed, err := s.metrics.CountSince(ctx, project.ID, domain.MetricDeploymentFailed, since)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	openIncidents, err := s.incidents.ListForProject(ctx, project.ID, domain.IncidentOpen, 100)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	openPRs, err := s.pullRequests.ListForProject(ctx, project.ID, "open", 100)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	recent, err := s.deployments.ListForProject(ctx, project.ID, 1)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}

	resp := healthResponse{
		ProjectID:        project.ID,
		Window:           "7d",
		Status:           healthStatus(openIncidents, failed),
		OpenIncidents:    len(openIncidents),
		OpenPullRequests: len(openPRs),
		Deployments: healthDeployments{
			Succeeded:   succeeded,
			Failed:      failed,
			SuccessRate: successRate(succeeded, failed),
		},
	}
	if len(recent) > 0 {
		resp.LastDeployment = &recent[0]
	}
	writeJSON(w, http.StatusOK, resp)
}

// healthStatus grades a project: any critical incident is a hard fail, other
// open incidents or recent failed deploys degrade it.
func healthStatus(incidents []domain.Incident, failedDeploys int) string {
	for _, i := range incidents {
		if i.Severity == domain.SeverityCritical {
			return "critical"
		}
	}
	if len(incidents) > 0 || failedDeploys > 0 {
		return "degraded"
	}
	return "healthy"
}

func successRate(succeeded, failed int) float64 {
	total := succeeded + failed
	if total == 0 {
		return 1
	}
	return float64(succeeded) / float64(total)
}

func (s *Server) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}
	list, err := s.deployments.ListForProject(r.Context(), project.ID, queryLimit(r, 25))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deployments": list})
}

func (s *Server) handleListPullRequests(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}
	state := r.URL.Query().Get("state")
	list, err := s.pullRequests.ListForProject(r.Context(), project.ID, state, queryLimit(r, 25))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"pull_requests": list})
}

func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	list, err := s.incidents.ListForProject(r.Context(), project.ID, status, queryLimit(r, 25))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"incidents": list})
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorizeProject(w, r)
	if !ok {
		return
	}
	list, err := s.events.ListForProject(r.Context(), project.ID, queryLimit(r, 50))
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": list})
}
