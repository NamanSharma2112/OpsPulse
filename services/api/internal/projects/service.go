// Package projects manages the repositories OpsPulse watches.
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

// Service exposes project operations.
type Service struct {
	projects store.Projects
	orgs     *orgs.Service
}

// NewService wires the project service.
func NewService(p store.Projects, o *orgs.Service) *Service {
	return &Service{projects: p, orgs: o}
}

// CreateInput describes a repository to start watching.
type CreateInput struct {
	OrgID         string
	Name          string
	Repo          string // "owner/name"
	DefaultBranch string
}

// Create registers a repository and mints the secret its webhook must sign
// deliveries with.
func (s *Service) Create(ctx context.Context, userID string, in CreateInput) (*domain.Project, string, error) {
	role, err := s.orgs.RequireMember(ctx, userID, in.OrgID)
	if err != nil {
		return nil, "", err
	}
	if role == domain.RoleMember {
		return nil, "", domain.ErrForbidden
	}

	owner, name, ok := strings.Cut(strings.TrimSpace(in.Repo), "/")
	if !ok || owner == "" || name == "" {
		return nil, "", fmt.Errorf(`%w: repo must look like "owner/name"`, domain.ErrInvalidInput)
	}
	display := strings.TrimSpace(in.Name)
	if display == "" {
		display = name
	}
	secret, err := newWebhookSecret()
	if err != nil {
		return nil, "", err
	}

	project := &domain.Project{
		OrgID:         in.OrgID,
		Name:          display,
		Slug:          orgs.Slugify(display),
		RepoOwner:     owner,
		RepoName:      name,
		DefaultBranch: cmpOr(strings.TrimSpace(in.DefaultBranch), "main"),
		WebhookSecret: secret,
	}
	if err := s.projects.Create(ctx, project); err != nil {
		return nil, "", err
	}
	// The plaintext secret is returned once, at creation, so the caller can
	// paste it into GitHub. It is never served again.
	return project, secret, nil
}

// ListForUser returns every project visible to the caller.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]domain.Project, error) {
	return s.projects.ListForUser(ctx, userID)
}

// ListForOrg returns the projects in one organisation.
func (s *Service) ListForOrg(ctx context.Context, userID, orgID string) ([]domain.Project, error) {
	if _, err := s.orgs.RequireMember(ctx, userID, orgID); err != nil {
		return nil, err
	}
	return s.projects.ListForOrg(ctx, orgID)
}

// Get returns a project the caller may see.
func (s *Service) Get(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if _, err := s.orgs.RequireMember(ctx, userID, project.OrgID); err != nil {
		// Hide the existence of projects in other organisations.
		return nil, domain.ErrNotFound
	}
	return project, nil
}

// GetByRepo resolves the project behind an incoming webhook. It performs no
// authorisation: the delivery signature is the credential.
func (s *Service) GetByRepo(ctx context.Context, owner, name string) (*domain.Project, error) {
	return s.projects.GetByRepo(ctx, owner, name)
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
