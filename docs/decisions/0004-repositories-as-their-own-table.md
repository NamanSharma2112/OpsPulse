# 4. Give repositories their own table

- **Status:** accepted
- **Date:** 2026-09-19

## Context

The first schema folded the watched repository into `projects`:
`repo_owner`, `repo_name`, `default_branch` and `webhook_secret` were columns
on the project, with a unique index across owner and name. One project meant
exactly one repository.

That held for a single GitHub repo per project and nothing else. It could not
express a project spanning several services, it had no room for a second
provider, and it put the webhook secret — the credential that authenticates
ingest — on a row that is otherwise ordinary metadata.

## Decision

`repositories` is its own table: `project_id`, `provider`, `external_id`,
`name`, `default_branch`, `webhook_secret`. A project has many.

`external_id` is the provider's own identifier, stored as the lookup key. For
GitHub that is the lowercased `owner/name`, which is exactly what arrives in a
webhook payload, so resolving a delivery is one indexed read against
`UNIQUE (provider, external_id)`. GitHub treats repository names as
case-insensitive, so the key is lowercased on write and on lookup.

The webhook secret moves here, because it belongs to one webhook on one
repository.

`events` gains `source` and `repository_id`. `source` names the system the
event came from; `repository_id` names the origin within it. Without the
latter, an event loses its origin the moment a project watches more than one
repository — the same information loss this decision exists to avoid.

## Consequences

- A project can watch several repositories, which is what a service-per-repo
  team needs.
- A leaked webhook secret is contained to one repository and rotated by
  reconnecting it.
- A second provider is a new `provider` value and a new ingest handler, not a
  migration of this shape again.
- `POST /v1/projects` no longer takes a repo. Connecting one is a second call
  to `POST /v1/projects/{id}/repositories`, which is where the secret is
  issued — once.
- `repository_id` is nullable on `events`, `deployments` and `pull_requests`,
  because a future event source may not be a repository at all. Rows created
  by GitHub ingest always set it.
- More joins on read paths that want the repository name. Acceptable: the
  dashboard reads projections, which carry `project_id` directly.
