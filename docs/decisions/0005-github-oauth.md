# 5. Sign in with GitHub, and install webhooks with the same token

- **Status:** accepted
- **Date:** 2026-09-19

## Context

Connecting a repository took six manual steps: create an account, an
organization and a project over the API, read a webhook secret out of a JSON
response, open the repository's settings on GitHub, and paste it in. Every one
of them is a place to stop.

OpsPulse already needs to read a repository's activity. The same OAuth grant
that identifies a user can also install the webhook, which removes the paste
step entirely and the three API calls before it.

## Decision

GitHub OAuth is the primary sign-in. `GET /v1/auth/github` redirects to
GitHub; `GET /v1/auth/github/callback` exchanges the code, upserts the user by
`github_id`, and returns them to the dashboard with a session.

Requested scopes are `read:user`, `user:email`, `repo` and `admin:repo_hook`.
`admin:repo_hook` is what makes automatic installation possible; the sign-in
page lists all four with a sentence each, because a consent screen a user
cannot interpret is not consent.

**CSRF.** The `state` parameter is 32 random bytes, stored in a short-lived
httpOnly cookie and compared on the way back. A callback without a matching
cookie is refused.

**The session is an httpOnly cookie**, not a token in `localStorage`. The
dashboard reads it server-side and forwards it to the API as a bearer token,
so browser script never holds a credential. The API accepts either, which
keeps the existing header-based clients working.

**The access token is encrypted at rest** with AES-256-GCM, the key held in
the environment rather than the database. A token grants access to someone's
repositories, so a database dump on its own should not yield working
credentials. This is why `TOKEN_ENCRYPTION_KEY` is required outside
development.

**Webhook installation is best effort.** The repository is connected first,
then the hook is installed. A failure — no admin rights, no public URL, a hook
already present — leaves the repository connected and reports the reason, with
the secret shown so it can still be added by hand.

Password sign-in stays. A self-hosted install without an OAuth app should not
be locked out, and `GET /v1/auth/providers` tells the dashboard which methods
exist so it can hide a button that cannot work.

## Consequences

- Connecting a repository is: sign in, pick it from a list, done. The list is
  filtered to repositories the user can administer, because those are the ones
  a webhook can be installed on.
- OpsPulse now holds a credential that can read private repositories. That is
  a meaningful escalation of what a breach costs, and the reason for
  encryption at rest and for requesting scopes per-user rather than a single
  service account.
- `PUBLIC_API_URL` must be reachable from GitHub. On a laptop that means a
  tunnel; `localhost` will install a hook that can never deliver.
- Tokens are not refreshed or revoked yet. A user who revokes the grant on
  GitHub leaves a stored token that fails on next use, and nothing prunes it.
- A GitHub App would scope access per-repository rather than per-user and
  avoid holding a user token at all. It is the better end state; an OAuth app
  is what gets the flow working now.
