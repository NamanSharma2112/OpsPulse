package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// UserRepo stores accounts.
type UserRepo struct{ pool *pgxpool.Pool }

const userColumns = `id, email, name, coalesce(password_hash, ''), github_id,
	coalesce(github_login, ''), coalesce(avatar_url, ''),
	coalesce(github_access_token, ''), coalesce(github_token_scopes, ''),
	github_connected_at, created_at, updated_at`

func (r *UserRepo) scan(row interface{ Scan(...any) error }) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.GitHubID,
		&u.GitHubLogin, &u.AvatarURL, &u.GitHubAccessToken, &u.GitHubTokenScopes,
		&u.GitHubConnectedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

// Create inserts a new account.
func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `INSERT INTO users (email, name, password_hash, github_id, github_login, avatar_url)
		VALUES ($1, $2, nullif($3, ''), $4, nullif($5, ''), nullif($6, ''))
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, u.Email, u.Name, u.PasswordHash, u.GitHubID, u.GitHubLogin, u.AvatarURL).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	return translate(err)
}

// GetByID loads one account.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scan(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

// GetByEmail loads an account by its login address.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scan(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1)`, email))
}

// GetByGitHubID loads an account linked to a GitHub user.
func (r *UserRepo) GetByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
	return r.scan(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE github_id = $1`, githubID))
}

// UpsertGitHub creates or refreshes the account behind a GitHub identity,
// storing the sealed access token so OpsPulse can act on the user's behalf.
//
// The email is only written on insert: a user who set a different address in
// OpsPulse should not have it overwritten every time they sign in.
func (r *UserRepo) UpsertGitHub(ctx context.Context, u *domain.User) error {
	const q = `INSERT INTO users
			(email, name, github_id, github_login, avatar_url,
			 github_access_token, github_token_scopes, github_connected_at)
		VALUES ($1, $2, $3, $4, nullif($5, ''), nullif($6, ''), nullif($7, ''), now())
		-- users_github_id_key is a partial index, so the conflict target has
		-- to repeat its predicate for Postgres to infer it.
		ON CONFLICT (github_id) WHERE github_id IS NOT NULL DO UPDATE
			SET name = excluded.name,
			    github_login = excluded.github_login,
			    avatar_url = excluded.avatar_url,
			    github_access_token = coalesce(excluded.github_access_token, users.github_access_token),
			    github_token_scopes = coalesce(excluded.github_token_scopes, users.github_token_scopes),
			    github_connected_at = now(),
			    updated_at = now()
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, u.Email, u.Name, u.GitHubID, u.GitHubLogin, u.AvatarURL,
		u.GitHubAccessToken, u.GitHubTokenScopes).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	return translate(err)
}
