package domain

import (
	"strings"
	"time"
)

// Providers OpsPulse can read repositories from.
const ProviderGitHub = "github"

// Repository is one source repository a project watches.
type Repository struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Provider  string `json:"provider"`
	// ExternalID is the provider's own identifier. For GitHub it is the
	// lowercased "owner/name", which is what arrives in a webhook payload.
	ExternalID    string `json:"external_id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	WebhookSecret string `json:"-"`
	// WebhookExternalID is GitHub's own hook id, set when OpsPulse installed
	// the webhook. Null means it was configured by hand.
	WebhookExternalID  *int64     `json:"webhook_external_id,omitempty"`
	WebhookInstalledAt *time.Time `json:"webhook_installed_at,omitempty"`
	ConnectedBy        *string    `json:"connected_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Owner and Name split the GitHub external id back into its two halves.
func (r Repository) Owner() string {
	owner, _, _ := strings.Cut(r.ExternalID, "/")
	return owner
}

func (r Repository) RepoName() string {
	_, name, _ := strings.Cut(r.ExternalID, "/")
	return name
}

// WebhookInstalled reports whether OpsPulse installed the webhook itself.
func (r Repository) WebhookInstalled() bool { return r.WebhookExternalID != nil }

// GitHubExternalID builds the lookup key used for a GitHub repository.
func GitHubExternalID(owner, name string) string {
	return strings.ToLower(strings.TrimSpace(owner)) + "/" + strings.ToLower(strings.TrimSpace(name))
}
