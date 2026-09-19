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
	ExternalID    string    `json:"external_id"`
	Name          string    `json:"name"`
	DefaultBranch string    `json:"default_branch"`
	WebhookSecret string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GitHubExternalID builds the lookup key used for a GitHub repository.
func GitHubExternalID(owner, name string) string {
	return strings.ToLower(strings.TrimSpace(owner)) + "/" + strings.ToLower(strings.TrimSpace(name))
}
