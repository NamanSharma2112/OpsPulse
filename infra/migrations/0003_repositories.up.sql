-- Aligns the schema with the agreed model: repositories become a first-class
-- table, and events record which provider and which repository they came from.
BEGIN;

-- Match the model's vocabulary. Indexes, constraints and foreign keys follow
-- a rename automatically.
ALTER TABLE orgs RENAME TO organizations;
ALTER TABLE org_members RENAME TO organization_members;
ALTER TABLE organization_members RENAME COLUMN org_id TO organization_id;
ALTER TABLE projects RENAME COLUMN org_id TO organization_id;

-- Repositories ------------------------------------------------------------
-- A project watches one or more repositories. Splitting them out of projects
-- is what lets a project cover several services, and leaves room for a second
-- provider without another migration of the same shape.
CREATE TABLE repositories (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    provider       text NOT NULL DEFAULT 'github' CHECK (provider IN ('github')),
    -- The provider's own identifier. For GitHub this is "owner/name", which
    -- is what arrives in a webhook payload.
    external_id    text NOT NULL,
    name           text NOT NULL,
    default_branch text NOT NULL DEFAULT 'main',
    -- Secret this repository's webhook deliveries are signed with. Lives here
    -- rather than on the project because it is per-webhook, and a leak should
    -- be contained to one repository.
    webhook_secret text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    -- One repository is watched once, so an incoming delivery resolves
    -- unambiguously.
    UNIQUE (provider, external_id)
);

CREATE INDEX repositories_project_id_idx ON repositories (project_id);

-- Carry the existing denormalised repositories across.
INSERT INTO repositories (project_id, provider, external_id, name, default_branch, webhook_secret, created_at)
SELECT id, 'github', lower(repo_owner) || '/' || lower(repo_name), repo_name,
       default_branch, webhook_secret, created_at
FROM projects;

-- Events ------------------------------------------------------------------
-- source names the system the event came from; repository_id names the
-- repository within it. Without the latter an event loses its origin as soon
-- as a project watches more than one repository.
ALTER TABLE events ADD COLUMN source text NOT NULL DEFAULT 'github';
ALTER TABLE events ADD COLUMN repository_id uuid REFERENCES repositories (id) ON DELETE CASCADE;

UPDATE events e SET repository_id = r.id
FROM repositories r WHERE r.project_id = e.project_id;

CREATE INDEX events_repository_occurred_idx ON events (repository_id, occurred_at DESC);
CREATE INDEX events_source_idx ON events (project_id, source);

-- Projections -------------------------------------------------------------
ALTER TABLE deployments ADD COLUMN repository_id uuid REFERENCES repositories (id) ON DELETE CASCADE;
UPDATE deployments d SET repository_id = r.id
FROM repositories r WHERE r.project_id = d.project_id;
ALTER TABLE deployments RENAME COLUMN sha TO commit_sha;

ALTER TABLE pull_requests ADD COLUMN repository_id uuid REFERENCES repositories (id) ON DELETE CASCADE;
UPDATE pull_requests p SET repository_id = r.id
FROM repositories r WHERE r.project_id = p.project_id;

-- Projects ----------------------------------------------------------------
-- The repository columns now live in their own table.
ALTER TABLE projects
    DROP COLUMN repo_owner,
    DROP COLUMN repo_name,
    DROP COLUMN default_branch,
    DROP COLUMN webhook_secret;

COMMIT;
