# API reference

Base URL in development: `http://localhost:8080`.

All request and response bodies are JSON. Errors share one shape:

```json
{ "error": { "message": "not found" } }
```

Authenticated endpoints expect `Authorization: Bearer <token>`.

## Health

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness. Answers while the process is up. |
| `GET` | `/readyz` | Readiness. Also pings Postgres; 503 when unreachable. |

## Auth

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `POST` | `/v1/auth/register` | — | Create an account, return a session |
| `POST` | `/v1/auth/login` | — | Exchange credentials for a session |
| `GET` | `/v1/auth/me` | bearer | The current user |

```bash
curl -X POST localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"You","password":"a-long-password"}'
```

Returns `{ "token": "...", "expires_at": "...", "user": { ... } }`. The token
is valid for `JWT_TTL` (24h by default).

## Organisations

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/v1/orgs` | Orgs the caller belongs to |
| `POST` | `/v1/orgs` | Create an org; the caller becomes its owner |
| `GET` | `/v1/orgs/{orgID}` | One org |
| `POST` | `/v1/orgs/{orgID}/members` | Add a member (owner/admin only) |
| `GET` | `/v1/orgs/{orgID}/projects` | Projects in the org |

## Projects

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/v1/projects` | Every project the caller can see |
| `POST` | `/v1/projects` | Start watching a repository |
| `GET` | `/v1/projects/{projectID}` | One project |

```bash
curl -X POST localhost:8080/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"org_id":"<uuid>","name":"OpsPulse","repo":"acme/opspulse","default_branch":"main"}'
```

The response carries `webhook_secret` **once**. Configure it on the
repository's webhook; it is never returned again.

## Dashboard reads

All require bearer auth and membership of the project's org.

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/v1/projects/{id}/health` | Rolling 7-day summary |
| `GET` | `/v1/projects/{id}/deployments` | `?limit=` (default 25, max 200) |
| `GET` | `/v1/projects/{id}/pull-requests` | `?state=open\|closed\|merged`, `?limit=` |
| `GET` | `/v1/projects/{id}/incidents` | `?status=open\|acknowledged\|resolved`, `?limit=` |
| `GET` | `/v1/projects/{id}/events` | Raw deliveries, `?limit=` (default 50, max 500) |

A project in an org the caller does not belong to returns **404**, not 403.

## Webhook ingest

| Method | Path | Auth |
| --- | --- | --- |
| `POST` | `/v1/webhooks/github` | `X-Hub-Signature-256` over the raw body |

Required headers: `X-GitHub-Event`, `X-GitHub-Delivery`,
`X-Hub-Signature-256`. A `ping` event returns `{"status":"pong"}`.

Responses:

| Status | Meaning |
| --- | --- |
| `202` | Accepted. Body reports `duplicate: true` for a replayed delivery. |
| `400` | Missing headers, unparseable body, or no repository in the payload. |
| `401` | Missing or invalid signature. |
| `404` | No project registered for that repository. |

Events understood today: `deployment`, `deployment_status`, `pull_request`,
`workflow_run`, `issues`. Anything else is stored in `events` but projects
nothing.

## Status codes

| Code | When |
| --- | --- |
| `400` | Invalid input |
| `401` | Missing/invalid credentials or signature |
| `403` | Authenticated but not permitted |
| `404` | Missing, or hidden from the caller |
| `409` | Already exists |
| `503` | Readiness probe failed |
