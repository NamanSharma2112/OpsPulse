# 6. Health score weighting is a product judgement, not a derived constant

Date: 2026-10-06

## Status

Accepted.

## Context

M3 asks for one number — `HEALTH: 82` — over three pillars: Reliability,
Delivery and Performance. Producing one number from nine signals means
choosing how much each one counts, and there is no way to derive those
choices from the data. They are opinions about what a healthy engineering
team looks like.

Three things had to be settled.

**How much each pillar counts.** Equal thirds is the obvious default and the
wrong one: a team that ships three times a day and takes production down
every other deploy is not as healthy as a team that ships weekly and stays
up. Reliability is load-bearing in a way raw speed is not.

**How a raw measurement becomes a 0–100 score.** Every signal needs a target
(100) and a floor (0).

**What happens when a signal is missing.** A new project has merged no pull
requests. Scoring that as zero would say "catastrophic" when the truth is
"nothing has happened yet".

## Decision

**Weights default to Reliability 0.40, Delivery 0.35, Performance 0.25**, and
are configurable per deployment via `HEALTH_WEIGHT_RELIABILITY`,
`HEALTH_WEIGHT_DELIVERY` and `HEALTH_WEIGHT_PERFORMANCE`. They are normalized
at the point of use, so a set that does not sum to 1 rescales rather than
silently inflating every project's score.

**Each signal scores on a straight line** between a floor and a target,
clamped at both ends:

| Pillar | Signal | Floor (0) | Target (100) |
| --- | --- | --- | --- |
| Reliability | Deployment success rate | 70% | 100% |
| Reliability | Open incidents | 5 | 0 |
| Reliability | Median time to resolve | 24h | 1h |
| Delivery | Deployment frequency | 1/month | 1/day |
| Delivery | Pull request merge rate | 50% | 90% |
| Delivery | Commit frequency | 0.2/day | 5/day |
| Performance | PR cycle time (p50) | 120h | 8h |
| Performance | PR cycle time (p90) | 240h | 24h |
| Performance | Open issues | 100 | 10 |

A straight line, not a curve. The thresholds are already judgements; a curve
would imply a precision about the middle of the range that nobody has. What
matters is that the direction and the endpoints are defensible and written
down where they can be argued with.

**An unknown signal is skipped, not scored.** This is why the M2 analytics
carry `nil` rather than `0` for every rate. A pillar with no known signals
scores `nil`, and the headline re-weights across whichever pillars are known
— so a project with no deployments yet is scored on what it does have rather
than punished for the silence. A project with nothing at all scores `nil`,
and the UI says so in words instead of showing a zero.

## Consequences

The scoring lives in `internal/health`, which is pure: analytics in, score
out, no database and no clock. A scoring rule is a thing people will argue
about, and an argument is far easier to settle against a table of test cases
than against a query. The test suite pins the endpoints, the clamping, both
directions of travel, the nil handling, and — importantly — that changing the
weights actually changes the score, so the configuration cannot quietly
become decorative.

The open-issues signal is the weakest of the nine: a 40-issue backlog means
something different for a team of three than for a team of thirty. It is one
signal of three inside the lightest-weighted pillar, which bounds the damage.
Normalising it by team size needs a notion of team size the product does not
have yet.

These numbers are a starting point, not a finding. They should be revisited
once there is real usage to calibrate against — at which point this record
should be superseded rather than edited, so the reasoning that produced the
first set stays legible.
