package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// AnomalyRepo persists detected anomalies.
type AnomalyRepo struct{ pool *pgxpool.Pool }

// NewAnomalyRepo builds an AnomalyRepo.
func NewAnomalyRepo(pool *pgxpool.Pool) *AnomalyRepo { return &AnomalyRepo{pool: pool} }

// Insert records an anomaly and reports whether it was new.
//
// Detection re-runs on a timer over a window that overlaps what it has
// already judged, so the same bucket is evaluated repeatedly. The unique key
// is what stops one spike raising an incident every few minutes.
func (r *AnomalyRepo) Insert(ctx context.Context, a *domain.Anomaly) (bool, error) {
	const q = `
		INSERT INTO anomalies (project_id, metric, bucket_start, observed,
		                       baseline_median, baseline_spread, deviation,
		                       direction, explanation, incident_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (project_id, metric, bucket_start) DO NOTHING
		RETURNING id, detected_at`
	err := r.pool.QueryRow(ctx, q, a.ProjectID, a.Metric, a.BucketStart, a.Observed,
		a.BaselineMedian, a.BaselineSpread, a.Deviation, a.Direction,
		a.Explanation, a.IncidentID).Scan(&a.ID, &a.DetectedAt)
	if err != nil {
		// DO NOTHING returns no row, which is the "already known" case rather
		// than a failure.
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, translate(err)
	}
	return true, nil
}

// ListForProject returns anomalies, most recently detected first.
func (r *AnomalyRepo) ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Anomaly, error) {
	const q = `
		SELECT id, project_id, metric, bucket_start, observed, baseline_median,
		       baseline_spread, deviation, direction, explanation, incident_id, detected_at
		FROM anomalies
		WHERE project_id = $1
		ORDER BY detected_at DESC
		LIMIT $2`
	rows, err := r.pool.Query(ctx, q, projectID, clampLimit(limit, 25, 200))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Anomaly{}
	for rows.Next() {
		var a domain.Anomaly
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.Metric, &a.BucketStart, &a.Observed,
			&a.BaselineMedian, &a.BaselineSpread, &a.Deviation, &a.Direction,
			&a.Explanation, &a.IncidentID, &a.DetectedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, a)
	}
	return out, translate(rows.Err())
}

// DailySeries buckets a metric into whole UTC days.
//
// generate_series supplies every day in the window, so a day on which nothing
// happened comes back as zero rather than as a gap. That matters: a baseline
// built only from days that had activity would treat "no deploys at all" as
// unremarkable, which is exactly the collapse worth detecting.
func (r *MetricRepo) DailySeries(ctx context.Context, projectID, name string, from, to time.Time) ([]domain.Bucket, error) {
	const q = `
		SELECT d.day, coalesce(sum(m.value), 0)
		FROM generate_series(date_trunc('day', $3::timestamptz),
		                     date_trunc('day', $4::timestamptz),
		                     interval '1 day') AS d(day)
		LEFT JOIN metrics m
		       ON m.project_id = $1
		      AND m.name = $2
		      AND date_trunc('day', m.recorded_at) = d.day
		GROUP BY d.day
		ORDER BY d.day`
	rows, err := r.pool.Query(ctx, q, projectID, name, from, to)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Bucket{}
	for rows.Next() {
		var b domain.Bucket
		if err := rows.Scan(&b.Start, &b.Value); err != nil {
			return nil, translate(err)
		}
		out = append(out, b)
	}
	return out, translate(rows.Err())
}
