package domain

import "time"

// Project is one GitHub repository OpsPulse watches.
type Project struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	RepoOwner     string    `json:"repo_owner"`
	RepoName      string    `json:"repo_name"`
	DefaultBranch string    `json:"default_branch"`
	WebhookSecret string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RepoFullName returns the "owner/name" form GitHub uses in payloads.
func (p Project) RepoFullName() string { return p.RepoOwner + "/" + p.RepoName }
