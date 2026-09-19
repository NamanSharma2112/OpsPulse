-- Read models the dashboard queries directly. Every row here is derived from
-- the events table and can be rebuilt by replaying it.
BEGIN;

CREATE TABLE deployments (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    -- GitHub deployment id, as text so other sources can be added later.
    external_id text NOT NULL,
    environment text NOT NULL DEFAULT 'production',
    ref         text NOT NULL DEFAULT '',
    sha         text NOT NULL DEFAULT '',
    status      text NOT NULL CHECK (status IN ('pending', 'running', 'success', 'failure', 'inactive')),
    actor       text,
    url         text,
    started_at  timestamptz NOT NULL,
    finished_at timestamptz,

    UNIQUE (project_id, external_id)
);

CREATE INDEX deployments_project_started_idx ON deployments (project_id, started_at DESC);
CREATE INDEX deployments_env_idx ON deployments (project_id, environment, started_at DESC);

CREATE TABLE pull_requests (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    number     integer NOT NULL,
    title      text NOT NULL,
    author     text NOT NULL DEFAULT '',
    state      text NOT NULL CHECK (state IN ('open', 'closed', 'merged')),
    draft      boolean NOT NULL DEFAULT false,
    url        text,
    opened_at  timestamptz NOT NULL,
    merged_at  timestamptz,
    closed_at  timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),

    UNIQUE (project_id, number)
);

CREATE INDEX pull_requests_project_state_idx ON pull_requests (project_id, state, updated_at DESC);

CREATE TABLE incidents (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    -- Stable key from the originating signal, e.g. "workflow_run:42".
    external_id text NOT NULL,
    title       text NOT NULL,
    severity    text NOT NULL CHECK (severity IN ('critical', 'major', 'minor')),
    status      text NOT NULL CHECK (status IN ('open', 'acknowledged', 'resolved')),
    source      text NOT NULL,
    url         text,
    opened_at   timestamptz NOT NULL,
    resolved_at timestamptz,

    UNIQUE (project_id, external_id),
    CONSTRAINT incidents_resolved_has_timestamp
        CHECK (status <> 'resolved' OR resolved_at IS NOT NULL)
);

CREATE INDEX incidents_project_status_idx ON incidents (project_id, status, opened_at DESC);

COMMIT;
