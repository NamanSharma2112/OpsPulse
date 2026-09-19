// Package projects manages projects and the repositories they watch.
package projects

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	gh "github.com/NamanSharma2112/OpsPulse/services/api/internal/github"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/orgs"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

// GitHubClientFactory hands back a GitHub client acting as a given user.
type GitHubClientFactory interface {
	ClientFor(user *domain.User) (*gh.Client, error)
}

// Service exposes project and repository operations.
type Service struct {
	projects     store.Projects
	repositories store.Repositories
	orgs         *orgs.Service
	github       GitHubClientFactory
	// webhookURL is the address GitHub delivers to. It must be reachable
	// from the internet for installation to be worth attempting.
	webhookURL string
	log        *slog.Logger
}

// NewService wires the project service.
func NewService(
	p store.Projects,
	r store.Repositories,
	o *orgs.Service,
	github GitHubClientFactory,
	webhookURL string,
	log *slog.Logger,
) *Service {
	return &Service{projects: p, repositories: r, orgs: o, github: github, webhookURL: webhookURL, log: log}
}

// ConnectResult reports what happened when a repository was connected.
type ConnectResult struct {
	Repository *domain.Repository `json:"repository"`
	// WebhookSecret is returned once. When OpsPulse installed the webhook
	// itself the user never needs it, but it is shown either way so a manual
	// setup stays possible.
	WebhookSecret string `json:"webhook_secret"`
	WebhookURL    string `json:"webhook_url"`
	// WebhookInstalled is false when the hook must be added by hand, with
	// ManualReason saying why.
	WebhookInstalled bool   `json:"webhook_installed"`
	ManualReason     string `json:"manual_reason,omitempty"`
}

// ListGitHubRepositories returns the repositories the caller can administer,
// which is the set OpsPulse can install a webhook on.
func (s *Service) ListGitHubRepositories(ctx context.Context, user *domain.User) ([]gh.Repo, error) {
	client, err := s.github.ClientFor(user)
	if err != nil {
		return nil, err
	}
	return client.ListRepositories(ctx)
}

// Create makes a project inside an organization the caller can write to.
func (s *Service) Create(ctx context.Context, userID, organizationID, name string) (*domain.Project, error) {
	if err := s.requireWriter(ctx, userID, organizationID); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: project name is required", domain.ErrInvalidInput)
	}
	slug := orgs.Slugify(name)
	if slug == "" {
		return nil, fmt.Errorf("%w: project name must contain letters or digits", domain.ErrInvalidInput)
	}

	project := &domain.Project{OrganizationID: organizationID, Name: name, Slug: slug}
	if err := s.projects.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

// ListForUser returns every project visible to the caller.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]domain.Project, error) {
	return s.projects.ListForUser(ctx, userID)
}

// ListForOrganization returns the projects in one organization.
func (s *Service) ListForOrganization(ctx context.Context, userID, organizationID string) ([]domain.Project, error) {
	if _, err := s.orgs.RequireMember(ctx, userID, organizationID); err != nil {
		return nil, err
	}
	return s.projects.ListForOrganization(ctx, organizationID)
}

// Get returns a project the caller may see.
func (s *Service) Get(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if _, err := s.orgs.RequireMember(ctx, userID, project.OrganizationID); err != nil {
		// Hide the existence of projects in other organizations.
		return nil, domain.ErrNotFound
	}
	return project, nil
}

// ConnectGitHub starts watching a GitHub repository, minting the secret its
// webhook must sign deliveries with. The plaintext secret is returned once,
// here, so it can be pasted into the repository's webhook settings; it is
// never served again.
func (s *Service) ConnectGitHub(ctx context.Context, user *domain.User, projectID, repo, defaultBranch string) (*ConnectResult, error) {
	project, err := s.Get(ctx, user.ID, projectID)
	if err != nil {
		return nil, err
	}
	if err := s.requireWriter(ctx, user.ID, project.OrganizationID); err != nil {
		return nil, err
	}

	owner, name, ok := strings.Cut(strings.TrimSpace(repo), "/")
	if !ok || owner == "" || name == "" {
		return nil, fmt.Errorf(`%w: repo must look like "owner/name"`, domain.ErrInvalidInput)
	}
	secret, err := newWebhookSecret()
	if err != nil {
		return nil, err
	}

	repository := &domain.Repository{
		ProjectID:     project.ID,
		Provider:      domain.ProviderGitHub,
		ExternalID:    domain.GitHubExternalID(owner, name),
		Name:          name,
		DefaultBranch: cmpOr(strings.TrimSpace(defaultBranch), "main"),
		WebhookSecret: secret,
		ConnectedBy:   &user.ID,
	}
	if err := s.repositories.Create(ctx, repository); err != nil {
		return nil, err
	}

	result := &ConnectResult{
		Repository:    repository,
		WebhookSecret: secret,
		WebhookURL:    s.webhookURL,
	}

	// Installing the webhook is best effort. The repository is already
	// connected, and a failure here leaves a working manual path rather than
	// losing the connection.
	hookID, reason := s.installWebhook(ctx, user, owner, name, secret)
	if hookID == 0 {
		result.ManualReason = reason
		return result, nil
	}
	if err := s.repositories.RecordWebhook(ctx, repository.ID, hookID); err != nil {
		s.log.Error("webhook installed but not recorded", "error", err, "repository_id", repository.ID)
		result.ManualReason = "webhook installed, but OpsPulse could not record it"
		return result, nil
	}
	repository.WebhookExternalID = &hookID
	result.WebhookInstalled = true
	return result, nil
}

// installWebhook adds the hook to the repository, returning GitHub's id or a
// human-readable reason it could not.
func (s *Service) installWebhook(ctx context.Context, user *domain.User, owner, name, secret string) (int64, string) {
	if s.webhookURL == "" {
		return 0, "no public webhook URL is configured on this server"
	}
	client, err := s.github.ClientFor(user)
	if err != nil {
		return 0, "sign in with GitHub to install webhooks automatically"
	}

	hook, err := client.CreateWebhook(ctx, owner, name, gh.WebhookRequest{
		URL:    s.webhookURL,
		Secret: secret,
		Events: gh.DefaultWebhookEvents,
	})
	if err != nil {
		var apiErr *gh.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.Status {
			case 403:
				return 0, "your GitHub account cannot administer this repository"
			case 404:
				return 0, "repository not found, or your token lacks admin:repo_hook"
			case 422:
				// GitHub rejects a second identical hook.
				return 0, "a webhook with this URL already exists on the repository"
			}
		}
		s.log.Warn("webhook installation failed", "error", err, "repo", owner+"/"+name)
		return 0, "GitHub refused the webhook; add it by hand"
	}
	return hook.ID, ""
}

// ListRepositories returns the repositories a project watches.
func (s *Service) ListRepositories(ctx context.Context, userID, projectID string) ([]domain.Repository, error) {
	project, err := s.Get(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return s.repositories.ListForProject(ctx, project.ID)
}

// RepositoryByExternalID resolves the repository behind an incoming webhook.
// It performs no authorisation: the delivery signature is the credential.
func (s *Service) RepositoryByExternalID(ctx context.Context, provider, externalID string) (*domain.Repository, error) {
	return s.repositories.GetByExternalID(ctx, provider, externalID)
}

// requireWriter allows owners and admins, refusing plain members.
func (s *Service) requireWriter(ctx context.Context, userID, organizationID string) error {
	role, err := s.orgs.RequireMember(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if role == domain.RoleMember {
		return domain.ErrForbidden
	}
	return nil
}

func newWebhookSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func cmpOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
