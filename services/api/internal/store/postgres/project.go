package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// ProjectRepo stores projects.
type ProjectRepo struct{ pool *pgxpool.Pool }

const projectColumns = `id, organization_id, name, slug, created_at, updated_at`

func scanProject(row interface{ Scan(...any) error }) (*domain.Project, error) {
	var p domain.Project
	err := row.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &p, nil
}

// Create inserts a project.
func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	const q = `INSERT INTO projects (organization_id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, p.OrganizationID, p.Name, p.Slug).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return translate(err)
}

// GetByID loads one project.
func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	return scanProject(r.pool.QueryRow(ctx, `SELECT `+projectColumns+` FROM projects WHERE id = $1`, id))
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

// ListForOrganization returns every project in an organization.
func (r *ProjectRepo) ListForOrganization(ctx context.Context, organizationID string) ([]domain.Project, error) {
	return r.list(ctx, `SELECT `+projectColumns+` FROM projects WHERE organization_id = $1 ORDER BY name`, organizationID)
}

// ListForUser returns every project the user can see through membership.
func (r *ProjectRepo) ListForUser(ctx context.Context, userID string) ([]domain.Project, error) {
	const q = `SELECT p.id, p.organization_id, p.name, p.slug, p.created_at, p.updated_at
		FROM projects p
		JOIN organization_members m ON m.organization_id = p.organization_id
		WHERE m.user_id = $1
		ORDER BY p.name`
	return r.list(ctx, q, userID)
}
