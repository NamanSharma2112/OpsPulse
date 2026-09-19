package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// DeploymentRepo stores the deployment projection.
type DeploymentRepo struct{ pool *pgxpool.Pool }

// Upsert writes the current state of a deployment, keyed by its GitHub id.
func (r *DeploymentRepo) Upsert(ctx context.Context, d *domain.Deployment) error {
	const q = `INSERT INTO deployments
			(project_id, repository_id, external_id, environment, ref, commit_sha, status,
			 actor, url, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, nullif($8, ''), nullif($9, ''), $10, $11)
		ON CONFLICT (project_id, external_id) DO UPDATE
			SET status = excluded.status,
			    environment = excluded.environment,
			    actor = coalesce(excluded.actor, deployments.actor),
			    url = coalesce(excluded.url, deployments.url),
			    finished_at = coalesce(excluded.finished_at, deployments.finished_at)
		RETURNING id`
	err := r.pool.QueryRow(ctx, q, d.ProjectID, d.RepositoryID, d.ExternalID, d.Environment,
		d.Ref, d.CommitSHA, d.Status, d.Actor, d.URL, d.StartedAt, d.FinishedAt).Scan(&d.ID)
	return translate(err)
}

// ListForProject returns recent deployments, newest first.
func (r *DeploymentRepo) ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Deployment, error) {
	const q = `SELECT id, project_id, repository_id, external_id, environment, ref, commit_sha,
			status, coalesce(actor, ''), coalesce(url, ''), started_at, finished_at
		FROM deployments WHERE project_id = $1 ORDER BY started_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, projectID, clampLimit(limit, 25, 200))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Deployment{}
	for rows.Next() {
		var d domain.Deployment
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.RepositoryID, &d.ExternalID, &d.Environment,
			&d.Ref, &d.CommitSHA, &d.Status, &d.Actor, &d.URL, &d.StartedAt, &d.FinishedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, d)
	}
	return out, translate(rows.Err())
}

// PullRequestRepo stores the pull request projection.
type PullRequestRepo struct{ pool *pgxpool.Pool }

// Upsert writes the current state of a pull request, keyed by its number.
func (r *PullRequestRepo) Upsert(ctx context.Context, pr *domain.PullRequest) error {
	const q = `INSERT INTO pull_requests
			(project_id, repository_id, number, title, author, state, draft, url,
			 opened_at, merged_at, closed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, nullif($8, ''), $9, $10, $11, now())
		ON CONFLICT (project_id, number) DO UPDATE
			SET title = excluded.title,
			    state = excluded.state,
			    draft = excluded.draft,
			    merged_at = coalesce(excluded.merged_at, pull_requests.merged_at),
			    closed_at = coalesce(excluded.closed_at, pull_requests.closed_at),
			    updated_at = now()
		RETURNING id, updated_at`
	err := r.pool.QueryRow(ctx, q, pr.ProjectID, pr.RepositoryID, pr.Number, pr.Title, pr.Author,
		pr.State, pr.Draft, pr.URL, pr.OpenedAt, pr.MergedAt, pr.ClosedAt).Scan(&pr.ID, &pr.UpdatedAt)
	return translate(err)
}

// ListForProject returns pull requests, optionally filtered by state.
func (r *PullRequestRepo) ListForProject(ctx context.Context, projectID, state string, limit int) ([]domain.PullRequest, error) {
	const q = `SELECT id, project_id, repository_id, number, title, author, state, draft,
			coalesce(url, ''), opened_at, merged_at, closed_at, updated_at
		FROM pull_requests
		WHERE project_id = $1 AND ($2 = '' OR state = $2)
		ORDER BY updated_at DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, q, projectID, state, clampLimit(limit, 25, 200))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.PullRequest{}
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.ProjectID, &pr.RepositoryID, &pr.Number, &pr.Title,
			&pr.Author, &pr.State, &pr.Draft, &pr.URL, &pr.OpenedAt, &pr.MergedAt,
			&pr.ClosedAt, &pr.UpdatedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, pr)
	}
	return out, translate(rows.Err())
}

// IncidentRepo stores the incident projection.
type IncidentRepo struct{ pool *pgxpool.Pool }

// Upsert opens or updates an incident, keyed by its source identifier.
func (r *IncidentRepo) Upsert(ctx context.Context, i *domain.Incident) error {
	const q = `INSERT INTO incidents
			(project_id, external_id, title, severity, status, source, url, opened_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, nullif($7, ''), $8, $9)
		ON CONFLICT (project_id, external_id) DO UPDATE
			SET title = excluded.title,
			    severity = excluded.severity,
			    status = excluded.status,
			    resolved_at = coalesce(excluded.resolved_at, incidents.resolved_at)
		RETURNING id`
	err := r.pool.QueryRow(ctx, q, i.ProjectID, i.ExternalID, i.Title, i.Severity, i.Status,
		i.Source, i.URL, i.OpenedAt, i.ResolvedAt).Scan(&i.ID)
	return translate(err)
}

// Resolve closes an open incident.
func (r *IncidentRepo) Resolve(ctx context.Context, projectID, externalID string, at time.Time) error {
	const q = `UPDATE incidents SET status = $3, resolved_at = $4
		WHERE project_id = $1 AND external_id = $2 AND status <> $3`
	_, err := r.pool.Exec(ctx, q, projectID, externalID, domain.IncidentResolved, at)
	return translate(err)
}

// ListForProject returns incidents, optionally filtered by status.
func (r *IncidentRepo) ListForProject(ctx context.Context, projectID, status string, limit int) ([]domain.Incident, error) {
	const q = `SELECT id, project_id, external_id, title, severity, status, source,
			coalesce(url, ''), opened_at, resolved_at
		FROM incidents
		WHERE project_id = $1 AND ($2 = '' OR status = $2)
		ORDER BY opened_at DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, q, projectID, status, clampLimit(limit, 25, 200))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Incident{}
	for rows.Next() {
		var i domain.Incident
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.ExternalID, &i.Title, &i.Severity, &i.Status,
			&i.Source, &i.URL, &i.OpenedAt, &i.ResolvedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, i)
	}
	return out, translate(rows.Err())
}
