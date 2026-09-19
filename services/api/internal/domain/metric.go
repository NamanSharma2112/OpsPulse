package domain

import "time"

// Metric names OpsPulse records while ingesting events.
const (
	MetricDeploymentStarted   = "deployment.started"
	MetricDeploymentSucceeded = "deployment.succeeded"
	MetricDeploymentFailed    = "deployment.failed"
	MetricPullRequestOpened   = "pull_request.opened"
	MetricPullRequestMerged   = "pull_request.merged"
	MetricIncidentOpened      = "incident.opened"
	MetricIncidentResolved    = "incident.resolved"
)

// Metric is a single time-series sample derived from an event.
type Metric struct {
	ID         int64             `json:"id"`
	ProjectID  string            `json:"project_id"`
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels,omitempty"`
	RecordedAt time.Time         `json:"recorded_at"`
}
