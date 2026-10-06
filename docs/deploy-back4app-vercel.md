# Deploying OpsPulse: API on Back4App, dashboard on Vercel

Three pieces, in this order. Each needs the address of the one before it,
which is why the order matters.

```
   Supabase            Back4App                 Vercel
  (Postgres)      →   Go API container    →   Next.js dashboard
                       reads DATABASE_URL      proxies /v1/* to the API
                                                      ↑
                                            GitHub webhooks land here
```

The dashboard is the public face of the whole product. GitHub talks to it, the
browser talks to it, and it forwards `/v1/*` to the API. That is deliberate:
the session cookie is host-only, so serving the API under the dashboard's
origin is what keeps you signed in. **Never point GitHub or the browser at the
Back4App URL directly.**

---

## 1. Postgres on Supabase

Create a project, then open **Connect**. Ignore the App Frameworks, Mobile
Frameworks and ORMs tabs — they all hand you a Supabase client library, and
the Go API does not use one. It speaks Postgres directly with `pgx`.

Scroll to **Connection String** and take the **Session pooler** one:

```
postgresql://postgres.[PROJECT-REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:5432/postgres?sslmode=require
```

Session pooler, not the other two, for concrete reasons:

- **Transaction pooler (6543) breaks two things.** `migrate.go` takes a
  session-scoped advisory lock around the migration run and releases it
  afterwards; in transaction mode the release can land on a different backend
  than the acquire, so migrations hang or the lock leaks. Separately, `pgx`
  uses prepared statements by default, which transaction mode does not
  support — you get `prepared statement "stmt_1" already exists` on random
  requests.
- **Direct connection is IPv6-only** for new Supabase projects unless you have
  the IPv4 add-on, and Back4App cannot reach it. The failure looks like
  `dial tcp: connect: network is unreachable`, which reads like a credentials
  problem and is not.

You do not need to run anything in Supabase's SQL editor. The API creates
every table itself on first boot, under an advisory lock so two instances
starting together cannot race.

## 2. API on Back4App Containers

Push this repository to GitHub, then in Back4App: **Create new app →
Containers as a Service**, and pick the repository.

### Settings

| Field | Value |
| --- | --- |
| **Root directory** | `/` — leave it at the default |
| **Branch** | whichever you deploy from |
| **Auto deploy** | on, if you want a push to redeploy |

**Root directory must be `/`.** Back4App treats that one directory as *both*
the Docker build context and the place it looks for a `Dockerfile`. The API
image copies from `services/api/` and `infra/migrations/`, so its context has
to be the repository root — which is exactly why the API's `Dockerfile` lives
at the root rather than under `infra/docker/` with the web one. Pointing the
root at a subdirectory makes every `COPY` fail.

Leave the port alone. Back4App injects `PORT` and the API binds to it.

### Environment variables

| Variable | Value |
| --- | --- |
| `OPSPULSE_ENV` | `production` |
| `DATABASE_URL` | the Session pooler string from step 1 |
| `JWT_SECRET` | `openssl rand -hex 32` |
| `TOKEN_ENCRYPTION_KEY` | `openssl rand -hex 32` — a **different** one |
| `SECURE_COOKIES` | `true` |
| `APP_URL` | your Vercel URL, e.g. `https://opspulse.vercel.app` |
| `PUBLIC_API_URL` | the same Vercel URL |
| `CORS_ORIGINS` | the same Vercel URL |
| `GITHUB_CLIENT_ID` | from the OAuth app in step 4 |
| `GITHUB_CLIENT_SECRET` | from the OAuth app in step 4 |
| `GITHUB_REDIRECT_URL` | `https://YOUR-VERCEL-URL/v1/auth/github/callback` |

Optional:

| Variable | Default | Effect |
| --- | --- | --- |
| `ANOMALY_INTERVAL` | `15m` | How often M5 sweeps. `0` disables it. |
| `HEALTH_WEIGHT_RELIABILITY` | `0.40` | M3 pillar weights. See `docs/decisions/0006`. |
| `HEALTH_WEIGHT_DELIVERY` | `0.35` | |
| `HEALTH_WEIGHT_PERFORMANCE` | `0.25` | |
| `LOG_LEVEL` | `info` | `debug` while you are setting up. |

You will not have the Vercel URL on the first pass. Deploy with placeholders,
finish step 3, then come back and fix the four URL variables.

The API **refuses to start** in production without `JWT_SECRET`,
`TOKEN_ENCRYPTION_KEY` and `SECURE_COOKIES=true`. That is on purpose — those
three are the ones that are quietly catastrophic to forget.

### Confirm it came up

The logs should show, in order:

```
applied migration version=0001_init
applied migration version=0002_projections
applied migration version=0003_repositories
applied migration version=0004_github_oauth
applied migration version=0005_analytics
applied migration version=0006_anomalies
api listening addr=:XXXX
anomaly detection started interval=15m0s
```

Then `https://YOUR-BACK4APP-URL/healthz` returns `{"status":"ok"}`.

Six migrations and then `api listening` is the shape of a healthy boot. If the
migrations appear but `api listening` does not, the database connected and
something else failed — read the line after the last migration.

## 3. Dashboard on Vercel

1. **New Project**, import the same repository.
2. **Root Directory** → `apps/web`. This is the step people miss in a
   monorepo; Vercel otherwise tries to build the repository root.
3. Framework preset: Next.js. Leave build and output settings alone.

| Variable | Value |
| --- | --- |
| `OPSPULSE_API_URL` | your Back4App URL, e.g. `https://opspulse-api.b4a.run` |
| `SECURE_COOKIES` | `true` |

`OPSPULSE_API_URL` is read per request, not baked at build time, so changing
it takes effect on the next request rather than needing a rebuild.

Deploy, then go back to Back4App and set `APP_URL`, `PUBLIC_API_URL`,
`CORS_ORIGINS` and `GITHUB_REDIRECT_URL` to the Vercel URL you now have.

## 4. GitHub OAuth app

At <https://github.com/settings/developers> → **New OAuth App**:

- **Homepage URL**: `https://YOUR-VERCEL-URL`
- **Authorization callback URL**:
  `https://YOUR-VERCEL-URL/v1/auth/github/callback`

The callback must match `GITHUB_REDIRECT_URL` character for character, and it
must point at **Vercel, not Back4App**. Copy the client ID and secret into
Back4App and redeploy.

## 5. Check it end to end

1. Open the Vercel URL and sign in with GitHub.
2. Connect a repository. OpsPulse installs the webhook itself with your token.
3. Push a commit. It should appear under **Events** within a second or two.
4. **Analytics** fills in as data arrives. Some panels stay empty at first by
   design: cycle time needs a merged pull request, and anomaly detection needs
   about nine days of history before it will judge anything.

---

## Things that will bite you

**The OAuth callback points at Back4App.** Sign-in appears to work, then fails
on the way back, because the state cookie was set on one origin and read from
another. It must be the Vercel URL.

**`SECURE_COOKIES` is false in production.** The browser silently drops the
session cookie over HTTPS and you land back on sign-in with no error at all.

**Root directory set to a subfolder on Back4App.** Every `COPY` in the
Dockerfile fails because the build context moved. Keep it `/`.

**Back4App's free tier sleeps.** The first request after idle is slow, and a
webhook delivered while the container is asleep can time out. GitHub retries,
so an event is rarely lost outright, but use a paid tier for anything real.
Anomaly sweeps are also skipped while the container is down.

**Blank dashboard panels are not always missing data.** The dashboard renders
an empty state rather than an error when the API is unreachable. Check
`OPSPULSE_API_URL` on Vercel and `/healthz` on Back4App before concluding
nothing was recorded.
