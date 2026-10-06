package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

/*
These run against a real database because the behaviour under test is the SQL
itself — percentile_cont, the FILTER clauses and the CROSS JOIN of five CTEs.
A fake store would only prove that Go can divide.
*/

// fixture builds an org, a user and a project to hang projections off, and
// clears everything first so each test starts from a known empty database.
func fixture(t *testing.T, db *DB) string {
	t.Helper()
	ctx := context.Background()

	// projects cascade to every projection, so this is enough to reset.
	for _, table := range []string{"projects", "organizations", "users"} {
		if _, err := db.Pool.Exec(ctx, "DELETE FROM "+table); err != nil {
			t.Fatalf("clean %s: %v", table, err)
		}
	}

	var orgID string
	if err := db.Pool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Acme', 'acme') RETURNING id`).Scan(&orgID); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	var projectID string
	if err := db.Pool.QueryRow(ctx, `
		INSERT INTO projects (organization_id, name, slug)
		VALUES ($1, 'Storefront', 'storefront')
		RETURNING id`, orgID).Scan(&projectID); err != nil {
		t.Fatalf("create project: %v", err)
	}
	return projectID
}

func hoursAgo(h float64) time.Time {
	return time.Now().UTC().Add(-time.Duration(h * float64(time.Hour)))
}

// mergedPR inserts a pull request that opened `openedHoursAgo` ago and merged
// `mergedHoursAgo` ago, so its cycle time is the difference.
func mergedPR(t *testing.T, db *DB, projectID string, number int, openedHoursAgo, mergedHoursAgo float64) {
	t.Helper()
	opened, merged := hoursAgo(openedHoursAgo), hoursAgo(mergedHoursAgo)
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO pull_requests (project_id, number, title, state, opened_at, merged_at, closed_at)
		VALUES ($1, $2, 'pr', 'merged', $3, $4, $4)`,
		projectID, number, opened, merged); err != nil {
		t.Fatalf("insert merged pr %d: %v", number, err)
	}
}

func closedPR(t *testing.T, db *DB, projectID string, number int, closedHoursAgo float64) {
	t.Helper()
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO pull_requests (project_id, number, title, state, opened_at, closed_at)
		VALUES ($1, $2, 'pr', 'closed', $3, $4)`,
		projectID, number, hoursAgo(closedHoursAgo+10), hoursAgo(closedHoursAgo)); err != nil {
		t.Fatalf("insert closed pr %d: %v", number, err)
	}
}

func openPR(t *testing.T, db *DB, projectID string, number int) {
	t.Helper()
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO pull_requests (project_id, number, title, state, opened_at)
		VALUES ($1, $2, 'pr', 'open', $3)`, projectID, number, hoursAgo(5)); err != nil {
		t.Fatalf("insert open pr %d: %v", number, err)
	}
}

func deployment(t *testing.T, db *DB, projectID, externalID, status, env string, hoursAgoStarted float64) {
	t.Helper()
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO deployments (project_id, external_id, environment, status, started_at)
		VALUES ($1, $2, $3, $4, $5)`,
		projectID, externalID, env, status, hoursAgo(hoursAgoStarted)); err != nil {
		t.Fatalf("insert deployment %s: %v", externalID, err)
	}
}

func issue(t *testing.T, db *DB, projectID string, number int, state string) {
	t.Helper()
	var closedAt *time.Time
	if state == domain.IssueClosed {
		at := hoursAgo(1)
		closedAt = &at
	}
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO issues (project_id, number, title, state, opened_at, closed_at)
		VALUES ($1, $2, 'issue', $3, $4, $5)`,
		projectID, number, state, hoursAgo(48), closedAt); err != nil {
		t.Fatalf("insert issue %d: %v", number, err)
	}
}

func commit(t *testing.T, db *DB, projectID, sha string, hoursAgoCommitted float64) {
	t.Helper()
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO commits (project_id, sha, committed_at) VALUES ($1, $2, $3)`,
		projectID, sha, hoursAgo(hoursAgoCommitted)); err != nil {
		t.Fatalf("insert commit %s: %v", sha, err)
	}
}

func TestSummaryOnEmptyProject(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	got, err := db.Analytics.Summary(context.Background(), projectID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	// Nothing has happened, so the rates are unknown rather than zero. This
	// is the distinction the health score depends on.
	if got.PRCycleTimeHours != nil {
		t.Errorf("cycle time = %v, want nil on an empty project", *got.PRCycleTimeHours)
	}
	if got.PRMergeRate != nil {
		t.Errorf("merge rate = %v, want nil on an empty project", *got.PRMergeRate)
	}
	if got.DeploymentSuccessRate != nil {
		t.Errorf("success rate = %v, want nil on an empty project", *got.DeploymentSuccessRate)
	}
	// Counts, by contrast, are genuinely zero.
	if got.OpenIssues != 0 || got.Commits != 0 || got.PRsMerged != 0 {
		t.Errorf("counts = issues %d, commits %d, merged %d; want all zero",
			got.OpenIssues, got.Commits, got.PRsMerged)
	}
}

func TestSummaryCycleTimeIsMedianNotMean(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	// Three fast pull requests and one that sat open for a fortnight. The mean
	// would be about 86 hours; the median is 4.
	mergedPR(t, db, projectID, 1, 6, 2)   // 4h
	mergedPR(t, db, projectID, 2, 7, 3)   // 4h
	mergedPR(t, db, projectID, 3, 8, 4)   // 4h
	mergedPR(t, db, projectID, 4, 338, 2) // 336h

	got, err := db.Analytics.Summary(context.Background(), projectID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got.PRCycleTimeHours == nil {
		t.Fatal("cycle time is nil, want a value")
	}
	if p50 := *got.PRCycleTimeHours; p50 < 3.9 || p50 > 4.1 {
		t.Errorf("median cycle time = %.2fh, want ~4h (a mean would be ~87h)", p50)
	}
	if got.PRCycleTimeP90Hours == nil {
		t.Fatal("p90 is nil, want a value")
	}
	// The outlier belongs in the p90, which is the point of reporting both.
	if p90 := *got.PRCycleTimeP90Hours; p90 < 100 {
		t.Errorf("p90 cycle time = %.2fh, want the outlier to show (>100h)", p90)
	}
	if got.PRsMerged != 4 {
		t.Errorf("merged = %d, want 4", got.PRsMerged)
	}
}

func TestSummaryMergeRateIgnoresOpenPullRequests(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	mergedPR(t, db, projectID, 1, 10, 2)
	mergedPR(t, db, projectID, 2, 10, 3)
	mergedPR(t, db, projectID, 3, 10, 4)
	closedPR(t, db, projectID, 4, 2)
	// Still open: undecided, so it must not drag the rate down.
	openPR(t, db, projectID, 5)
	openPR(t, db, projectID, 6)

	got, err := db.Analytics.Summary(context.Background(), projectID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got.PRsDecided != 4 {
		t.Errorf("decided = %d, want 4 (open pull requests excluded)", got.PRsDecided)
	}
	if got.PRMergeRate == nil {
		t.Fatal("merge rate is nil, want a value")
	}
	if rate := *got.PRMergeRate; rate < 0.74 || rate > 0.76 {
		t.Errorf("merge rate = %.3f, want 0.75", rate)
	}
}

func TestSummaryDeploymentFrequencyCountsProductionSuccesses(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	deployment(t, db, projectID, "1", domain.DeploymentSuccess, "production", 10)
	deployment(t, db, projectID, "2", domain.DeploymentSuccess, "production", 20)
	deployment(t, db, projectID, "3", domain.DeploymentFailure, "production", 30)
	// Neither of these counts: one is still running, one is not production.
	deployment(t, db, projectID, "4", domain.DeploymentRunning, "production", 1)
	deployment(t, db, projectID, "5", domain.DeploymentSuccess, "staging", 5)
	// Outside the window.
	deployment(t, db, projectID, "6", domain.DeploymentSuccess, "production", 24*45)

	got, err := db.Analytics.Summary(context.Background(), projectID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got.DeploymentsSucceeded != 2 {
		t.Errorf("succeeded = %d, want 2", got.DeploymentsSucceeded)
	}
	if got.DeploymentsFinished != 3 {
		t.Errorf("finished = %d, want 3 (running excluded)", got.DeploymentsFinished)
	}
	if got.DeploymentSuccessRate == nil {
		t.Fatal("success rate is nil, want a value")
	}
	if rate := *got.DeploymentSuccessRate; rate < 0.66 || rate > 0.67 {
		t.Errorf("success rate = %.3f, want 2/3", rate)
	}
	// Two successes over thirty days.
	if want := 2.0 / 30.0; got.DeploymentsPerDay < want-0.001 || got.DeploymentsPerDay > want+0.001 {
		t.Errorf("deployments per day = %.5f, want %.5f", got.DeploymentsPerDay, want)
	}
}

func TestSummaryOpenIssuesIgnoreTheWindow(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	issue(t, db, projectID, 1, domain.IssueOpen)
	issue(t, db, projectID, 2, domain.IssueOpen)
	issue(t, db, projectID, 3, domain.IssueClosed)

	// An issue open for a year is still open today, so a short window must
	// not hide it: this is a level, not a rate.
	got, err := db.Analytics.Summary(context.Background(), projectID, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got.OpenIssues != 2 {
		t.Errorf("open issues = %d, want 2", got.OpenIssues)
	}
}

func TestSummaryCommitFrequencyRespectsTheWindow(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	// Three inside a 7-day window (168h), two outside it but inside 30 days.
	for i, h := range []float64{1, 10, 30, 200, 24 * 20} {
		commit(t, db, projectID, string(rune('a'+i))+"0000000", h)
	}
	// Older than any window under test.
	commit(t, db, projectID, "z0000000", 24*120)

	got, err := db.Analytics.Summary(context.Background(), projectID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got.Commits != 5 {
		t.Errorf("commits = %d, want 5", got.Commits)
	}
	if want := 5.0 / 30.0; got.CommitsPerDay < want-0.001 || got.CommitsPerDay > want+0.001 {
		t.Errorf("commits per day = %.5f, want %.5f", got.CommitsPerDay, want)
	}

	// A 7-day window sees only the three recent ones, at a higher daily rate.
	week, err := db.Analytics.Summary(context.Background(), projectID, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if week.Commits != 3 {
		t.Errorf("commits in 7d = %d, want 3", week.Commits)
	}
	if week.CommitsPerDay <= got.CommitsPerDay {
		t.Errorf("7d rate %.4f should exceed 30d rate %.4f", week.CommitsPerDay, got.CommitsPerDay)
	}
}

func TestSummaryIsScopedToOneProject(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	var otherID string
	if err := db.Pool.QueryRow(context.Background(), `
		INSERT INTO projects (organization_id, name, slug)
		SELECT organization_id, 'Other', 'other' FROM projects WHERE id = $1
		RETURNING id`, projectID).Scan(&otherID); err != nil {
		t.Fatalf("create second project: %v", err)
	}

	commit(t, db, projectID, "aaaaaaa", 1)
	commit(t, db, otherID, "bbbbbbb", 1)
	issue(t, db, otherID, 1, domain.IssueOpen)

	got, err := db.Analytics.Summary(context.Background(), projectID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got.Commits != 1 {
		t.Errorf("commits = %d, want 1 — the other project's commit leaked in", got.Commits)
	}
	if got.OpenIssues != 0 {
		t.Errorf("open issues = %d, want 0 — the other project's issue leaked in", got.OpenIssues)
	}
}
