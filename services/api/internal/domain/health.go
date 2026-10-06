package domain

// Pillar names. Three, because they answer three different questions:
// does what we ship stay up, are we shipping at all, and how fast does work
// move through the system.
const (
	PillarReliability = "Reliability"
	PillarDelivery    = "Delivery"
	PillarPerformance = "Performance"
)

// Health is the M3 score: one headline number over three pillars, each built
// from signals that are reported individually.
//
// Every score is optional. A project with no data has no health — reporting
// zero would read as "catastrophic" when the truth is "nothing has happened
// yet", and that distinction is the whole reason the M2 analytics carry nil
// rather than zero.
type Health struct {
	ProjectID string   `json:"project_id"`
	Window    string   `json:"window"`
	Score     *int     `json:"score"`
	Pillars   []Pillar `json:"pillars"`
	// Why the score is what it is, in a sentence, for the people who will
	// never open the breakdown.
	Summary string `json:"summary"`
}

// Pillar is one branch of the score.
type Pillar struct {
	Name    string   `json:"name"`
	Score   *int     `json:"score"`
	Weight  float64  `json:"weight"`
	Signals []Signal `json:"signals"`
}

// Signal is one measured input and how it scored.
//
// Target is the human-readable goal the score was measured against, so the
// dashboard can say "4.2h against a target of 8h" rather than leaving someone
// to guess what a 74 means.
type Signal struct {
	Name   string   `json:"name"`
	Value  *float64 `json:"value"`
	Unit   string   `json:"unit"`
	Score  *int     `json:"score"`
	Target string   `json:"target"`
	// Known is false when the window held no data for this signal. An unknown
	// signal is skipped rather than scored as zero.
	Known bool `json:"known"`
}

// Weights decide how much each pillar contributes to the headline score.
//
// The default leans on reliability because a team that ships fast and breaks
// production is not healthy, while a team that ships steadily and stays up
// mostly is. It is a product judgement rather than a derived constant, which
// is exactly why it is configurable — see docs/decisions/0006.
type Weights struct {
	Reliability float64 `json:"reliability"`
	Delivery    float64 `json:"delivery"`
	Performance float64 `json:"performance"`
}

// DefaultWeights is the shipped judgement: reliability first, delivery
// second, speed last.
var DefaultWeights = Weights{Reliability: 0.40, Delivery: 0.35, Performance: 0.25}

// Normalized returns the weights scaled to sum to 1, so a misconfigured set
// cannot silently inflate or deflate every score. A set that sums to zero
// falls back to the default.
func (w Weights) Normalized() Weights {
	total := w.Reliability + w.Delivery + w.Performance
	if total <= 0 {
		return DefaultWeights
	}
	return Weights{
		Reliability: w.Reliability / total,
		Delivery:    w.Delivery / total,
		Performance: w.Performance / total,
	}
}
