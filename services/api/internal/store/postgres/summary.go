package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// AnalyticsRepo computes the M2 metric set.
type AnalyticsRepo struct{ pool *pgxpool.Pool }

// NewAnalyticsRepo builds an AnalyticsRepo.
func NewAnalyticsRepo(pool *pgxpool.Pool) *AnalyticsRepo { return &AnalyticsRepo{pool: pool} }

// summaryQuery computes every M2 metric in one round trip.
//
// Each CTE is independent and aggregate-only, so each returns exactly one row
// and the final CROSS JOIN cannot drop or multiply anything — including on a
// project with no events at all, where every count comes back zero and every
// percentile comes back NULL.
//
// Hours, not seconds, are returned: the hour is the unit a cycle time is read
// in, and converting here keeps the unit next to the query that defines it.
const summaryQuery = `
WITH cycle AS (
    SELECT
        percentile_cont(0.5) WITHIN GROUP (
            ORDER BY EXTRACT(EPOCH FROM (merged_at - opened_at)) / 3600.0) AS p50,
        percentile_cont(0.9) WITHIN GROUP (
            ORDER BY EXTRACT(EPOCH FROM (merged_at - opened_at)) / 3600.0) AS p90,
        count(*) AS merged
    FROM pull_requests
    WHERE project_id = $1 AND merged_at >= $2
),
decided AS (
    -- A pull request still open has not been decided, so it counts towards
    -- neither side of the merge rate.
    SELECT
        count(*) FILTER (WHERE state = 'merged')              AS merged,
        count(*) FILTER (WHERE state IN ('merged', 'closed')) AS decided
    FROM pull_requests
    WHERE project_id = $1 AND coalesce(merged_at, closed_at) >= $2
),
deploys AS (
    -- Deployment frequency counts what actually reached production, so
    -- pending and running deployments are excluded from both figures.
    SELECT
        count(*) FILTER (WHERE status = 'success')                AS succeeded,
        count(*) FILTER (WHERE status IN ('success', 'failure'))  AS finished
    FROM deployments
    WHERE project_id = $1 AND environment = 'production' AND started_at >= $2
),
open_issues AS (
    -- A level rather than a rate: how many are open now, not how many opened
    -- during the window.
    SELECT count(*) AS n FROM issues WHERE project_id = $1 AND state = 'open'
),
commit_count AS (
    SELECT count(*) AS n FROM commits WHERE project_id = $1 AND committed_at >= $2
)
SELECT cycle.p50, cycle.p90, cycle.merged,
       decided.merged, decided.decided,
       deploys.succeeded, deploys.finished,
       open_issues.n, commit_count.n
FROM cycle, decided, deploys, open_issues, commit_count`

// Summary computes the analytics for one project over one window.
func (r *AnalyticsRepo) Summary(ctx context.Context, projectID string, window time.Duration) (domain.Analytics, error) {
	to := time.Now().UTC()
	from := to.Add(-window)

	a := domain.Analytics{ProjectID: projectID, From: from, To: to}

	var decidedMerged, decidedTotal int
	if err := r.pool.QueryRow(ctx, summaryQuery, projectID, from).Scan(
		&a.PRCycleTimeHours, &a.PRCycleTimeP90Hours, &a.PRsMerged,
		&decidedMerged, &decidedTotal,
		&a.DeploymentsSucceeded, &a.DeploymentsFinished,
		&a.OpenIssues, &a.Commits,
	); err != nil {
		return domain.Analytics{}, translate(err)
	}

	a.PRsDecided = decidedTotal
	// A rate over nothing is not zero, it is unknown.
	if decidedTotal > 0 {
		rate := float64(decidedMerged) / float64(decidedTotal)
		a.PRMergeRate = &rate
	}
	if a.DeploymentsFinished > 0 {
		rate := float64(a.DeploymentsSucceeded) / float64(a.DeploymentsFinished)
		a.DeploymentSuccessRate = &rate
	}

	days := a.Days()
	a.DeploymentsPerDay = float64(a.DeploymentsSucceeded) / days
	a.CommitsPerDay = float64(a.Commits) / days

	return a, nil
}
