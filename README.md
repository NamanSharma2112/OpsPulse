# OpsPulse

Delivery and reliability signals for your GitHub repositories. OpsPulse
ingests webhooks, keeps the raw events, and renders what a team actually
wants on a Monday morning: is it healthy, what shipped, what is waiting for
review, and what is broken.

```
GitHub webhooks ──▶ Go API ──▶ PostgreSQL ──▶ Next.js dashboard
                 (auth,       (users, orgs,    (health, deployments,
                  projects,    projects,        PRs, incidents)
                  webhooks)    events, metrics)
```

## Layout

```
opspulse/
├── apps/web/          Next.js app — marketing site at /, dashboard at /dashboard
├── services/api/      Go API: auth, projects, webhook ingest, reads
├── packages/types/    TypeScript contracts shared with the dashboard
├── infra/
│   ├── docker/        Dockerfiles for the API and the dashboard
│   └── migrations/    Plain SQL, applied in filename order
├── docs/
│   ├── architecture/  How it fits together, plus the API reference
│   └── decisions/     Architecture decision records
└── docker-compose.yml
```

## Requirements

Docker Desktop is all you need to run the stack. To work on the code outside
containers you also want Go 1.24+ and Node 20+.

## Getting started

Docker is the supported path on every platform, Windows included.

```bash
cp .env.example .env     # PowerShell: copy .env.example .env
docker compose up --build
```

Postgres applies everything in `infra/migrations` the first time its volume
is created. The API is on `:8080`; the web app is on `:3000`, serving the
landing page at `/` and the dashboard at `/dashboard`.

### Sign in with GitHub

Create an OAuth app at **GitHub → Settings → Developer settings → OAuth Apps**:

- **Homepage URL** — `http://localhost:3000`
- **Authorization callback URL** — `http://localhost:8080/v1/auth/github/callback`

Put the client id and secret in `.env`, along with an encryption key for the
access tokens:

```bash
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...
TOKEN_ENCRYPTION_KEY=$(openssl rand -hex 32)
```

Then open http://localhost:3000/signin and continue with GitHub. From there:
pick a repository, and OpsPulse installs the webhook itself.

> **Webhooks need a public address.** GitHub cannot reach `localhost`, so for
> local development run a tunnel and set `PUBLIC_API_URL` to its https URL:
>
> ```bash
> cloudflared tunnel --url http://localhost:8080   # or: ngrok http 8080
> ```
>
> Without it the repository still connects, but no deliveries arrive.

### Without GitHub

Password sign-in still works, and the webhook is configured by hand:

```bash
TOKEN=$(curl -sS -X POST localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"You","password":"a-long-password"}' \
  | jq -r .token)

ORG=$(curl -sS -X POST localhost:8080/v1/organizations \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Acme"}' | jq -r .id)

PROJECT=$(curl -sS -X POST localhost:8080/v1/projects \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"organization_id\":\"$ORG\",\"name\":\"OpsPulse\"}" | jq -r .id)

curl -sS -X POST localhost:8080/v1/projects/$PROJECT/repositories \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"repo":"acme/opspulse","default_branch":"main"}'
```

The response includes `webhook_secret` **once**. In the repository's
*Settings → Webhooks*, add:

- **Payload URL** — `https://<your-host>/v1/webhooks/github`
- **Content type** — `application/json`
- **Secret** — the `webhook_secret` from above
- **Events** — pull requests, deployments, deployment statuses, workflow
  runs, issues

Finally point the dashboard at the API by setting `OPSPULSE_API_TOKEN` (the
token above) in `.env`, and restart it.

## Running without Docker

This path needs `psql` and `make` on your PATH, so it assumes macOS, Linux or
WSL. On Windows use Docker above, or run these inside WSL.

```bash
createdb opspulse
make migrate-up                   # DATABASE_URL overridable
make api                          # API on :8080
npm install && make web           # dashboard on :3000
```

Without `make`, the same steps directly:

```bash
psql "$DATABASE_URL" -f infra/migrations/0001_init.up.sql
psql "$DATABASE_URL" -f infra/migrations/0002_projections.up.sql
cd services/api && go run ./cmd/api
npm install && npm run dev
```

## Developing

```bash
make test        # Go test suite
make lint        # go vet + tsc --noEmit
make fmt         # gofmt
make build       # compile both applications
make reset       # drop the stack and its database volume
```

`make help` lists every target.

## Configuration

The API reads its configuration from the environment:

| Variable | Default | Notes |
| --- | --- | --- |
| `OPSPULSE_ENV` | `development` | `production` enables strict checks |
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_URL` | local Postgres | libpq connection string |
| `JWT_SECRET` | dev placeholder | **Required** in production |
| `JWT_TTL` | `24h` | Session lifetime |
| `GITHUB_CLIENT_ID` / `_SECRET` | — | OAuth app; enables sign-in and webhook install |
| `GITHUB_REDIRECT_URL` | `…:8080/v1/auth/github/callback` | Must match the OAuth app |
| `TOKEN_ENCRYPTION_KEY` | dev placeholder | Seals GitHub tokens; **required** in production |
| `APP_URL` | `http://localhost:3000` | Where sign-in returns the browser |
| `PUBLIC_API_URL` | `http://localhost:8080` | Address GitHub delivers webhooks to |
| `SECURE_COOKIES` | `false` | Must be `true` in production |
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `SHUTDOWN_TIMEOUT` | `15s` | Grace period for in-flight requests |

The dashboard reads `OPSPULSE_API_URL`, `OPSPULSE_API_TOKEN` and the optional
`OPSPULSE_PROJECT_ID`. See `apps/web/.env.example`.

## Design

The marketing surface follows `DESIGN.md` at the repository root — an
Intercom-derived editorial system: a cream canvas (`#f5f1ec`) rather than
white, white cards lifting off it with hairline borders instead of shadows,
charcoal as the system primary, and the accent orange reserved for AI
surfaces. Tokens live in `apps/web/app/globals.css`; the marketing styles in
`apps/web/app/(marketing)/marketing.css` reference them and nothing else.

Saans is proprietary, so the build substitutes Inter (weight 500 for display)
and JetBrains Mono, both self-hosted through `next/font`.

The signed-in dashboard keeps its own dark theme, scoped to `.dash` in
`apps/web/app/(app)/dashboard.css`.

## Documentation

- [Architecture overview](docs/architecture/overview.md) — the write path, the
  read path, layering, and what is deliberately not built
- [API reference](docs/architecture/api.md) — every endpoint
- [Decision records](docs/decisions/) — why things are the way they are
- [Migrations](infra/migrations/README.md) — schema conventions
- [DESIGN.md](DESIGN.md) — the design system the marketing site implements

## Data model

```
users ── organization_members ── organizations
                                      │
                                   projects
                                      │
                         ┌────────────┼────────────┐
                    repositories    events      projections
                                   (raw, append-only)
                                                 deployments
                                                 pull_requests
                                                 incidents
                                                 metrics
```

`events` is deliberately generic — one table holding every delivery verbatim,
rather than a table per event type. Projections are derived from it and can be
rebuilt by replaying it. See
[decision 0002](docs/decisions/0002-events-as-source-of-truth.md) and
[decision 0004](docs/decisions/0004-repositories-as-their-own-table.md).

## Status

**M1 — working GitHub integration: complete, and driveable from the browser.**
Sign in with GitHub → pick a repository → OpsPulse installs the webhook →
GitHub delivers → signature verified → stored in Postgres → visible on the
dashboard.

Verified end to end against a real Postgres and a stubbed GitHub: the OAuth
round trip, the `state` check, the httpOnly session cookie, webhook
installation, acceptance of a delivery signed with the installed secret,
rejection of a wrong signature, replay suppression, and encryption of the
stored access token.

The landing page's hero demo runs on fixed sample scenarios, not live data —
it illustrates the pipeline rather than reading from it. The correlation and
AI-explanation stages it depicts are the product direction; the API today
records events, projections and metric samples, and the statistical and
explanation stages are not implemented yet.

Not built: token refresh and revocation (a grant revoked on GitHub leaves a
stale row that fails on next use), a GitHub App to replace the OAuth app and
scope access per repository rather than per user, backfill from the GitHub
REST API, and a retention policy for `events`.
