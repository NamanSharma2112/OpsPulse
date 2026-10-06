# Deploying OpsPulse: API on Back4App, dashboard on Vercel

Three pieces, in this order. Each one needs the address of the one before it,
which is why the order matters.

```
   Postgres            Back4App                 Vercel
  (Neon/Supabase)  →   Go API container    →   Next.js dashboard
                        reads DATABASE_URL      proxies /v1/* to the API
                                                      ↑
                                            GitHub webhooks land here
```

The dashboard is the public face of the whole product. GitHub talks to it, the
browser talks to it, and it forwards `/v1/*` to the API. That is deliberate:
the session cookie is host-only, so serving the API under the dashboard's
origin is what keeps you signed in. **Do not point GitHub or the browser at
the Back4App URL directly.**

---

## 1. Postgres

Back4App Containers does not give you a Postgres, so bring one. Any managed
Postgres 14+ works; the free tiers of [Neon](https://neon.tech) and
[Supabase](https://supabase.com) are both fine to start.

Create a database and copy its connection string. It must include SSL:

```
postgres://USER:PASSWORD@HOST/DBNAME?sslmode=require
```

You do not need to run any migrations by hand. The API applies them on
startup, under an advisory lock so two instances booting together cannot race.

## 2. API on Back4App Containers

Push this repository to GitHub, then in Back4App:

1. **Create new app → Containers as a Service**, and pick the repository.
2. Set **Dockerfile path** to `infra/docker/api.Dockerfile`, and leave the
   build context at the repository root. The Dockerfile copies both
   `services/api/` and `infra/migrations/`, so it needs the root.
3. Leave the port alone. Back4App injects `PORT` and the API honours it.

Environment variables:

| Variable | Value |
| --- | --- |
| `OPSPULSE_ENV` | `production` |
| `DATABASE_URL` | the connection string from step 1 |
| `JWT_SECRET` | `openssl rand -hex 32` |
| `TOKEN_ENCRYPTION_KEY` | `openssl rand -hex 32` — a **different** one |
| `SECURE_COOKIES` | `true` |
| `APP_URL` | your Vercel URL, e.g. `https://opspulse.vercel.app` |
| `PUBLIC_API_URL` | the same Vercel URL |
| `CORS_ORIGINS` | the same Vercel URL |
| `GITHUB_CLIENT_ID` | from the OAuth app in step 4 |
| `GITHUB_CLIENT_SECRET` | from the OAuth app in step 4 |
| `GITHUB_REDIRECT_URL` | `https://YOUR-VERCEL-URL/v1/auth/github/callback` |

You will not have the Vercel URL yet on the first pass. Deploy with
placeholders, finish step 3, then come back and set the three URL variables
properly. The API refuses to start in production without `JWT_SECRET`,
`TOKEN_ENCRYPTION_KEY` and `SECURE_COOKIES=true`, which is intentional — those
are the three that are catastrophic to forget.

Check it came up: `https://YOUR-BACK4APP-URL/healthz` should return
`{"status":"ok"}`.

## 3. Dashboard on Vercel

1. **New Project**, import the same repository.
2. Set **Root Directory** to `apps/web`. This is the step people miss — the
   repository is a monorepo and Vercel will otherwise try to build the root.
3. Framework preset: Next.js. Leave the build and output settings alone.

Environment variables:

| Variable | Value |
| --- | --- |
| `OPSPULSE_API_URL` | your Back4App URL, e.g. `https://opspulse-api.b4a.run` |
| `SECURE_COOKIES` | `true` |

`OPSPULSE_API_URL` is read at request time, not at build time, so changing it
takes effect on the next request rather than needing a rebuild.

Deploy, then go back to Back4App and set `APP_URL`, `PUBLIC_API_URL`,
`CORS_ORIGINS` and `GITHUB_REDIRECT_URL` to the Vercel URL you now have.

## 4. GitHub OAuth app

At <https://github.com/settings/developers> → **New OAuth App**:

- **Homepage URL**: `https://YOUR-VERCEL-URL`
- **Authorization callback URL**:
  `https://YOUR-VERCEL-URL/v1/auth/github/callback`

The callback must match `GITHUB_REDIRECT_URL` character for character, and it
must point at **Vercel, not Back4App**. Copy the client ID and secret into the
Back4App environment and redeploy.

## 5. Check it end to end

1. Open the Vercel URL and sign in with GitHub.
2. Connect a repository. OpsPulse installs the webhook itself using your
   OAuth token.
3. Push a commit. Within a second or two it should appear under **Events**.
4. **Analytics** fills in as data arrives. Some panels stay empty at first by
   design — cycle time needs a merged pull request, and anomaly detection
   needs about a week of history before it will judge anything.

---

## Things that will bite you

**The callback URL points at Back4App.** Sign-in appears to work, then fails
on the way back, because the state cookie was set on one origin and read on
another. It must be the Vercel URL.

**`SECURE_COOKIES` is false in production.** The browser silently drops the
session cookie over HTTPS and you land back on the sign-in page with no error.
The API refuses to start without it, which is the point.

**Back4App free tier sleeps.** The first request after an idle period is slow,
and a GitHub webhook delivered while the container is asleep can time out.
GitHub retries, so an event is rarely lost outright, but for anything real use
a paid tier.

**Anomaly detection runs in the API process.** It sweeps every 15 minutes
(`ANOMALY_INTERVAL`). On a free tier that sleeps, sweeps are skipped while the
container is down. Set `ANOMALY_INTERVAL=0` to turn it off entirely.

**The dashboard shows an empty state, not an error, when the API is
unreachable.** If panels are blank, check `OPSPULSE_API_URL` on Vercel and
`/healthz` on Back4App before assuming the data is missing.
