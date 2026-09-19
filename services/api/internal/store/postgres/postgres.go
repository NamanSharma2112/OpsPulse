// Package postgres implements the store interfaces on top of PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// DB bundles the connection pool with the repositories built on top of it.
type DB struct {
	Pool *pgxpool.Pool

	Users        *UserRepo
	Orgs         *OrgRepo
	Projects     *ProjectRepo
	Events       *EventRepo
	Metrics      *MetricRepo
	Deployments  *DeploymentRepo
	PullRequests *PullRequestRepo
	Incidents    *IncidentRepo
}

// Connect opens a pool and verifies the database answers.
func Connect(ctx context.Context, url string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{
		Pool:         pool,
		Users:        &UserRepo{pool},
		Orgs:         &OrgRepo{pool},
		Projects:     &ProjectRepo{pool},
		Events:       &EventRepo{pool},
		Metrics:      &MetricRepo{pool},
		Deployments:  &DeploymentRepo{pool},
		PullRequests: &PullRequestRepo{pool},
		Incidents:    &IncidentRepo{pool},
	}, nil
}

// Ping reports whether the database is reachable.
func (db *DB) Ping(ctx context.Context) error { return db.Pool.Ping(ctx) }

// Close releases every pooled connection.
func (db *DB) Close() { db.Pool.Close() }

// translate maps driver errors onto domain sentinels so HTTP handlers stay
// free of pgx specifics.
func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}
