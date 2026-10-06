package anomaly

import (
	"context"
	"log/slog"
	"time"
)

// Projects is the slice of the project store the sweep needs.
type Projects interface {
	ListAll(ctx context.Context) ([]string, error)
}

// Sweep evaluates every project on an interval.
//
// This is the piece M4's worker will take over. It lives here rather than in
// main so that the move is a change of caller: the worker would call Run with
// the same arguments, and nothing else would change.
type Sweep struct {
	svc      *Service
	projects Projects
	interval time.Duration
	log      *slog.Logger
}

// NewSweep builds the periodic evaluator.
func NewSweep(svc *Service, projects Projects, interval time.Duration, log *slog.Logger) *Sweep {
	return &Sweep{svc: svc, projects: projects, interval: interval, log: log}
}

// Run evaluates every project until the context is cancelled.
//
// It returns on cancellation rather than treating it as an error: a shutdown
// is the normal way this ends.
func (s *Sweep) Run(ctx context.Context) {
	if s.interval <= 0 {
		s.log.Info("anomaly detection is disabled")
		return
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("anomaly detection started", "interval", s.interval)
	for {
		select {
		case <-ctx.Done():
			s.log.Info("anomaly detection stopped")
			return
		case <-ticker.C:
			s.once(ctx)
		}
	}
}

// once is one pass over every project.
func (s *Sweep) once(ctx context.Context) {
	// Bounded so a slow pass cannot overlap the next tick or outlive a
	// shutdown that is already underway.
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	ids, err := s.projects.ListAll(ctx)
	if err != nil {
		s.log.Error("anomaly sweep could not list projects", "error", err)
		return
	}

	raised := 0
	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}
		found, err := s.svc.Evaluate(ctx, id)
		if err != nil {
			// One project's failure must not stop the rest of the sweep.
			s.log.Error("anomaly evaluation failed", "error", err, "project_id", id)
			continue
		}
		raised += len(found)
		for _, a := range found {
			s.log.Info("anomaly detected",
				"project_id", id, "metric", a.Metric, "observed", a.Observed,
				"baseline", a.BaselineMedian, "deviation", a.Deviation)
		}
	}
	if raised > 0 {
		s.log.Info("anomaly sweep complete", "projects", len(ids), "raised", raised)
	}
}
