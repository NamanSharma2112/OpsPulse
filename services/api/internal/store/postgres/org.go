package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// OrgRepo stores organisations and membership.
type OrgRepo struct{ pool *pgxpool.Pool }

const orgColumns = `id, name, slug, coalesce(github_login, ''), created_at, updated_at`

func scanOrg(row interface{ Scan(...any) error }) (*domain.Org, error) {
	var o domain.Org
	if err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.GitHubLogin, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, translate(err)
	}
	return &o, nil
}

// Create inserts an organisation.
func (r *OrgRepo) Create(ctx context.Context, o *domain.Org) error {
	const q = `INSERT INTO orgs (name, slug, github_login)
		VALUES ($1, $2, nullif($3, ''))
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, o.Name, o.Slug, o.GitHubLogin).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	return translate(err)
}

// GetByID loads one organisation.
func (r *OrgRepo) GetByID(ctx context.Context, id string) (*domain.Org, error) {
	return scanOrg(r.pool.QueryRow(ctx, `SELECT `+orgColumns+` FROM orgs WHERE id = $1`, id))
}

// GetBySlug loads an organisation by its URL slug.
func (r *OrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Org, error) {
	return scanOrg(r.pool.QueryRow(ctx, `SELECT `+orgColumns+` FROM orgs WHERE slug = $1`, slug))
}

// ListForUser returns every organisation the user belongs to.
func (r *OrgRepo) ListForUser(ctx context.Context, userID string) ([]domain.Org, error) {
	const q = `SELECT o.id, o.name, o.slug, coalesce(o.github_login, ''), o.created_at, o.updated_at
		FROM orgs o
		JOIN org_members m ON m.org_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.name`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Org{}
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, translate(rows.Err())
}

// AddMember grants a user a role in an organisation.
func (r *OrgRepo) AddMember(ctx context.Context, orgID, userID, role string) error {
	const q = `INSERT INTO org_members (org_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_id, user_id) DO UPDATE SET role = excluded.role`
	_, err := r.pool.Exec(ctx, q, orgID, userID, role)
	return translate(err)
}

// RoleOf returns the caller's role, or domain.ErrNotFound when not a member.
func (r *OrgRepo) RoleOf(ctx context.Context, orgID, userID string) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx, `SELECT role FROM org_members WHERE org_id = $1 AND user_id = $2`, orgID, userID).Scan(&role)
	return role, translate(err)
}
