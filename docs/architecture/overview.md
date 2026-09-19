# OpsPulse architecture

OpsPulse turns a repository's GitHub activity into delivery and reliability
signals: is the service healthy, what shipped, what is waiting for review, and
what is currently broken.

## Shape

```
        GitHub (webhooks + REST API)
                    │
                    ▼
        ┌───────────────────────┐
        │        Go API         │
        │  auth · projects ·    │
        │  webhooks · ingest    │
        └───────────┬───────────┘
                    ▼
        ┌───────────────────────┐
        │      PostgreSQL       │
        │  users · orgs ·       │
        │  projects · events ·  │
        │  metrics              │
        └───────────┬───────────┘
                    ▼
        ┌───────────────────────┐
        │   Next.js dashboard   │
        │  health · deployments │
        │  PRs · incidents      │
        └───────────────────────┘
```

Both applications live in one repository: the Go service under
`services/api`, the dashboard under `apps/web`, shared TypeScript contracts
under `packages/types`.

## The write path

1. GitHub posts a delivery to `POST /v1/webhooks/github`.
2. The handler reads `X-GitHub-Event` and `X-GitHub-Delivery`, parses just
   enough of the body to find `owner/repo`, and looks up the project.
3. The delivery's `X-Hub-Signature-256` is checked against that project's
   webhook secret. An unsigned or mis-signed delivery is rejected with 401
   before anything is written.
4. The raw payload is inserted into `events`. `delivery_id` is unique, so a
   replayed delivery is recognised and skipped — ingest is idempotent.
5. The delivery is applied to the projections and to `metrics`.

A projection failure is logged but does not fail the request: the event is
already durable, and returning an error would make GitHub retry a delivery
that was accepted.

## Events and projections

`events` is append-only and is the source of truth. Everything else is derived
and can be dropped and rebuilt by replaying it.

| Table | Built from | Feeds |
| --- | --- | --- |
| `deployments` | `deployment`, `deployment_status` | Deployments page, health tiles |
| `pull_requests` | `pull_request` | Pull requests page |
| `incidents` | `workflow_run`, `issues` | Incidents page, health status |
| `metrics` | all of the above | Success rate, counts over a window |

### What becomes an incident

- A `workflow_run` that completes with `failure` or `timed_out` **on the
  project's default branch**. A failing feature branch is the author's
  problem; a failing default branch is everyone's. A later successful run of
  the same workflow resolves it.
- An issue labelled `incident` or `outage`. Severity comes from a
  `severity:*` label or the `sev1`/`sev2`/`sev3` shorthand, defaulting to
  minor. Closing the issue resolves the incident.

## The read path

The dashboard renders server-side. Each page resolves the active project, then
reads the projection endpoints with a bearer token. Reads are `no-store`:
operational data is stale almost immediately, so there is nothing to cache.

`GET /v1/projects/{id}/health` grades a project over a rolling 7-day window:
any open critical incident is `critical`, any other open incident or recent
failed deploy is `degraded`, otherwise `healthy`.

## Authentication and tenancy

Two separate mechanisms, deliberately:

- **People** get a signed HS256 token from `/v1/auth/register` or
  `/v1/auth/login`, sent as `Authorization: Bearer`. Passwords are stored as
  salted PBKDF2-HMAC-SHA256, with the parameters encoded in the hash so they
  can be raised later without invalidating existing credentials.
- **GitHub** authenticates per delivery with an HMAC signature over the body.
  No user token is involved in ingest.

Every project belongs to an org, and a user reaches a project only through
`org_members`. A request for a project in an org the caller does not belong to
returns 404 rather than 403, so the API does not confirm that a project
exists to someone who cannot see it.

## Layering

```
cmd/api                 process wiring, signals, graceful shutdown
  internal/httpapi      transport: routing, middleware, JSON, status codes
  internal/auth         sessions, password hashing
  internal/orgs         tenancy and roles
  internal/projects     watched repositories, webhook secrets
  internal/ingest       deliveries → events, projections, metrics
  internal/github       signature verification, payload decoding
  internal/store        persistence interfaces
  internal/store/postgres   the implementation
  internal/domain       types and sentinel errors, depended on by everything
```

Dependencies point inwards. `domain` imports nothing of ours; `httpapi`
imports services, never `store/postgres` directly. Services take the `store`
interfaces, so they can be tested against fakes.

Errors surface as sentinels from `internal/domain` (`ErrNotFound`,
`ErrForbidden`, …) and one function in the transport layer maps them onto
status codes. Handlers never mention pgx, and driver errors never leak into
responses.

## Deliberate choices

- **No router dependency.** `net/http`'s `ServeMux` has matched methods and
  path parameters since Go 1.22.
- **No ORM.** Queries are hand-written SQL in one package per table.
- **Enums as `text` + `CHECK`.** Adding a value is an ordinary migration
  rather than a lock on a Postgres enum type.
- **Timestamps are `timestamptz`, always UTC.**

## Not built yet

- GitHub OAuth sign-in (config is read; the callback is not implemented).
- Interactive sign-in in the dashboard — it currently reads a token from the
  environment.
- A worker for polling the GitHub REST API to backfill history.
- Metric rollups; the health endpoint counts raw samples per request.
