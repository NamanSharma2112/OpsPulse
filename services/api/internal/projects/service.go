// Package projects manages projects and the repositories they watch.
package projects

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/orgs"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

// Service exposes project and repository operations.
type Service struct {
	projects     store.Projects
	repositories store.Repositories
	orgs         *orgs.Service
}

// NewService wires the project service.
func NewService(p store.Projects, r store.Repositories, o *orgs.Service) *Service {
	return &Service{projects: p, repositories: r, orgs: o}
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
func (s *Service) ConnectGitHub(ctx context.Context, userID, projectID, repo, defaultBranch string) (*domain.Repository, string, error) {
	project, err := s.Get(ctx, userID, projectID)
	if err != nil {
		return nil, "", err
	}
	if err := s.requireWriter(ctx, userID, project.OrganizationID); err != nil {
		return nil, "", err
	}

	owner, name, ok := strings.Cut(strings.TrimSpace(repo), "/")
	if !ok || owner == "" || name == "" {
		return nil, "", fmt.Errorf(`%w: repo must look like "owner/name"`, domain.ErrInvalidInput)
	}
	secret, err := newWebhookSecret()
	if err != nil {
		return nil, "", err
	}

	repository := &domain.Repository{
		ProjectID:     project.ID,
		Provider:      domain.ProviderGitHub,
		ExternalID:    domain.GitHubExternalID(owner, name),
		Name:          name,
		DefaultBranch: cmpOr(strings.TrimSpace(defaultBranch), "main"),
		WebhookSecret: secret,
	}
	if err := s.repositories.Create(ctx, repository); err != nil {
		return nil, "", err
	}
	return repository, secret, nil
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
