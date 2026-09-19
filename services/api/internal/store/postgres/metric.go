package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// MetricRepo stores derived time-series samples.
type MetricRepo struct{ pool *pgxpool.Pool }

// Insert records one sample.
func (r *MetricRepo) Insert(ctx context.Context, m *domain.Metric) error {
	const q = `INSERT INTO metrics (project_id, name, value, labels, recorded_at)
		VALUES ($1, $2, $3, $4, coalesce($5, now()))
		RETURNING id, recorded_at`
	labels := m.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	var at *time.Time
	if !m.RecordedAt.IsZero() {
		at = &m.RecordedAt
	}
	err := r.pool.QueryRow(ctx, q, m.ProjectID, m.Name, m.Value, labels, at).Scan(&m.ID, &m.RecordedAt)
	return translate(err)
}

// ListForProject returns samples for the given metric names since a cutoff.
// An empty name list returns every metric.
func (r *MetricRepo) ListForProject(ctx context.Context, projectID string, names []string, since time.Time) ([]domain.Metric, error) {
	const q = `SELECT id, project_id, name, value, coalesce(labels, '{}'::jsonb), recorded_at
		FROM metrics
		WHERE project_id = $1
		  AND recorded_at >= $2
		  AND (cardinality($3::text[]) = 0 OR name = ANY($3::text[]))
		ORDER BY recorded_at DESC
		LIMIT 1000`
	if names == nil {
		names = []string{}
	}
	rows, err := r.pool.Query(ctx, q, projectID, since, names)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Metric{}
	for rows.Next() {
		var m domain.Metric
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Value, &m.Labels, &m.RecordedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, m)
	}
	return out, translate(rows.Err())
}

// CountSince counts samples of one metric in a window, which is what the
// dashboard health tiles need.
func (r *MetricRepo) CountSince(ctx context.Context, projectID, name string, since time.Time) (int, error) {
	const q = `SELECT count(*) FROM metrics WHERE project_id = $1 AND name = $2 AND recorded_at >= $3`
	var n int
	err := r.pool.QueryRow(ctx, q, projectID, name, since).Scan(&n)
	return n, translate(err)
}
