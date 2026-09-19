package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// ProjectRepo stores watched repositories.
type ProjectRepo struct{ pool *pgxpool.Pool }

const projectColumns = `id, org_id, name, slug, repo_owner, repo_name,
	coalesce(default_branch, 'main'), webhook_secret, created_at, updated_at`

func scanProject(row interface{ Scan(...any) error }) (*domain.Project, error) {
	var p domain.Project
	err := row.Scan(&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.RepoOwner, &p.RepoName,
		&p.DefaultBranch, &p.WebhookSecret, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &p, nil
}

// Create inserts a project.
func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	const q = `INSERT INTO projects (org_id, name, slug, repo_owner, repo_name, default_branch, webhook_secret)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, p.OrgID, p.Name, p.Slug, p.RepoOwner, p.RepoName, p.DefaultBranch, p.WebhookSecret).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return translate(err)
}

// GetByID loads one project.
func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	return scanProject(r.pool.QueryRow(ctx, `SELECT `+projectColumns+` FROM projects WHERE id = $1`, id))
}

// GetByRepo resolves the project a webhook delivery belongs to.
func (r *ProjectRepo) GetByRepo(ctx context.Context, owner, name string) (*domain.Project, error) {
	const q = `SELECT ` + projectColumns + ` FROM projects
		WHERE lower(repo_owner) = lower($1) AND lower(repo_name) = lower($2)`
	return scanProject(r.pool.QueryRow(ctx, q, owner, name))
}

func (r *ProjectRepo) list(ctx context.Context, q string, args ...any) ([]domain.Project, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, translate(rows.Err())
}

// ListForOrg returns every project in an organisation.
func (r *ProjectRepo) ListForOrg(ctx context.Context, orgID string) ([]domain.Project, error) {
	return r.list(ctx, `SELECT `+projectColumns+` FROM projects WHERE org_id = $1 ORDER BY name`, orgID)
}

// ListForUser returns every project the user can see through membership.
func (r *ProjectRepo) ListForUser(ctx context.Context, userID string) ([]domain.Project, error) {
	const q = `SELECT p.id, p.org_id, p.name, p.slug, p.repo_owner, p.repo_name,
			coalesce(p.default_branch, 'main'), p.webhook_secret, p.created_at, p.updated_at
		FROM projects p
		JOIN org_members m ON m.org_id = p.org_id
		WHERE m.user_id = $1
		ORDER BY p.name`
	return r.list(ctx, q, userID)
}
