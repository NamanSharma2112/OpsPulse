-- OpsPulse core schema: identity, tenancy, and the raw event log.
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users ---------------------------------------------------------------------
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         text NOT NULL,
    name          text NOT NULL,
    password_hash text,
    github_id     bigint,
    github_login  text,
    avatar_url    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    -- An account authenticates by password, by GitHub, or by both.
    CONSTRAINT users_has_credential CHECK (password_hash IS NOT NULL OR github_id IS NOT NULL)
);

CREATE UNIQUE INDEX users_email_key ON users (lower(email));
CREATE UNIQUE INDEX users_github_id_key ON users (github_id) WHERE github_id IS NOT NULL;

-- Organisations -------------------------------------------------------------
CREATE TABLE orgs (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name         text NOT NULL,
    slug         text NOT NULL UNIQUE,
    github_login text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE org_members (
    org_id     uuid NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role       text NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, user_id)
);

CREATE INDEX org_members_user_id_idx ON org_members (user_id);

-- Projects ------------------------------------------------------------------
CREATE TABLE projects (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         uuid NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    name           text NOT NULL,
    slug           text NOT NULL,
    repo_owner     text NOT NULL,
    repo_name      text NOT NULL,
    default_branch text NOT NULL DEFAULT 'main',
    -- Shared secret GitHub signs this repository's deliveries with.
    webhook_secret text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    UNIQUE (org_id, slug)
);

-- One repository is watched by exactly one project, so an incoming delivery
-- resolves unambiguously.
CREATE UNIQUE INDEX projects_repo_key ON projects (lower(repo_owner), lower(repo_name));
CREATE INDEX projects_org_id_idx ON projects (org_id);

-- Events: the append-only log every projection is derived from -------------
CREATE TABLE events (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    -- X-GitHub-Delivery, unique per delivery: makes ingest idempotent.
    delivery_id text NOT NULL UNIQUE,
    type        text NOT NULL,
    action      text,
    actor       text,
    payload     jsonb NOT NULL,
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX events_project_occurred_idx ON events (project_id, occurred_at DESC);
CREATE INDEX events_type_idx ON events (project_id, type);

-- Metrics: derived time series ---------------------------------------------
CREATE TABLE metrics (
    id          bigserial PRIMARY KEY,
    project_id  uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    name        text NOT NULL,
    value       double precision NOT NULL,
    labels      jsonb NOT NULL DEFAULT '{}'::jsonb,
    recorded_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX metrics_project_name_time_idx ON metrics (project_id, name, recorded_at DESC);

COMMIT;
