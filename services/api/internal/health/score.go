// Package health turns the M2 analytics into the M3 engineering health score.
//
// It is deliberately pure: analytics in, score out, no database and no clock.
// A scoring rule is a product judgement that will be argued about and changed,
// and an argument is much easier to settle against a table of test cases than
// against a query.
package health

import (
	"fmt"
	"math"
	"strings"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

/*
Each signal is scored on a straight line between a target (100) and a floor
(0), clamped at both ends.

A straight line is the honest choice here. The thresholds themselves are
judgements, and dressing them in a curve would imply a precision about the
middle of the range that nobody has. What matters is that the direction and
the endpoints are defensible and written down.
*/

// scale maps a value onto 0–100 between a floor and a target.
//
// It handles both directions: when target < floor the metric is one where
// lower is better, like cycle time.
func scale(value, floor, target float64) int {
	if floor == target {
		// A degenerate rule would divide by zero; treat meeting the target as
		// a pass rather than panicking on a typo in a threshold.
		if value == target {
			return 100
		}
		return 0
	}
	ratio := (value - floor) / (target - floor)
	return int(math.Round(100 * math.Min(1, math.Max(0, ratio))))
}

// signal builds a scored signal, or an unknown one when the window held no
// data for it.
func signal(name, unit, target string, value *float64, floor, goal float64) domain.Signal {
	s := domain.Signal{Name: name, Unit: unit, Target: target}
	if value == nil {
		return s
	}
	v := *value
	score := scale(v, floor, goal)
	s.Value, s.Score, s.Known = &v, &score, true
	return s
}

// pillar averages the signals that are known. A pillar with nothing known
// scores nil rather than zero.
func pillar(name string, weight float64, signals []domain.Signal) domain.Pillar {
	p := domain.Pillar{Name: name, Weight: weight, Signals: signals}

	total, n := 0, 0
	for _, s := range signals {
		if s.Known {
			total += *s.Score
			n++
		}
	}
	if n == 0 {
		return p
	}
	score := int(math.Round(float64(total) / float64(n)))
	p.Score = &score
	return p
}

// Inputs are everything the score reads. Analytics carries the M2 window;
// OpenIncidents is a level and so is passed separately.
type Inputs struct {
	Analytics      domain.Analytics
	OpenIncidents  int
	MedianResolveH *float64
}

// Score computes the health of a project.
func Score(in Inputs, w domain.Weights) domain.Health {
	w = w.Normalized()
	a := in.Analytics

	incidents := float64(in.OpenIncidents)
	issues := float64(a.OpenIssues)

	reliability := pillar(domain.PillarReliability, w.Reliability, []domain.Signal{
		// A tenth of deployments failing is a bad week; none failing is the goal.
		signal("Deployment success rate", "%", "100%",
			percent(a.DeploymentSuccessRate), 70, 100),
		// Incidents open right now. Five at once is a floor, zero is the goal.
		signal("Open incidents", "", "0", &incidents, 5, 0),
		// Time to resolve. A day is the floor, an hour is the goal.
		signal("Median time to resolve", "h", "≤ 1h",
			in.MedianResolveH, 24, 1),
	})

	delivery := pillar(domain.PillarDelivery, w.Delivery, []domain.Signal{
		// One production deployment a day is the target; one a month is the floor.
		signal("Deployment frequency", "/day", "≥ 1/day",
			ptr(a.DeploymentsPerDay), 1.0/30.0, 1),
		// Most opened pull requests should land.
		signal("Pull request merge rate", "%", "≥ 90%",
			percent(a.PRMergeRate), 50, 90),
		// Sustained commit activity, not a burst.
		signal("Commit frequency", "/day", "≥ 5/day",
			ptr(a.CommitsPerDay), 0.2, 5),
	})

	performance := pillar(domain.PillarPerformance, w.Performance, []domain.Signal{
		// A pull request that merges inside a working day keeps its context.
		signal("Pull request cycle time", "h", "≤ 8h",
			a.PRCycleTimeHours, 120, 8),
		// The tail matters too: a good median hides a bad p90.
		signal("Cycle time (p90)", "h", "≤ 24h",
			a.PRCycleTimeP90Hours, 240, 24),
		// Backlog depth. Crude — it takes no account of team size — which is
		// why it is one signal of three rather than a pillar of its own.
		signal("Open issues", "", "≤ 10", &issues, 100, 10),
	})

	pillars := []domain.Pillar{reliability, delivery, performance}

	h := domain.Health{ProjectID: a.ProjectID, Window: a.Window, Pillars: pillars}

	// The headline re-weights across whichever pillars are known, so a project
	// with no deployments yet is still scored on what it does have rather than
	// being punished for the silence.
	var sum, weight float64
	for _, p := range pillars {
		if p.Score == nil {
			continue
		}
		sum += float64(*p.Score) * p.Weight
		weight += p.Weight
	}
	if weight > 0 {
		score := int(math.Round(sum / weight))
		h.Score = &score
	}
	h.Summary = summarize(h)
	return h
}

// summarize writes the one-line explanation. It names the weakest known
// signal, because that is the thing to go and fix.
func summarize(h domain.Health) string {
	if h.Score == nil {
		return "Not enough data yet — connect a repository and let some events arrive."
	}

	var worst *domain.Signal
	for i := range h.Pillars {
		for j := range h.Pillars[i].Signals {
			s := &h.Pillars[i].Signals[j]
			if !s.Known {
				continue
			}
			if worst == nil || *s.Score < *worst.Score {
				worst = s
			}
		}
	}

	band := "needs attention"
	switch {
	case *h.Score >= 85:
		band = "healthy"
	case *h.Score >= 70:
		band = "steady"
	case *h.Score >= 50:
		band = "strained"
	}

	if worst == nil || *worst.Score >= 85 {
		return fmt.Sprintf("Health %d — %s, with no signal below target.", *h.Score, band)
	}
	return fmt.Sprintf("Health %d — %s. The weakest signal is %s at %s against a target of %s.",
		*h.Score, band, strings.ToLower(worst.Name), format(*worst.Value, worst.Unit), worst.Target)
}

func format(v float64, unit string) string {
	switch unit {
	case "%":
		return fmt.Sprintf("%.0f%%", v)
	case "/day":
		return fmt.Sprintf("%.2f/day", v)
	case "h":
		return fmt.Sprintf("%.1fh", v)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

func ptr(v float64) *float64 { return &v }

// percent converts a 0–1 rate to a 0–100 percentage, preserving nil.
func percent(rate *float64) *float64 {
	if rate == nil {
		return nil
	}
	v := *rate * 100
	return &v
}
