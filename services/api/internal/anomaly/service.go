package anomaly

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"unicode"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

/*
Scheduling note.

M4 puts this work in a Redis-backed worker. That is deferred, so for now the
evaluation runs on a ticker inside the API process. The seam is the same
either way: Evaluate takes a project id and does one complete pass, so moving
it behind a queue later is a change of caller, not a rewrite.
*/

// Watched is the metric set evaluated for anomalies, with the floor below
// which a movement is too small to be worth a page.
//
// The floors are what stop a quiet repository generating noise: a project
// that normally deploys zero times a day has a baseline spread of zero, so a
// single deploy is infinitely deviant and completely unremarkable.
var Watched = []struct {
	Metric string
	Label  string
	Floor  float64
}{
	{domain.MetricDeploymentFailed, "failed deployments", 2},
	{domain.MetricIncidentOpened, "incidents opened", 2},
	{domain.MetricCommitPushed, "commits pushed", 5},
	{domain.MetricPullRequestMerged, "pull requests merged", 3},
	{domain.MetricDeploymentSucceeded, "successful deployments", 2},
}

// History is how many days of baseline each evaluation reads. Thirty days
// covers four weeks, so a weekly rhythm is inside the baseline rather than
// being read as an anomaly every Monday.
const History = 30

// Service evaluates projects for anomalies and raises incidents for them.
type Service struct {
	metrics interface {
		DailySeries(ctx context.Context, projectID, name string, from, to time.Time) ([]domain.Bucket, error)
	}
	anomalies store.Anomalies
	incidents store.Incidents
	log       *slog.Logger
}

// NewService wires the detector to its storage.
func NewService(
	metrics interface {
		DailySeries(ctx context.Context, projectID, name string, from, to time.Time) ([]domain.Bucket, error)
	},
	anomalies store.Anomalies,
	incidents store.Incidents,
	log *slog.Logger,
) *Service {
	return &Service{metrics: metrics, anomalies: anomalies, incidents: incidents, log: log}
}

// Evaluate runs every watched metric for one project and returns the
// anomalies it newly raised.
func (s *Service) Evaluate(ctx context.Context, projectID string) ([]domain.Anomaly, error) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -History)

	var raised []domain.Anomaly
	for _, w := range Watched {
		series, err := s.metrics.DailySeries(ctx, projectID, w.Metric, from, to)
		if err != nil {
			return nil, fmt.Errorf("series for %s: %w", w.Metric, err)
		}
		// The last bucket is today, which is still filling up: judging it
		// would flag every morning as a collapse. Yesterday is the most
		// recent complete day.
		if len(series) < MinHistory+2 {
			continue
		}
		observed := series[len(series)-2]

		history := make([]float64, 0, len(series)-2)
		for _, b := range series[:len(series)-2] {
			history = append(history, b.Value)
		}

		result := Detect(observed.Value, history, Options{MinAbsoluteChange: w.Floor})
		if !result.Anomaly {
			continue
		}

		a := domain.Anomaly{
			ProjectID:      projectID,
			Metric:         w.Metric,
			BucketStart:    observed.Start,
			Observed:       result.Observed,
			BaselineMedian: result.Baseline.Median,
			BaselineSpread: result.Baseline.Spread,
			Deviation:      result.Deviation,
			Direction:      result.Direction,
			Explanation:    explain(w.Label, observed.Start, result),
		}

		fresh, err := s.anomalies.Insert(ctx, &a)
		if err != nil {
			return nil, fmt.Errorf("record anomaly for %s: %w", w.Metric, err)
		}
		if !fresh {
			// Already raised on an earlier pass over the same bucket.
			continue
		}

		if err := s.raise(ctx, &a, w.Label); err != nil {
			// The anomaly is recorded either way; failing to open an incident
			// is logged rather than losing the detection.
			s.log.Error("could not raise incident for anomaly",
				"error", err, "project_id", projectID, "metric", w.Metric)
		}
		raised = append(raised, a)
	}
	return raised, nil
}

// raise opens an incident for an anomaly.
func (s *Service) raise(ctx context.Context, a *domain.Anomaly, label string) error {
	// Keyed by metric and bucket so the incident is the same object the
	// anomaly is, and a replay cannot open a second one.
	externalID := fmt.Sprintf("anomaly:%s:%s", a.Metric, a.BucketStart.Format("2006-01-02"))

	incident := domain.Incident{
		ProjectID:  a.ProjectID,
		ExternalID: externalID,
		Title:      fmt.Sprintf("Unusual %s on %s", label, a.BucketStart.Format("2 Jan")),
		Severity:   severityFor(a.Deviation),
		Status:     domain.IncidentOpen,
		Source:     "anomaly",
		OpenedAt:   a.BucketStart,
	}
	if err := s.incidents.Upsert(ctx, &incident); err != nil {
		return err
	}
	a.IncidentID = &incident.ID
	return nil
}

// severityFor grades an anomaly by how far outside normal it fell.
func severityFor(deviation float64) string {
	d := deviation
	if d < 0 {
		d = -d
	}
	switch {
	case d >= 10:
		return domain.SeverityCritical
	case d >= 6:
		return domain.SeverityMajor
	default:
		return domain.SeverityMinor
	}
}

// explain writes the human-readable sentence stored with the anomaly. This is
// the product's whole thesis — a number nobody noticed becoming something
// somebody can act on — so it names the metric, the day, the two values and
// the usual range rather than printing a z-score.
func explain(label string, day time.Time, r Result) string {
	direction := "above"
	if r.Direction == Below {
		direction = "below"
	}
	return fmt.Sprintf(
		"%s on %s: %.0f, %s the usual %.0f. Over the previous %d days this metric "+
			"stayed within about %.1f of its median, so this reading is %.1f× that spread.",
		capitalize(label), day.Format("2 January"), r.Observed, direction,
		r.Baseline.Median, r.Baseline.Samples, r.Baseline.Spread, abs(r.Deviation))
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
