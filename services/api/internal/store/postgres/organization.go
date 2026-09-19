package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// OrganizationRepo stores organizations and membership.
type OrganizationRepo struct{ pool *pgxpool.Pool }

const organizationColumns = `id, name, slug, coalesce(github_login, ''), created_at, updated_at`

func scanOrganization(row interface{ Scan(...any) error }) (*domain.Organization, error) {
	var o domain.Organization
	if err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.GitHubLogin, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, translate(err)
	}
	return &o, nil
}

// Create inserts an organization.
func (r *OrganizationRepo) Create(ctx context.Context, o *domain.Organization) error {
	const q = `INSERT INTO organizations (name, slug, github_login)
		VALUES ($1, $2, nullif($3, ''))
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, o.Name, o.Slug, o.GitHubLogin).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	return translate(err)
}

// GetByID loads one organization.
func (r *OrganizationRepo) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	return scanOrganization(r.pool.QueryRow(ctx, `SELECT `+organizationColumns+` FROM organizations WHERE id = $1`, id))
}

// GetBySlug loads an organization by its URL slug.
func (r *OrganizationRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	return scanOrganization(r.pool.QueryRow(ctx, `SELECT `+organizationColumns+` FROM organizations WHERE slug = $1`, slug))
}

// ListForUser returns every organization the user belongs to.
func (r *OrganizationRepo) ListForUser(ctx context.Context, userID string) ([]domain.Organization, error) {
	const q = `SELECT o.id, o.name, o.slug, coalesce(o.github_login, ''), o.created_at, o.updated_at
		FROM organizations o
		JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.name`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Organization{}
	for rows.Next() {
		o, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, translate(rows.Err())
}

// AddMember grants a user a role in an organization.
func (r *OrganizationRepo) AddMember(ctx context.Context, organizationID, userID, role string) error {
	const q = `INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = excluded.role`
	_, err := r.pool.Exec(ctx, q, organizationID, userID, role)
	return translate(err)
}

// RoleOf returns the caller's role, or domain.ErrNotFound when not a member.
func (r *OrganizationRepo) RoleOf(ctx context.Context, organizationID, userID string) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx, `SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`, organizationID, userID).Scan(&role)
	return role, translate(err)
}
