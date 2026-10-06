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

// Organizations persists organizations and their membership.
type Organizations interface {
	Create(ctx context.Context, o *domain.Organization) error
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Organization, error)
	AddMember(ctx context.Context, organizationID, userID, role string) error
	RoleOf(ctx context.Context, organizationID, userID string) (string, error)
}

// Projects persists projects.
type Projects interface {
	Create(ctx context.Context, p *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	ListForOrganization(ctx context.Context, organizationID string) ([]domain.Project, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Project, error)
}

// Repositories persists the source repositories a project watches.
type Repositories interface {
	Create(ctx context.Context, r *domain.Repository) error
	GetByID(ctx context.Context, id string) (*domain.Repository, error)
	// GetByExternalID resolves the repository an incoming delivery belongs to.
	GetByExternalID(ctx context.Context, provider, externalID string) (*domain.Repository, error)
	ListForProject(ctx context.Context, projectID string) ([]domain.Repository, error)
	RecordWebhook(ctx context.Context, repositoryID string, hookID int64) error
	Delete(ctx context.Context, repositoryID string) error
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

// Issues persists the issue projection. Every issue lands here, not only the
// ones labelled as incidents.
type Issues interface {
	Upsert(ctx context.Context, i *domain.Issue) error
	ListForProject(ctx context.Context, projectID, state string, limit int) ([]domain.Issue, error)
}

// Commits persists the commit projection, one row per commit in a push.
type Commits interface {
	// InsertBatch stores a push's commits and reports how many were new. The
	// same commit arrives again when a branch merges, so the repeat is a
	// no-op rather than an error.
	InsertBatch(ctx context.Context, commits []domain.Commit) (int, error)
	ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Commit, error)
}

// Analytics computes the derived metric set the dashboard and the health
// score read.
type Analytics interface {
	Summary(ctx context.Context, projectID string, window time.Duration) (domain.Analytics, error)
}

// Anomalies persists detected anomalies.
type Anomalies interface {
	// Insert records an anomaly and reports whether it was new. Detection
	// re-runs over buckets it has already judged, so the repeat is a no-op.
	Insert(ctx context.Context, a *domain.Anomaly) (bool, error)
	ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Anomaly, error)
}
