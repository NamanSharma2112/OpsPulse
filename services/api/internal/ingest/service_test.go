package ingest

import (
	"strings"
	"testing"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/github"
)

func TestIncidentSeverity(t *testing.T) {
	tests := []struct {
		name     string
		labels   []string
		want     string
		incident bool
	}{
		{"no labels", nil, "", false},
		{"unrelated labels", []string{"bug", "docs"}, "", false},
		{"plain incident", []string{"incident"}, domain.SeverityMinor, true},
		{"outage", []string{"outage"}, domain.SeverityMinor, true},
		{"severity prefix", []string{"severity:critical"}, domain.SeverityCritical, true},
		{"sev shorthand", []string{"sev2"}, domain.SeverityMajor, true},
		{"mixed", []string{"bug", "incident", "sev1"}, domain.SeverityCritical, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := incidentSeverity(tc.labels)
			if ok != tc.incident {
				t.Fatalf("incidentSeverity(%v) ok = %v, want %v", tc.labels, ok, tc.incident)
			}
			if ok && got != tc.want {
				t.Fatalf("incidentSeverity(%v) = %q, want %q", tc.labels, got, tc.want)
			}
		})
	}
}

func TestMapDeploymentState(t *testing.T) {
	tests := map[string]string{
		"success":     domain.DeploymentSuccess,
		"failure":     domain.DeploymentFailure,
		"error":       domain.DeploymentFailure,
		"in_progress": domain.DeploymentRunning,
		"queued":      domain.DeploymentRunning,
		"inactive":    domain.DeploymentInactive,
		"pending":     domain.DeploymentPending,
		"surprise":    domain.DeploymentPending,
	}
	for state, want := range tests {
		if got := mapDeploymentState(state); got != want {
			t.Errorf("mapDeploymentState(%q) = %q, want %q", state, got, want)
		}
	}
}

func TestOccurredAtPrefersPayloadTimestamps(t *testing.T) {
	want := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	p := &github.Payload{}
	p.PullRequest = &struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		State     string     `json:"state"`
		Draft     bool       `json:"draft"`
		Merged    bool       `json:"merged"`
		HTMLURL   string     `json:"html_url"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		MergedAt  *time.Time `json:"merged_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
	}{UpdatedAt: want}

	if got := occurredAt(p); !got.Equal(want) {
		t.Fatalf("occurredAt() = %v, want %v", got, want)
	}
}

func TestOccurredAtFallsBackToNow(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	got := occurredAt(&github.Payload{})
	if got.Before(before) {
		t.Fatalf("occurredAt() = %v, want a recent timestamp", got)
	}
}

func TestFirstLineIsTheCommitSubject(t *testing.T) {
	tests := map[string]string{
		"Fix the thing":                "Fix the thing",
		"Fix the thing\n\nWith a body": "Fix the thing",
		"  padded  \n body":            "padded",
		"":                             "",
	}
	for message, want := range tests {
		if got := firstLine(message); got != want {
			t.Errorf("firstLine(%q) = %q, want %q", message, got, want)
		}
	}

	// A pathological message must not be stored whole.
	long := strings.Repeat("x", 500)
	if got := firstLine(long); len(got) != 200 {
		t.Errorf("firstLine(500 chars) length = %d, want 200", len(got))
	}
}

func TestBranchIgnoresTags(t *testing.T) {
	tests := map[string]string{
		"refs/heads/main":       "main",
		"refs/heads/feat/thing": "feat/thing",
		"refs/tags/v1.0.0":      "",
		"refs/pull/12/merge":    "",
		"":                      "",
	}
	for ref, want := range tests {
		p := &github.Payload{Ref: ref}
		if got := p.Branch(); got != want {
			t.Errorf("Branch(%q) = %q, want %q", ref, got, want)
		}
	}
}
