# Migrations

Plain SQL, applied in filename order. Every migration ships an `.up.sql` and a
matching `.down.sql`, and each runs inside a single transaction.

| File | Contents |
| --- | --- |
| `0001_init` | `users`, `orgs`, `org_members`, `projects`, `events`, `metrics` |
| `0002_projections` | `deployments`, `pull_requests`, `incidents` |
| `0003_repositories` | `repositories`; renames `orgs`/`org_members` to `organizations`/`organization_members`; adds `events.source` and `repository_id` |

## Applying them

The API applies pending `.up.sql` files on startup, in filename order, inside
a Postgres advisory lock so two instances starting together cannot race. What
it applied is recorded in `schema_migrations`. This is the same path locally
and in production, so there is one mechanism to reason about.

Set `RUN_MIGRATIONS=false` to turn it off, and `MIGRATIONS_DIR` to point
somewhere other than the default.

Against an existing database, apply one by hand:

```bash
psql "$DATABASE_URL" -f infra/migrations/0001_init.up.sql
```

Or use the Make targets from the repository root:

```bash
make migrate-up     # apply every .up.sql in order
make migrate-down   # roll back every .down.sql in reverse order
```

## Conventions

- Identifiers are `uuid` with `gen_random_uuid()` defaults, except `metrics`,
  which uses `bigserial` because it is append-only and high volume.
- Timestamps are `timestamptz`, always stored in UTC.
- Enumerations are `text` plus a `CHECK` constraint rather than Postgres enum
  types, so adding a value is an ordinary migration rather than a lock.
- `events` is append-only and is the source of truth. Tables in `0002` are
  projections: they can be dropped and rebuilt by replaying `events`.
