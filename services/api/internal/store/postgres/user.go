package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
// There are two unique indexes on users — github_id and lower(email) — so a
// single ON CONFLICT cannot cover both. An address that already has a
// password account must adopt the GitHub identity rather than insert a second
// row, which would fail on users_email_key.
func (r *UserRepo) UpsertGitHub(ctx context.Context, u *domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return translate(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	// 1. A returning user: refresh the profile and the token.
	const refresh = `UPDATE users SET
			name = $2,
			github_login = $3,
			avatar_url = coalesce(nullif($4, ''), avatar_url),
			github_access_token = coalesce(nullif($5, ''), github_access_token),
			github_token_scopes = coalesce(nullif($6, ''), github_token_scopes),
			github_connected_at = now(),
			updated_at = now()
		WHERE github_id = $1
		RETURNING id, email, created_at, updated_at`
	err = tx.QueryRow(ctx, refresh, u.GitHubID, u.Name, u.GitHubLogin, u.AvatarURL,
		u.GitHubAccessToken, u.GitHubTokenScopes).
		Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		return translate(tx.Commit(ctx))
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return translate(err)
	}

	// 2. An account already exists on this address with no GitHub identity —
	//    someone who signed up with a password. Link the two rather than
	//    stranding them with a duplicate-email failure.
	const link = `UPDATE users SET
			github_id = $2,
			github_login = $3,
			avatar_url = coalesce(nullif($4, ''), avatar_url),
			github_access_token = coalesce(nullif($5, ''), github_access_token),
			github_token_scopes = coalesce(nullif($6, ''), github_token_scopes),
			github_connected_at = now(),
			updated_at = now()
		WHERE lower(email) = lower($1) AND github_id IS NULL
		RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, link, u.Email, u.GitHubID, u.GitHubLogin, u.AvatarURL,
		u.GitHubAccessToken, u.GitHubTokenScopes).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		return translate(tx.Commit(ctx))
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return translate(err)
	}

	// 3. The address is taken by a different GitHub account. Linking it would
	//    hand one person another's data, so this stops with a reason the
	//    caller can show.
	var taken bool
	if err := tx.QueryRow(ctx,
		`SELECT true FROM users WHERE lower(email) = lower($1)`, u.Email).Scan(&taken); err == nil {
		return fmt.Errorf("%w: %s is already signed in to OpsPulse with a different GitHub account",
			domain.ErrConflict, u.Email)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return translate(err)
	}

	// 4. Nobody here yet.
	const create = `INSERT INTO users
			(email, name, github_id, github_login, avatar_url,
			 github_access_token, github_token_scopes, github_connected_at)
		VALUES ($1, $2, $3, $4, nullif($5, ''), nullif($6, ''), nullif($7, ''), now())
		RETURNING id, created_at, updated_at`
	if err := tx.QueryRow(ctx, create, u.Email, u.Name, u.GitHubID, u.GitHubLogin,
		u.AvatarURL, u.GitHubAccessToken, u.GitHubTokenScopes).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return translate(err)
	}
	return translate(tx.Commit(ctx))
}
