package domain

import (
	"fmt"
	"time"
)

// Analytics is the M2 metric set for one project over one window.
//
// The optional fields are pointers rather than zero values because "no merged
// pull requests in this window" and "a cycle time of zero hours" are different
// answers, and the health score in M3 has to be able to tell them apart: an
// unknown input is skipped, a genuinely bad one is scored.
type Analytics struct {
	ProjectID string    `json:"project_id"`
	Window    string    `json:"window"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`

	// Median and 90th percentile hours from opening a pull request to
	// merging it. Median, not mean: one pull request left open over a holiday
	// would otherwise move the headline number for a month.
	PRCycleTimeHours    *float64 `json:"pr_cycle_time_hours"`
	PRCycleTimeP90Hours *float64 `json:"pr_cycle_time_p90_hours"`
	PRsMerged           int      `json:"prs_merged"`

	// Merged as a share of pull requests that reached a decision in the
	// window. Pull requests still open are not counted either way — they have
	// not been decided yet.
	PRMergeRate *float64 `json:"pr_merge_rate"`
	PRsDecided  int      `json:"prs_decided"`

	// Successful production deployments per day, and the share of finished
	// deployments that succeeded.
	DeploymentsPerDay     float64  `json:"deployments_per_day"`
	DeploymentsSucceeded  int      `json:"deployments_succeeded"`
	DeploymentsFinished   int      `json:"deployments_finished"`
	DeploymentSuccessRate *float64 `json:"deployment_success_rate"`

	// Issues open right now. A level, not a rate, so the window does not
	// apply to it.
	OpenIssues int `json:"open_issues"`

	// Incidents open right now, and how long the ones resolved in the window
	// took. Both feed the reliability pillar of the health score.
	OpenIncidents      int      `json:"open_incidents"`
	MedianResolveHours *float64 `json:"median_resolve_hours"`
	IncidentsResolved  int      `json:"incidents_resolved"`

	CommitsPerDay float64 `json:"commits_per_day"`
	Commits       int     `json:"commits"`
}

// Days is the window length in days, never zero, so it is always safe to
// divide by.
func (a Analytics) Days() float64 {
	d := a.To.Sub(a.From).Hours() / 24
	if d < 1 {
		return 1
	}
	return d
}

// Windows the API accepts. Anything else is rejected rather than silently
// coerced, so a typo cannot quietly change what a number means.
var windows = map[string]time.Duration{
	"7d":  7 * 24 * time.Hour,
	"14d": 14 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
	"90d": 90 * 24 * time.Hour,
}

// DefaultWindow is used when a request does not name one.
const DefaultWindow = "30d"

// ParseWindow resolves a window name to its duration.
func ParseWindow(name string) (time.Duration, error) {
	if name == "" {
		name = DefaultWindow
	}
	d, ok := windows[name]
	if !ok {
		return 0, fmt.Errorf("%w: unknown window %q (use 7d, 14d, 30d or 90d)", ErrInvalidInput, name)
	}
	return d, nil
}
