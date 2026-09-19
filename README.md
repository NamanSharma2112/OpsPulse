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
├── apps/web/          Next.js dashboard (App Router, server components)
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

Go 1.24+, Node 20+, and either Docker or a local PostgreSQL 16.

## Getting started

```bash
cp .env.example .env
make up          # postgres + api + dashboard
```

Postgres applies everything in `infra/migrations` the first time its volume
is created. The API is on `:8080`, the dashboard on `:3000`.

Create an account and start watching a repository:

```bash
TOKEN=$(curl -sS -X POST localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"You","password":"a-long-password"}' \
  | jq -r .token)

ORG=$(curl -sS -X POST localhost:8080/v1/orgs \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Acme"}' | jq -r .id)

curl -sS -X POST localhost:8080/v1/projects \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"org_id\":\"$ORG\",\"name\":\"OpsPulse\",\"repo\":\"acme/opspulse\"}"
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

```bash
createdb opspulse
make migrate-up                   # DATABASE_URL overridable
make api                          # API on :8080
npm install && make web           # dashboard on :3000
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
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `SHUTDOWN_TIMEOUT` | `15s` | Grace period for in-flight requests |

The dashboard reads `OPSPULSE_API_URL`, `OPSPULSE_API_TOKEN` and the optional
`OPSPULSE_PROJECT_ID`. See `apps/web/.env.example`.

## Documentation

- [Architecture overview](docs/architecture/overview.md) — the write path, the
  read path, layering, and what is deliberately not built
- [API reference](docs/architecture/api.md) — every endpoint
- [Decision records](docs/decisions/) — why things are the way they are
- [Migrations](infra/migrations/README.md) — schema conventions

## Status

Early. The ingest pipeline, tenancy model, dashboard reads and schema are in
place and tested end to end. Not yet built: GitHub OAuth sign-in, interactive
sign-in in the dashboard, backfill from the GitHub REST API, and a retention
policy for the `events` table.
