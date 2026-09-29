# Deploying OpsPulse

Three pieces: Postgres, the Go API, and the Next.js dashboard.

**The dashboard is the only public address.** It proxies `/v1/*` to the API,
so the whole product is served from one origin. That matters for two reasons:
the session cookie is host-only, so a second hostname would mean the
dashboard never receives it; and GitHub needs exactly one webhook URL. The
API does not need to be reachable from a browser at all.

```
        browser                     GitHub
           │                           │
           ▼                           ▼
   ┌────────────────────────────────────────┐
   │        dashboard  (public origin)      │
   │   pages  ·  /v1/*  ──proxy──▶  API     │
   └────────────────────────────────┬───────┘
                                    ▼
                                Postgres
```

The API applies its own migrations at startup, so a deployment is: point it
at an empty database and start it.

## Render (blueprint)

`render.yaml` in the repository root describes all three services.

1. **Create a GitHub OAuth app** — GitHub → Settings → Developer settings →
   OAuth Apps → New. Leave the callback blank for now; you fill it in once
   you know the dashboard's URL.
2. **Render → New → Blueprint**, pick this repository. It creates the
   database and both services.
3. Render will ask for the values marked `sync: false`:
   - `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` from step 1.
   - `TOKEN_ENCRYPTION_KEY` — generate with `openssl rand -hex 32`.
   - `GITHUB_REDIRECT_URL` — `https://<dashboard>.onrender.com/v1/auth/github/callback`.
     You will know the host after the first deploy; set it then and redeploy.
4. **Set the OAuth app's callback URL** to the same value as
   `GITHUB_REDIRECT_URL`. They must match exactly or GitHub refuses the
   sign-in.
5. Open the dashboard URL and sign in with GitHub, then connect a repository.
   OpsPulse installs the webhook itself.

On Render's free plan services sleep when idle. A sleeping service misses
webhook deliveries — GitHub retries, but not forever. Use a paid instance for
anything you care about.

## Anywhere else

Any host that runs two containers and a Postgres works the same way. Build
`infra/docker/api.Dockerfile` and `infra/docker/web.Dockerfile`, both with
the repository root as the build context.

**API** — needs to reach Postgres; does not need to be public.

| Variable | Value |
| --- | --- |
| `OPSPULSE_ENV` | `production` |
| `DATABASE_URL` | Postgres connection string |
| `JWT_SECRET` | `openssl rand -hex 32` |
| `TOKEN_ENCRYPTION_KEY` | `openssl rand -hex 32`, exactly 64 hex characters |
| `SECURE_COOKIES` | `true` (the API refuses to start in production without it) |
| `APP_URL` | the dashboard's public URL |
| `PUBLIC_API_URL` | the dashboard's public URL — what GitHub delivers to |
| `GITHUB_REDIRECT_URL` | `<dashboard URL>/v1/auth/github/callback` |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | from the OAuth app |
| `CORS_ORIGINS` | the dashboard's public URL |

**Dashboard** — the public one.

| Variable | Value |
| --- | --- |
| `OPSPULSE_API_URL` | where the API is, reachable from the dashboard's network |
| `NEXT_OUTPUT` | `standalone` (the Dockerfile sets this) |

**Postgres** — any managed instance. Nothing to prepare: the API creates
the schema on first boot and records what it applied in `schema_migrations`.

## Keys, and what losing them costs

Both are generated once and kept.

- `JWT_SECRET` signs sessions. Changing it signs everyone out.
- `TOKEN_ENCRYPTION_KEY` encrypts stored GitHub access tokens. Changing it
  makes existing tokens unreadable — users sign in again, and until they do,
  OpsPulse cannot install webhooks on their behalf. It must be exactly 32
  bytes as 64 hex characters.

## Checking a deployment

```bash
curl https://<dashboard>/healthz              # the API is running
curl https://<dashboard>/readyz               # it can reach Postgres
curl https://<dashboard>/v1/auth/providers    # {"github":true,"password":true}
```

`"github": false` means the OAuth credentials did not reach the API — sign-in
with GitHub and automatic webhook installation will be unavailable, and the
API logs a warning saying so at startup.

## Running it locally instead

```bash
cp .env.example .env
docker compose up --build
```

The dashboard is on `http://localhost:3000` and is the origin to use.
Webhooks need a public address, so for local development run a tunnel and
point `PUBLIC_API_URL` and `GITHUB_REDIRECT_URL` at it:

```bash
cloudflared tunnel --url http://localhost:3000
```

Without that, sign-in and the dashboard work, but no deliveries arrive.

## Schema changes after the first deploy

Add a numbered pair to `infra/migrations` (`0005_thing.up.sql` and
`.down.sql`). The API applies pending `.up.sql` files in filename order on
its next start, inside an advisory lock so two instances cannot race. The
`.down.sql` files are for rolling back by hand; nothing applies them
automatically.
