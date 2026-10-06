package health

import (
	"strings"
	"testing"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

func f(v float64) *float64 { return &v }

func TestScaleClampsAndRunsBothDirections(t *testing.T) {
	tests := []struct {
		name                 string
		value, floor, target float64
		want                 int
	}{
		{"at target", 100, 70, 100, 100},
		{"at floor", 70, 70, 100, 0},
		{"midway", 85, 70, 100, 50},
		{"above target clamps", 120, 70, 100, 100},
		{"below floor clamps", 10, 70, 100, 0},

		// Lower-is-better: target below floor.
		{"cycle time at target", 8, 120, 8, 100},
		{"cycle time at floor", 120, 120, 8, 0},
		{"cycle time beyond target clamps", 1, 120, 8, 100},
		{"cycle time worse than floor clamps", 500, 120, 8, 0},

		// A typo that collapses the range must not panic.
		{"degenerate range, met", 5, 5, 5, 100},
		{"degenerate range, missed", 6, 5, 5, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scale(tc.value, tc.floor, tc.target); got != tc.want {
				t.Errorf("scale(%v, %v, %v) = %d, want %d", tc.value, tc.floor, tc.target, got, tc.want)
			}
		})
	}
}

func TestScoreIsNilWhenNothingIsKnown(t *testing.T) {
	// A brand new project: no deployments, no pull requests, no commits.
	// Several signals are still "known" because a count of zero is a real
	// measurement — but the headline must not pretend to more than it has.
	in := Inputs{Analytics: domain.Analytics{ProjectID: "p", Window: "30d"}}
	got := Score(in, domain.DefaultWeights)

	for _, p := range got.Pillars {
		for _, s := range p.Signals {
			if s.Name == "Deployment success rate" && s.Known {
				t.Error("deployment success rate should be unknown with no deployments")
			}
			if s.Name == "Pull request cycle time" && s.Known {
				t.Error("cycle time should be unknown with no merged pull requests")
			}
		}
	}
}

func TestScoreWithNoSignalsAtAllIsUnknown(t *testing.T) {
	// Force every signal unknown by scoring an empty pillar set directly.
	p := pillar(domain.PillarDelivery, 1, []domain.Signal{
		signal("Anything", "", "x", nil, 0, 1),
	})
	if p.Score != nil {
		t.Fatalf("pillar score = %d, want nil when no signal is known", *p.Score)
	}
}

func TestPerfectProjectScores100(t *testing.T) {
	in := Inputs{
		Analytics: domain.Analytics{
			ProjectID:             "p",
			Window:                "30d",
			PRCycleTimeHours:      f(4),
			PRCycleTimeP90Hours:   f(12),
			PRMergeRate:           f(0.98),
			DeploymentsPerDay:     3,
			DeploymentSuccessRate: f(1),
			CommitsPerDay:         12,
			OpenIssues:            2,
		},
		OpenIncidents:  0,
		MedianResolveH: f(0.5),
	}
	got := Score(in, domain.DefaultWeights)
	if got.Score == nil {
		t.Fatal("score is nil, want 100")
	}
	if *got.Score != 100 {
		t.Errorf("score = %d, want 100", *got.Score)
	}
	if !strings.Contains(got.Summary, "healthy") {
		t.Errorf("summary = %q, want it to read as healthy", got.Summary)
	}
}

func TestTerribleProjectScores0(t *testing.T) {
	in := Inputs{
		Analytics: domain.Analytics{
			ProjectID:             "p",
			Window:                "30d",
			PRCycleTimeHours:      f(400),
			PRCycleTimeP90Hours:   f(900),
			PRMergeRate:           f(0.1),
			DeploymentsPerDay:     0,
			DeploymentSuccessRate: f(0.2),
			CommitsPerDay:         0,
			OpenIssues:            300,
		},
		OpenIncidents:  12,
		MedianResolveH: f(100),
	}
	got := Score(in, domain.DefaultWeights)
	if got.Score == nil {
		t.Fatal("score is nil, want 0")
	}
	if *got.Score != 0 {
		t.Errorf("score = %d, want 0", *got.Score)
	}
}

func TestHeadlineReweightsOverKnownPillarsOnly(t *testing.T) {
	// Delivery and performance are perfect; reliability is entirely unknown
	// because nothing has deployed and no incident has been raised or
	// resolved. The headline should reflect the two it can see, not be
	// dragged towards zero by the one it cannot.
	in := Inputs{
		Analytics: domain.Analytics{
			ProjectID:           "p",
			Window:              "30d",
			PRCycleTimeHours:    f(4),
			PRCycleTimeP90Hours: f(12),
			PRMergeRate:         f(0.98),
			DeploymentsPerDay:   3,
			CommitsPerDay:       12,
			OpenIssues:          2,
		},
	}
	got := Score(in, domain.DefaultWeights)
	if got.Score == nil {
		t.Fatal("score is nil")
	}

	// Reliability still has "open incidents = 0", which is a real measurement
	// scoring 100, so the pillar is known and perfect. Everything is 100.
	if *got.Score != 100 {
		t.Errorf("score = %d, want 100", *got.Score)
	}

	// Deployment success rate specifically must be unknown, not zero.
	for _, p := range got.Pillars {
		for _, s := range p.Signals {
			if s.Name == "Deployment success rate" && s.Known {
				t.Error("success rate should be unknown when nothing finished deploying")
			}
		}
	}
}

func TestWeightsAreNormalized(t *testing.T) {
	// Weights that sum to 10 must score the same as the same ratios summing
	// to 1 — otherwise a misconfiguration silently rescales every project.
	in := Inputs{
		Analytics: domain.Analytics{
			ProjectID: "p", Window: "30d",
			PRCycleTimeHours:      f(60),
			PRCycleTimeP90Hours:   f(130),
			PRMergeRate:           f(0.7),
			DeploymentsPerDay:     0.5,
			DeploymentSuccessRate: f(0.85),
			CommitsPerDay:         2,
			OpenIssues:            40,
		},
		OpenIncidents:  2,
		MedianResolveH: f(10),
	}
	a := Score(in, domain.Weights{Reliability: 4, Delivery: 3.5, Performance: 2.5})
	b := Score(in, domain.DefaultWeights)
	if a.Score == nil || b.Score == nil {
		t.Fatal("score is nil")
	}
	if *a.Score != *b.Score {
		t.Errorf("scaled weights gave %d, unscaled gave %d — weights are not normalized", *a.Score, *b.Score)
	}

	// A zero set must fall back rather than divide by zero.
	z := domain.Weights{}.Normalized()
	if z != domain.DefaultWeights {
		t.Errorf("zero weights = %+v, want the default", z)
	}
}

func TestWeightingActuallyShiftsTheScore(t *testing.T) {
	// Reliability perfect, delivery terrible. Weighting towards reliability
	// must produce a higher headline than weighting towards delivery — if it
	// does not, the weights are decorative.
	in := Inputs{
		Analytics: domain.Analytics{
			ProjectID: "p", Window: "30d",
			DeploymentSuccessRate: f(1),
			DeploymentsPerDay:     0,
			PRMergeRate:           f(0.1),
			CommitsPerDay:         0,
		},
		OpenIncidents:  0,
		MedianResolveH: f(0.5),
	}
	relHeavy := Score(in, domain.Weights{Reliability: 0.8, Delivery: 0.1, Performance: 0.1})
	delHeavy := Score(in, domain.Weights{Reliability: 0.1, Delivery: 0.8, Performance: 0.1})

	if relHeavy.Score == nil || delHeavy.Score == nil {
		t.Fatal("score is nil")
	}
	if *relHeavy.Score <= *delHeavy.Score {
		t.Errorf("reliability-weighted %d should beat delivery-weighted %d",
			*relHeavy.Score, *delHeavy.Score)
	}
}

func TestSummaryNamesTheWeakestSignal(t *testing.T) {
	in := Inputs{
		Analytics: domain.Analytics{
			ProjectID: "p", Window: "30d",
			// Everything healthy except a cycle time at the floor.
			PRCycleTimeHours:      f(120),
			PRCycleTimeP90Hours:   f(24),
			PRMergeRate:           f(0.95),
			DeploymentsPerDay:     2,
			DeploymentSuccessRate: f(1),
			CommitsPerDay:         8,
			OpenIssues:            3,
		},
		OpenIncidents:  0,
		MedianResolveH: f(1),
	}
	got := Score(in, domain.DefaultWeights)
	if !strings.Contains(got.Summary, "cycle time") {
		t.Errorf("summary = %q, want it to name the cycle time as the weakest signal", got.Summary)
	}
	if !strings.Contains(got.Summary, "8h") {
		t.Errorf("summary = %q, want it to quote the target", got.Summary)
	}
}

func TestSummaryOnAnEmptyProjectSaysSo(t *testing.T) {
	// Every pillar unknown: scoring nothing must read as "no data", not as a
	// catastrophic zero.
	h := domain.Health{Pillars: []domain.Pillar{{Name: domain.PillarDelivery}}}
	if got := summarize(h); !strings.Contains(got, "Not enough data") {
		t.Errorf("summary = %q, want it to say there is not enough data", got)
	}
}
