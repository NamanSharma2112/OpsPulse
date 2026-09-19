package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// RepositoryRepo stores the source repositories a project watches.
type RepositoryRepo struct{ pool *pgxpool.Pool }

const repositoryColumns = `id, project_id, provider, external_id, name,
	default_branch, webhook_secret, created_at, updated_at`

func scanRepository(row interface{ Scan(...any) error }) (*domain.Repository, error) {
	var r domain.Repository
	err := row.Scan(&r.ID, &r.ProjectID, &r.Provider, &r.ExternalID, &r.Name,
		&r.DefaultBranch, &r.WebhookSecret, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &r, nil
}

// Create registers a repository.
func (r *RepositoryRepo) Create(ctx context.Context, repo *domain.Repository) error {
	const q = `INSERT INTO repositories
			(project_id, provider, external_id, name, default_branch, webhook_secret)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, repo.ProjectID, repo.Provider, repo.ExternalID,
		repo.Name, repo.DefaultBranch, repo.WebhookSecret).
		Scan(&repo.ID, &repo.CreatedAt, &repo.UpdatedAt)
	return translate(err)
}

// GetByID loads one repository.
func (r *RepositoryRepo) GetByID(ctx context.Context, id string) (*domain.Repository, error) {
	return scanRepository(r.pool.QueryRow(ctx, `SELECT `+repositoryColumns+` FROM repositories WHERE id = $1`, id))
}

// GetByExternalID resolves the repository an incoming delivery belongs to.
func (r *RepositoryRepo) GetByExternalID(ctx context.Context, provider, externalID string) (*domain.Repository, error) {
	const q = `SELECT ` + repositoryColumns + ` FROM repositories
		WHERE provider = $1 AND external_id = $2`
	return scanRepository(r.pool.QueryRow(ctx, q, provider, externalID))
}

// ListForProject returns the repositories a project watches.
func (r *RepositoryRepo) ListForProject(ctx context.Context, projectID string) ([]domain.Repository, error) {
	const q = `SELECT ` + repositoryColumns + ` FROM repositories
		WHERE project_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, q, projectID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Repository{}
	for rows.Next() {
		repo, err := scanRepository(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *repo)
	}
	return out, translate(rows.Err())
}
