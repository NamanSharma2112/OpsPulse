// Package anomaly turns a metric's history into a baseline, a deviation and,
// when the deviation is large enough, an anomaly worth raising an incident
// for.
//
// The detection maths is pure and lives apart from the storage and the
// scheduling, because the question "is 240 unusual for this metric" is the
// part that has to be right, and it is the part that is cheapest to get
// wrong silently.
package anomaly

import (
	"fmt"
	"math"
	"sort"
)

/*
Why median and MAD rather than mean and standard deviation.

A baseline exists to answer "what does normal look like here". Mean and
standard deviation answer it badly for operational data, because the
incidents you most want to catch are already in the history you are learning
from: one four-hour outage last week inflates the standard deviation enough
that this week's outage looks ordinary. The anomaly poisons the baseline that
is supposed to detect it.

The median and the median absolute deviation have a breakdown point of 50%,
so half the history can be garbage before the baseline moves. That is the
right trade for a system whose input is, by definition, full of incidents.

The 0.6745 scaling makes the modified z-score comparable to a standard score
for normally distributed data, which is what lets 3.5 be a meaningful
threshold rather than an arbitrary one (Iglewicz and Hoaglin, 1993).
*/

const (
	// madScale makes MAD a consistent estimator of the standard deviation
	// for normal data: 0.6745 is the 75th percentile of the standard normal.
	madScale = 0.6745

	// meanADScale does the same job for the mean absolute deviation, used
	// only when MAD collapses to zero.
	meanADScale = 1.253314

	// DefaultThreshold is the modified z-score at which a point is an
	// anomaly. 3.5 is the conventional value for this test.
	DefaultThreshold = 3.5

	// MinHistory is how many prior observations a baseline needs before it
	// will claim anything. Below this, three quiet days followed by a normal
	// one would read as a spike.
	MinHistory = 7
)

// Direction says which way a metric moved.
const (
	Above = "above"
	Below = "below"
)

// Baseline is what normal looks like for one metric.
type Baseline struct {
	Median float64 `json:"median"`
	// MAD is the median absolute deviation, already scaled so it is
	// comparable to a standard deviation.
	Spread  float64 `json:"spread"`
	Samples int     `json:"samples"`
}

// Result is one evaluation of an observation against its baseline.
type Result struct {
	Observed  float64  `json:"observed"`
	Baseline  Baseline `json:"baseline"`
	Deviation float64  `json:"deviation"`
	Direction string   `json:"direction"`
	Anomaly   bool     `json:"anomaly"`
	// Explanation says, in a sentence, what was seen and why it was or was
	// not unusual. It is populated even when there is no anomaly, because
	// "this looked fine" is also worth being able to show.
	Explanation string `json:"explanation"`
}

// Options tune a detection run.
type Options struct {
	// Threshold is the modified z-score at which a point counts. Zero means
	// DefaultThreshold.
	Threshold float64
	// MinAbsoluteChange suppresses anomalies whose raw movement is trivially
	// small. A metric that sits at exactly 0 all week has a spread of zero,
	// which makes any non-zero reading infinitely deviant — statistically
	// true and operationally useless. Zero means no floor.
	MinAbsoluteChange float64
}

func (o Options) threshold() float64 {
	if o.Threshold <= 0 {
		return DefaultThreshold
	}
	return o.Threshold
}

// Detect evaluates observed against the baseline implied by history.
//
// history must hold the prior observations only — the observed value must not
// be included, or it drags its own baseline towards itself.
func Detect(observed float64, history []float64, opts Options) Result {
	base := Fit(history)
	r := Result{Observed: observed, Baseline: base}

	if base.Samples < MinHistory {
		r.Explanation = fmt.Sprintf(
			"Not enough history to judge: %d prior observations, %d needed.",
			base.Samples, MinHistory)
		return r
	}

	delta := observed - base.Median
	r.Direction = Above
	if delta < 0 {
		r.Direction = Below
	}

	r.Deviation = deviation(delta, base.Spread, history, base.Median)

	switch {
	case math.Abs(delta) < opts.MinAbsoluteChange:
		r.Explanation = fmt.Sprintf(
			"%.2f against a usual %.2f — within the %.2f absolute floor, so not flagged.",
			observed, base.Median, opts.MinAbsoluteChange)
	case math.Abs(r.Deviation) >= opts.threshold():
		r.Anomaly = true
		r.Explanation = fmt.Sprintf(
			"%.2f is %s the usual %.2f by %.1f×, well outside the normal spread of %.2f "+
				"measured over %d prior observations.",
			observed, r.Direction, base.Median, math.Abs(r.Deviation), base.Spread, base.Samples)
	default:
		r.Explanation = fmt.Sprintf(
			"%.2f against a usual %.2f — %.1f× the normal spread, inside the %.1f threshold.",
			observed, base.Median, math.Abs(r.Deviation), opts.threshold())
	}
	return r
}

// deviation is the modified z-score, with the fallbacks that keep it finite.
func deviation(delta, spread float64, history []float64, median float64) float64 {
	if spread > 0 {
		return madScale * delta / spread
	}

	// MAD collapses to zero whenever more than half the history is identical
	// — common for a metric that is usually flat. Fall back to the mean
	// absolute deviation, which only collapses when every point is identical.
	if mad := meanAbsoluteDeviation(history, median); mad > 0 {
		return delta / (meanADScale * mad)
	}

	// Every historical point is the same. Any change is unbounded deviation;
	// no change is none. Returning ±inf here would be true but would also
	// make every downstream format verb print "+Inf", so it is capped at a
	// value that is unambiguously over any sane threshold.
	if delta == 0 {
		return 0
	}
	return math.Copysign(1e3, delta)
}

// Fit computes the baseline for a series.
func Fit(history []float64) Baseline {
	b := Baseline{Samples: len(history)}
	if len(history) == 0 {
		return b
	}

	b.Median = median(history)

	deviations := make([]float64, len(history))
	for i, v := range history {
		deviations[i] = math.Abs(v - b.Median)
	}
	b.Spread = median(deviations)
	return b
}

// median sorts a copy, so the caller's slice is left alone — the history
// belongs to the caller and may be reused for another metric.
func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func meanAbsoluteDeviation(values []float64, center float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, v := range values {
		total += math.Abs(v - center)
	}
	return total / float64(len(values))
}
