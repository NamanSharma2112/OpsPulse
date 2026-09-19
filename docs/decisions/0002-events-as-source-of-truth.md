# 2. Store raw GitHub events as the source of truth

- **Status:** accepted
- **Date:** 2026-09-19

## Context

OpsPulse renders deployments, pull requests and incidents. The obvious design
is to write each webhook straight into the table the dashboard reads.

That design loses information. GitHub sends a payload far richer than any
projection keeps, and the fields we care about change as the product grows.
Once a delivery has been reduced to a row, the discarded part is gone: adding
"time to first review" later would mean waiting weeks for fresh data.

Webhook delivery is also at-least-once. GitHub retries on non-2xx, and its
UI offers manual redelivery, so the same event arrives more than once.

## Decision

Every verified delivery is inserted into `events` verbatim, with its payload
in a `jsonb` column, before anything else happens. The projections
(`deployments`, `pull_requests`, `incidents`) and `metrics` are derived from
it and are treated as disposable caches.

`events.delivery_id` — GitHub's `X-GitHub-Delivery` — carries a unique
constraint. Ingest inserts with `ON CONFLICT DO NOTHING` and skips projection
when no row was inserted, which makes the whole path idempotent.

A projection failure is logged, not returned. The event is already durable, so
failing the request would only make GitHub retry something we accepted.

## Consequences

- New metrics can be backfilled by replaying stored events.
- A projection bug is fixable: drop the table, rebuild, no data lost.
- Redeliveries are free, so the endpoint can be retried safely.
- `events` grows without bound and is the largest table by far. It will need a
  retention policy or partitioning by `occurred_at`; neither is implemented.
- Projections can lag or drift from `events` after a failure. There is no
  reconciliation job yet, so drift is currently silent — the replay tooling
  that fixes this is not built.
