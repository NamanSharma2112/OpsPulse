package anomaly

import (
	"math"
	"strings"
	"testing"
)

func steady(value float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = value
	}
	return out
}

func TestFitUsesMedianNotMean(t *testing.T) {
	// Nine quiet days and one disaster. The mean is 109; the median is 10.
	history := []float64{10, 11, 9, 10, 12, 10, 9, 11, 10, 1000}

	b := Fit(history)
	if b.Median < 9.9 || b.Median > 10.6 {
		t.Errorf("median = %v, want ~10 (a mean would be ~109)", b.Median)
	}
	if b.Samples != 10 {
		t.Errorf("samples = %d, want 10", b.Samples)
	}
}

func TestDetectRefusesToJudgeWithoutEnoughHistory(t *testing.T) {
	for _, n := range []int{0, 1, 3, MinHistory - 1} {
		r := Detect(500, steady(10, n), Options{})
		if r.Anomaly {
			t.Errorf("with %d prior observations, got an anomaly; want none", n)
		}
		if !strings.Contains(r.Explanation, "Not enough history") {
			t.Errorf("with %d observations, explanation = %q, want it to say so", n, r.Explanation)
		}
	}

	// At the threshold it is willing to judge.
	r := Detect(500, steady(10, MinHistory), Options{})
	if !r.Anomaly {
		t.Errorf("with exactly %d observations, want a judgement; got %+v", MinHistory, r)
	}
}

func TestDetectFlagsASpike(t *testing.T) {
	history := []float64{10, 11, 9, 10, 12, 10, 9, 11, 10, 10}

	r := Detect(240, history, Options{})
	if !r.Anomaly {
		t.Fatalf("240 against a steady ~10 should be an anomaly; got %+v", r)
	}
	if r.Direction != Above {
		t.Errorf("direction = %q, want %q", r.Direction, Above)
	}
	if r.Deviation <= DefaultThreshold {
		t.Errorf("deviation = %v, want > %v", r.Deviation, DefaultThreshold)
	}
	if !strings.Contains(r.Explanation, "240") || !strings.Contains(r.Explanation, "above") {
		t.Errorf("explanation = %q, want it to name the value and the direction", r.Explanation)
	}
}

func TestDetectFlagsACollapse(t *testing.T) {
	// Deployment frequency falling off a cliff matters as much as a spike.
	history := []float64{50, 55, 48, 52, 51, 49, 53, 50, 47, 52}

	r := Detect(2, history, Options{})
	if !r.Anomaly {
		t.Fatalf("2 against a steady ~50 should be an anomaly; got %+v", r)
	}
	if r.Direction != Below {
		t.Errorf("direction = %q, want %q", r.Direction, Below)
	}
	if r.Deviation >= 0 {
		t.Errorf("deviation = %v, want negative for a drop", r.Deviation)
	}
}

func TestDetectIgnoresNormalVariation(t *testing.T) {
	history := []float64{10, 14, 9, 12, 16, 8, 11, 13, 10, 15}

	for _, v := range []float64{11, 13, 9, 16} {
		r := Detect(v, history, Options{})
		if r.Anomaly {
			t.Errorf("%v is ordinary for this series but was flagged: %+v", v, r)
		}
	}
}

// The property that motivates the whole choice of estimator.
func TestPastIncidentDoesNotMaskTheNextOne(t *testing.T) {
	// Last week had a genuine outage: one day at 400.
	history := []float64{10, 11, 400, 10, 12, 9, 11, 10, 9, 10}

	r := Detect(380, history, Options{})
	if !r.Anomaly {
		t.Fatalf("a repeat of last week's outage must still flag; got %+v", r)
	}

	// With mean and standard deviation this would not fire: the mean is ~49
	// and the standard deviation ~122, so 380 sits only 2.7 sigma out —
	// under the usual 3-sigma rule. Assert the baseline stayed robust.
	if r.Baseline.Median > 20 {
		t.Errorf("baseline median = %v, want ~10 — the old outage moved it",
			r.Baseline.Median)
	}
}

func TestFlatSeriesDoesNotProduceInfiniteDeviation(t *testing.T) {
	// Every observation identical: MAD and mean absolute deviation are both
	// zero, so a naive modified z-score divides by zero.
	r := Detect(5, steady(0, 10), Options{})

	if math.IsInf(r.Deviation, 0) || math.IsNaN(r.Deviation) {
		t.Fatalf("deviation = %v, want a finite number", r.Deviation)
	}
	if !r.Anomaly {
		t.Error("a jump off a perfectly flat series should still flag")
	}

	// And no movement at all is not an anomaly, however flat the series.
	quiet := Detect(0, steady(0, 10), Options{})
	if quiet.Anomaly {
		t.Errorf("no change on a flat series flagged: %+v", quiet)
	}
	if quiet.Deviation != 0 {
		t.Errorf("deviation = %v, want 0", quiet.Deviation)
	}
}

func TestMADZeroFallsBackToMeanAbsoluteDeviation(t *testing.T) {
	// Six identical values and four different ones: the MAD is zero because
	// more than half the points sit on the median, but the series is not
	// actually flat.
	history := []float64{10, 10, 10, 10, 10, 10, 12, 8, 14, 6}

	b := Fit(history)
	if b.Spread != 0 {
		t.Fatalf("precondition failed: spread = %v, want 0 for this series", b.Spread)
	}

	// 11 is well inside the real variation and must not flag.
	if r := Detect(11, history, Options{}); r.Anomaly {
		t.Errorf("11 is ordinary here but flagged: %+v", r)
	}
	// 200 clearly is not.
	if r := Detect(200, history, Options{}); !r.Anomaly {
		t.Errorf("200 should flag even though MAD is zero: %+v", r)
	}
}

func TestMinAbsoluteChangeSuppressesTrivialMovement(t *testing.T) {
	// A metric that is almost always zero: one deploy on one day is
	// statistically enormous and operationally meaningless.
	history := []float64{0, 0, 0, 0, 0, 0, 0, 0, 1, 0}

	if r := Detect(1, history, Options{}); !r.Anomaly {
		t.Fatalf("precondition: without a floor this should flag; got %+v", r)
	}

	r := Detect(1, history, Options{MinAbsoluteChange: 5})
	if r.Anomaly {
		t.Errorf("a movement of 1 should be suppressed by a floor of 5: %+v", r)
	}
	if !strings.Contains(r.Explanation, "floor") {
		t.Errorf("explanation = %q, want it to mention the floor", r.Explanation)
	}

	// The floor must not suppress a genuine spike.
	if big := Detect(90, history, Options{MinAbsoluteChange: 5}); !big.Anomaly {
		t.Errorf("90 should still flag with a floor of 5: %+v", big)
	}
}

func TestThresholdIsConfigurable(t *testing.T) {
	history := []float64{10, 11, 9, 10, 12, 10, 9, 11, 10, 10}

	// A value that sits just outside the default threshold.
	const probe = 16

	strict := Detect(probe, history, Options{Threshold: 50})
	if strict.Anomaly {
		t.Errorf("with a threshold of 50, %v should not flag: %+v", probe, strict)
	}
	loose := Detect(probe, history, Options{Threshold: 1})
	if !loose.Anomaly {
		t.Errorf("with a threshold of 1, %v should flag: %+v", probe, loose)
	}
}

func TestDetectDoesNotMutateHistory(t *testing.T) {
	// The caller reuses the slice across metrics, so sorting it in place
	// would corrupt the next evaluation.
	history := []float64{30, 10, 20, 50, 40, 15, 25, 35, 45, 5}
	before := make([]float64, len(history))
	copy(before, history)

	Detect(100, history, Options{})

	for i := range history {
		if history[i] != before[i] {
			t.Fatalf("history was reordered: got %v, want %v", history, before)
		}
	}
}
