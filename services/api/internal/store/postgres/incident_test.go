package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// Incidents are keyed by the signal that raised them, so one key breaks,
// resolves and breaks again. The reopen rules live in the ON CONFLICT clause,
// which is why these run against a real database.

func openIncident(t *testing.T, db *DB, projectID, externalID string, at time.Time) {
	t.Helper()
	i := domain.Incident{
		ProjectID:  projectID,
		ExternalID: externalID,
		Title:      "CI failed on main",
		Severity:   domain.SeverityMajor,
		Status:     domain.IncidentOpen,
		Source:     "workflow_run",
		OpenedAt:   at,
	}
	if err := db.Incidents.Upsert(context.Background(), &i); err != nil {
		t.Fatalf("open incident: %v", err)
	}
}

func readIncident(t *testing.T, db *DB, projectID, externalID string) (string, time.Time, *time.Time) {
	t.Helper()
	var status string
	var opened time.Time
	var resolved *time.Time
	err := db.Pool.QueryRow(context.Background(),
		`SELECT status, opened_at, resolved_at FROM incidents
		 WHERE project_id = $1 AND external_id = $2`, projectID, externalID).
		Scan(&status, &opened, &resolved)
	if err != nil {
		t.Fatalf("read incident: %v", err)
	}
	return status, opened, resolved
}

// sameInstant compares a timestamp that has been through Postgres, whose
// timestamptz holds microseconds where Go holds nanoseconds.
func sameInstant(a, b time.Time) bool {
	return a.Sub(b).Abs() < time.Microsecond
}

func TestIncidentRepeatWhileOpenKeepsItsStart(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)
	const key = "workflow_run:1:main"

	first := hoursAgo(10)
	openIncident(t, db, projectID, key, first)
	// The workflow fails again before anyone has fixed it. That is the same
	// outage, so time-to-resolve must still be measured from the first break.
	openIncident(t, db, projectID, key, hoursAgo(6))

	status, opened, resolved := readIncident(t, db, projectID, key)
	if status != domain.IncidentOpen {
		t.Fatalf("status = %q, want %q", status, domain.IncidentOpen)
	}
	if !sameInstant(opened, first) {
		t.Fatalf("opened_at = %v, want the first failure at %v", opened, first)
	}
	if resolved != nil {
		t.Fatalf("resolved_at = %v on an open incident, want null", resolved)
	}
}

func TestIncidentReopeningAfterResolutionStartsANewOutage(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	projectID := fixture(t, db)
	const key = "workflow_run:1:main"

	openIncident(t, db, projectID, key, hoursAgo(48))
	if err := db.Incidents.Resolve(ctx, projectID, key, hoursAgo(47)); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// A week later the same workflow breaks again. Carrying the old start over
	// would report a two-day outage, and carrying the old resolution over
	// would leave an open incident claiming it was already fixed.
	again := hoursAgo(2)
	openIncident(t, db, projectID, key, again)

	status, opened, resolved := readIncident(t, db, projectID, key)
	if status != domain.IncidentOpen {
		t.Fatalf("status = %q, want %q", status, domain.IncidentOpen)
	}
	if !sameInstant(opened, again) {
		t.Fatalf("opened_at = %v, want the new failure at %v", opened, again)
	}
	if resolved != nil {
		t.Fatalf("resolved_at = %v on a reopened incident, want null", resolved)
	}
}
