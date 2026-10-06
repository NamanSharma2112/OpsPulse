package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// metricSample records one derived sample at a point in time.
func metricSample(t *testing.T, db *DB, projectID, name string, value float64, at time.Time) {
	t.Helper()
	if _, err := db.Pool.Exec(context.Background(), `
		INSERT INTO metrics (project_id, name, value, recorded_at) VALUES ($1, $2, $3, $4)`,
		projectID, name, value, at); err != nil {
		t.Fatalf("insert metric: %v", err)
	}
}

func TestDailySeriesFillsEmptyDaysWithZero(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	now := time.Now().UTC()
	// Samples on two days only, inside a ten-day window.
	metricSample(t, db, projectID, domain.MetricCommitPushed, 3, now.AddDate(0, 0, -9))
	metricSample(t, db, projectID, domain.MetricCommitPushed, 2, now.AddDate(0, 0, -9))
	metricSample(t, db, projectID, domain.MetricCommitPushed, 7, now.AddDate(0, 0, -2))

	series, err := db.Metrics.DailySeries(context.Background(), projectID,
		domain.MetricCommitPushed, now.AddDate(0, 0, -9), now)
	if err != nil {
		t.Fatalf("daily series: %v", err)
	}

	// Ten days inclusive, every one present — a day with no activity must be
	// a zero, not a gap, or the baseline never learns what quiet looks like.
	if len(series) != 10 {
		t.Fatalf("got %d buckets, want 10 (one per day, gaps filled)", len(series))
	}

	var total float64
	zeros := 0
	for _, b := range series {
		total += b.Value
		if b.Value == 0 {
			zeros++
		}
	}
	if total != 12 {
		t.Errorf("total = %v, want 12 (3+2 on one day, 7 on another)", total)
	}
	if zeros != 8 {
		t.Errorf("zero days = %d, want 8", zeros)
	}
	// Same-day samples must be summed into one bucket.
	if series[0].Value != 5 {
		t.Errorf("first bucket = %v, want 5 (two samples summed)", series[0].Value)
	}
}

func TestDailySeriesIsScopedToProjectAndMetric(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	var otherID string
	if err := db.Pool.QueryRow(context.Background(), `
		INSERT INTO projects (organization_id, name, slug)
		SELECT organization_id, 'Other', 'other' FROM projects WHERE id = $1
		RETURNING id`, projectID).Scan(&otherID); err != nil {
		t.Fatalf("create second project: %v", err)
	}

	now := time.Now().UTC()
	metricSample(t, db, projectID, domain.MetricCommitPushed, 5, now.AddDate(0, 0, -1))
	metricSample(t, db, projectID, domain.MetricDeploymentFailed, 99, now.AddDate(0, 0, -1))
	metricSample(t, db, otherID, domain.MetricCommitPushed, 42, now.AddDate(0, 0, -1))

	series, err := db.Metrics.DailySeries(context.Background(), projectID,
		domain.MetricCommitPushed, now.AddDate(0, 0, -3), now)
	if err != nil {
		t.Fatalf("daily series: %v", err)
	}

	var total float64
	for _, b := range series {
		total += b.Value
	}
	if total != 5 {
		t.Errorf("total = %v, want 5 — another project's or metric's samples leaked in", total)
	}
}

func TestAnomalyInsertIsIdempotentPerBucket(t *testing.T) {
	db := testDB(t)
	projectID := fixture(t, db)

	bucket := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -1)
	a := domain.Anomaly{
		ProjectID: projectID, Metric: domain.MetricDeploymentFailed,
		BucketStart: bucket, Observed: 40, BaselineMedian: 1, BaselineSpread: 1,
		Deviation: 26, Direction: "above", Explanation: "spike",
	}

	fresh, err := db.Anomalies.Insert(context.Background(), &a)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if !fresh {
		t.Fatal("first insert reported not fresh")
	}

	// The sweep re-judges the same bucket every fifteen minutes. The second
	// pass must be a no-op, or one spike pages the team all day.
	again := a
	again.ID = ""
	fresh, err = db.Anomalies.Insert(context.Background(), &again)
	if err != nil {
		t.Fatalf("second insert: %v", err)
	}
	if fresh {
		t.Error("re-judging the same bucket raised a second anomaly")
	}

	list, err := db.Anomalies.ListForProject(context.Background(), projectID, 25)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("stored %d anomalies, want 1", len(list))
	}
}
