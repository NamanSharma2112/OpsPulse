// Package store defines the persistence contracts the services depend on.
// The concrete implementation lives in store/postgres.
package store

import (
	"context"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// Users persists accounts.
type Users interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByGitHubID(ctx context.Context, githubID int64) (*domain.User, error)
	UpsertGitHub(ctx context.Context, u *domain.User) error
}

// Orgs persists organisations and their membership.
type Orgs interface {
	Create(ctx context.Context, o *domain.Org) error
	GetByID(ctx context.Context, id string) (*domain.Org, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Org, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Org, error)
	AddMember(ctx context.Context, orgID, userID, role string) error
	RoleOf(ctx context.Context, orgID, userID string) (string, error)
}

// Projects persists watched repositories.
type Projects interface {
	Create(ctx context.Context, p *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByRepo(ctx context.Context, owner, name string) (*domain.Project, error)
	ListForOrg(ctx context.Context, orgID string) ([]domain.Project, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Project, error)
}

// Events persists raw webhook deliveries.
type Events interface {
	// Insert stores an event and reports false when the delivery was already
	// recorded, which makes webhook processing idempotent.
	Insert(ctx context.Context, e *domain.Event) (bool, error)
	ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Event, error)
}

// Metrics persists derived time-series samples.
type Metrics interface {
	Insert(ctx context.Context, m *domain.Metric) error
	ListForProject(ctx context.Context, projectID string, names []string, since time.Time) ([]domain.Metric, error)
	CountSince(ctx context.Context, projectID, name string, since time.Time) (int, error)
}

// Deployments persists the deployment projection.
type Deployments interface {
	Upsert(ctx context.Context, d *domain.Deployment) error
	ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Deployment, error)
}

// PullRequests persists the pull request projection.
type PullRequests interface {
	Upsert(ctx context.Context, pr *domain.PullRequest) error
	ListForProject(ctx context.Context, projectID, state string, limit int) ([]domain.PullRequest, error)
}

// Incidents persists the incident projection.
type Incidents interface {
	Upsert(ctx context.Context, i *domain.Incident) error
	Resolve(ctx context.Context, projectID, externalID string, at time.Time) error
	ListForProject(ctx context.Context, projectID, status string, limit int) ([]domain.Incident, error)
}
