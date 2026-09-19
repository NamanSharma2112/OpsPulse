// Package orgs manages organisations and who belongs to them.
package orgs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

// Service exposes organisation operations.
type Service struct{ orgs store.Orgs }

// NewService wires the organisation service.
func NewService(o store.Orgs) *Service { return &Service{orgs: o} }

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify turns a display name into a URL-safe slug.
func Slugify(s string) string {
	s = nonSlug.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "-")
	return strings.Trim(s, "-")
}

// Create makes an organisation and installs the caller as its owner.
func (s *Service) Create(ctx context.Context, userID, name, githubLogin string) (*domain.Org, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: organisation name is required", domain.ErrInvalidInput)
	}
	slug := Slugify(name)
	if slug == "" {
		return nil, fmt.Errorf("%w: organisation name must contain letters or digits", domain.ErrInvalidInput)
	}

	org := &domain.Org{Name: name, Slug: slug, GitHubLogin: strings.TrimSpace(githubLogin)}
	if err := s.orgs.Create(ctx, org); err != nil {
		return nil, err
	}
	if err := s.orgs.AddMember(ctx, org.ID, userID, domain.RoleOwner); err != nil {
		return nil, err
	}
	return org, nil
}

// ListForUser returns the organisations the caller belongs to.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]domain.Org, error) {
	return s.orgs.ListForUser(ctx, userID)
}

// Get returns an organisation the caller is a member of.
func (s *Service) Get(ctx context.Context, userID, orgID string) (*domain.Org, error) {
	if _, err := s.RequireMember(ctx, userID, orgID); err != nil {
		return nil, err
	}
	return s.orgs.GetByID(ctx, orgID)
}

// RequireMember returns the caller's role, or ErrForbidden when they are not
// a member of the organisation.
func (s *Service) RequireMember(ctx context.Context, userID, orgID string) (string, error) {
	role, err := s.orgs.RoleOf(ctx, orgID, userID)
	if err != nil {
		// A non-member must not be able to tell an organisation apart from a
		// missing one.
		return "", domain.ErrForbidden
	}
	return role, nil
}

// AddMember grants another user access. Only owners and admins may do this.
func (s *Service) AddMember(ctx context.Context, actorID, orgID, userID, role string) error {
	actorRole, err := s.RequireMember(ctx, actorID, orgID)
	if err != nil {
		return err
	}
	if actorRole != domain.RoleOwner && actorRole != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	switch role {
	case domain.RoleOwner, domain.RoleAdmin, domain.RoleMember:
	default:
		return fmt.Errorf("%w: unknown role %q", domain.ErrInvalidInput, role)
	}
	return s.orgs.AddMember(ctx, orgID, userID, role)
}
