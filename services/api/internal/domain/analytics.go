package domain

import "time"

// Issue states, mapped straight from GitHub.
const (
	IssueOpen   = "open"
	IssueClosed = "closed"
)

// Issue is a projection of issues events. Every issue lands here, not only
// the ones labelled as incidents — "how many issues are open" is a delivery
// signal in its own right.
type Issue struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	RepositoryID *string    `json:"repository_id,omitempty"`
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Author       string     `json:"author"`
	State        string     `json:"state"`
	URL          string     `json:"url,omitempty"`
	OpenedAt     time.Time  `json:"opened_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Commit is a projection of push events, one row per commit.
type Commit struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	RepositoryID *string   `json:"repository_id,omitempty"`
	SHA          string    `json:"sha"`
	Message      string    `json:"message"`
	Author       string    `json:"author"`
	Branch       string    `json:"branch"`
	CommittedAt  time.Time `json:"committed_at"`
}

// Analytics metric names recorded alongside the projections.
const (
	MetricCommitPushed = "commit.pushed"
	MetricIssueOpened  = "issue.opened"
	MetricIssueClosed  = "issue.closed"
)
