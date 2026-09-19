package domain

import "time"

// Deployment statuses, mapped from GitHub deployment_status states.
const (
	DeploymentPending  = "pending"
	DeploymentRunning  = "running"
	DeploymentSuccess  = "success"
	DeploymentFailure  = "failure"
	DeploymentInactive = "inactive"
)

// Deployment is a projection of deployment / deployment_status events.
type Deployment struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	ExternalID  string     `json:"external_id"`
	Environment string     `json:"environment"`
	Ref         string     `json:"ref"`
	SHA         string     `json:"sha"`
	Status      string     `json:"status"`
	Actor       string     `json:"actor,omitempty"`
	URL         string     `json:"url,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

// PullRequest is a projection of pull_request events.
type PullRequest struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	State     string     `json:"state"`
	Draft     bool       `json:"draft"`
	URL       string     `json:"url,omitempty"`
	OpenedAt  time.Time  `json:"opened_at"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Incident severities and statuses.
const (
	SeverityCritical = "critical"
	SeverityMajor    = "major"
	SeverityMinor    = "minor"

	IncidentOpen     = "open"
	IncidentAcked    = "acknowledged"
	IncidentResolved = "resolved"
)

// Incident is a projection of failing workflows and issues labelled as
// incidents.
type Incident struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"project_id"`
	ExternalID string     `json:"external_id"`
	Title      string     `json:"title"`
	Severity   string     `json:"severity"`
	Status     string     `json:"status"`
	Source     string     `json:"source"`
	URL        string     `json:"url,omitempty"`
	OpenedAt   time.Time  `json:"opened_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}
