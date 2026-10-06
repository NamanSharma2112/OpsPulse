package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// IssueRepo persists the issue projection.
type IssueRepo struct{ pool *pgxpool.Pool }

// NewIssueRepo builds an IssueRepo.
func NewIssueRepo(pool *pgxpool.Pool) *IssueRepo { return &IssueRepo{pool: pool} }

// Upsert applies an issue event to the projection.
//
// Last writer wins, matching the deployment and pull request projections.
// closed_at is coalesced because it is written once and must survive a later
// "reopened" event arriving out of order — except when the issue is open,
// where the schema requires it to be null.
func (r *IssueRepo) Upsert(ctx context.Context, i *domain.Issue) error {
	const q = `
		INSERT INTO issues (project_id, repository_id, number, title, author,
		                    state, url, opened_at, closed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, nullif($7, ''), $8, $9, now())
		ON CONFLICT (project_id, number) DO UPDATE SET
			title      = excluded.title,
			author     = excluded.author,
			state      = excluded.state,
			url        = coalesce(excluded.url, issues.url),
			closed_at  = CASE WHEN excluded.state = 'closed'
			                  THEN coalesce(excluded.closed_at, issues.closed_at)
			                  ELSE NULL END,
			updated_at = now()
		RETURNING id, updated_at`
	return translate(r.pool.QueryRow(ctx, q,
		i.ProjectID, i.RepositoryID, i.Number, i.Title, i.Author,
		i.State, i.URL, i.OpenedAt, i.ClosedAt,
	).Scan(&i.ID, &i.UpdatedAt))
}

// ListForProject returns issues, newest first. An empty state means any.
func (r *IssueRepo) ListForProject(ctx context.Context, projectID, state string, limit int) ([]domain.Issue, error) {
	const q = `
		SELECT id, project_id, repository_id, number, title, author, state,
		       coalesce(url, ''), opened_at, closed_at, updated_at
		FROM issues
		WHERE project_id = $1 AND ($2 = '' OR state = $2)
		ORDER BY opened_at DESC
		LIMIT $3`
	rows, err := r.pool.Query(ctx, q, projectID, state, clampLimit(limit, 25, 200))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Issue{}
	for rows.Next() {
		var i domain.Issue
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.RepositoryID, &i.Number, &i.Title,
			&i.Author, &i.State, &i.URL, &i.OpenedAt, &i.ClosedAt, &i.UpdatedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, i)
	}
	return out, translate(rows.Err())
}

// CommitRepo persists the commit projection.
type CommitRepo struct{ pool *pgxpool.Pool }

// NewCommitRepo builds a CommitRepo.
func NewCommitRepo(pool *pgxpool.Pool) *CommitRepo { return &CommitRepo{pool: pool} }

// InsertBatch stores every commit in one push and reports how many were new.
//
// The same commit arrives again when a branch is merged into the default
// branch, so the conflict is a no-op rather than an error — and the count of
// new rows is what commit frequency is actually measuring.
func (r *CommitRepo) InsertBatch(ctx context.Context, commits []domain.Commit) (int, error) {
	if len(commits) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, translate(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	const q = `
		INSERT INTO commits (project_id, repository_id, sha, message, author, branch, committed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (project_id, sha) DO NOTHING`

	inserted := 0
	for _, c := range commits {
		tag, err := tx.Exec(ctx, q, c.ProjectID, c.RepositoryID, c.SHA, c.Message,
			c.Author, c.Branch, c.CommittedAt)
		if err != nil {
			return 0, translate(err)
		}
		inserted += int(tag.RowsAffected())
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, translate(err)
	}
	return inserted, nil
}

// ListForProject returns commits, newest first.
func (r *CommitRepo) ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Commit, error) {
	const q = `
		SELECT id, project_id, repository_id, sha, message, author, branch, committed_at
		FROM commits
		WHERE project_id = $1
		ORDER BY committed_at DESC
		LIMIT $2`
	rows, err := r.pool.Query(ctx, q, projectID, clampLimit(limit, 25, 200))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Commit{}
	for rows.Next() {
		var c domain.Commit
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.RepositoryID, &c.SHA, &c.Message,
			&c.Author, &c.Branch, &c.CommittedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, c)
	}
	return out, translate(rows.Err())
}
