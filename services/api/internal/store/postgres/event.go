package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// EventRepo stores raw webhook deliveries.
type EventRepo struct{ pool *pgxpool.Pool }

// Insert records a delivery. It returns false when the delivery id was seen
// before, so callers can skip re-projecting a replayed webhook.
func (r *EventRepo) Insert(ctx context.Context, e *domain.Event) (bool, error) {
	const q = `INSERT INTO events (project_id, delivery_id, type, action, actor, payload, occurred_at)
		VALUES ($1, $2, $3, nullif($4, ''), nullif($5, ''), $6, $7)
		ON CONFLICT (delivery_id) DO NOTHING
		RETURNING id, received_at`
	err := r.pool.QueryRow(ctx, q, e.ProjectID, e.DeliveryID, e.Type, e.Action, e.Actor, e.Payload, e.OccurredAt).
		Scan(&e.ID, &e.ReceivedAt)
	if err != nil {
		if translated := translate(err); translated == domain.ErrNotFound {
			return false, nil // conflict: already ingested
		}
		return false, translate(err)
	}
	return true, nil
}

// ListForProject returns the most recent deliveries, newest first.
func (r *EventRepo) ListForProject(ctx context.Context, projectID string, limit int) ([]domain.Event, error) {
	const q = `SELECT id, project_id, delivery_id, type, coalesce(action, ''), coalesce(actor, ''),
			occurred_at, received_at
		FROM events
		WHERE project_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2`
	rows, err := r.pool.Query(ctx, q, projectID, clampLimit(limit, 50, 500))
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	out := []domain.Event{}
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.DeliveryID, &e.Type, &e.Action, &e.Actor,
			&e.OccurredAt, &e.ReceivedAt); err != nil {
			return nil, translate(err)
		}
		out = append(out, e)
	}
	return out, translate(rows.Err())
}

// clampLimit keeps caller-supplied page sizes inside sane bounds.
func clampLimit(limit, fallback, max int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > max {
		return max
	}
	return limit
}
