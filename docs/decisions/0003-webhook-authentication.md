# 3. Authenticate webhooks per project with HMAC signatures

- **Status:** accepted
- **Date:** 2026-09-19

## Context

`POST /v1/webhooks/github` is a public, unauthenticated-by-token endpoint:
GitHub cannot present a user session. Something has to stop anyone who knows
the URL from injecting fake deployments and incidents.

GitHub signs each delivery with HMAC-SHA256 over the raw body, using a secret
configured on the webhook, and sends it as `X-Hub-Signature-256`.

The alternatives were a single shared secret for the whole installation, or
allow-listing GitHub's published IP ranges. A shared secret means one leak
compromises every project and rotation is all-or-nothing. IP allow-lists
authenticate the network, not the payload, and GitHub's ranges change.

## Decision

Each project gets its own webhook secret, 32 random bytes generated at
creation and stored in `projects.webhook_secret`. It is returned exactly once,
in the `POST /v1/projects` response, so it can be pasted into the repository's
webhook settings; it is never served again — `Project.WebhookSecret` is
`json:"-"`.

The handler resolves the project from the payload's repository first, then
verifies the signature against *that project's* secret. Comparison uses
`hmac.Equal`, never `==`. A delivery with a missing, malformed or wrong
signature is rejected with 401 before any write.

## Consequences

- A leaked secret affects one project and is rotated by re-creating it.
- Bodies must be read and buffered before parsing, since the signature covers
  raw bytes. The endpoint caps a delivery at GitHub's own 25 MB limit.
- Resolving the project happens before authentication, so an attacker can
  learn whether a repository is registered (404 versus 401). This is accepted:
  the repository names are usually public anyway.
- Secrets are stored in plaintext because HMAC verification needs the original
  value. They are therefore as sensitive as the database itself; encrypting
  them at rest with a KMS-held key is a possible follow-up.
