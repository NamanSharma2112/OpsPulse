BEGIN;

ALTER TABLE projects
    ADD COLUMN repo_owner     text NOT NULL DEFAULT '',
    ADD COLUMN repo_name      text NOT NULL DEFAULT '',
    ADD COLUMN default_branch text NOT NULL DEFAULT 'main',
    ADD COLUMN webhook_secret text NOT NULL DEFAULT '';

-- Fold the first repository of each project back onto the project. A project
-- watching several repositories keeps only its oldest, and owner/name come
-- back lowercased because external_id is stored as the lookup key.
UPDATE projects p SET
    repo_owner     = split_part(r.external_id, '/', 1),
    repo_name      = split_part(r.external_id, '/', 2),
    default_branch = r.default_branch,
    webhook_secret = r.webhook_secret
FROM (
    SELECT DISTINCT ON (project_id) project_id, external_id, default_branch, webhook_secret
    FROM repositories ORDER BY project_id, created_at
) r
WHERE r.project_id = p.id;

CREATE UNIQUE INDEX projects_repo_key ON projects (lower(repo_owner), lower(repo_name));

ALTER TABLE pull_requests DROP COLUMN repository_id;
ALTER TABLE deployments RENAME COLUMN commit_sha TO sha;
ALTER TABLE deployments DROP COLUMN repository_id;

DROP INDEX IF EXISTS events_source_idx;
DROP INDEX IF EXISTS events_repository_occurred_idx;
ALTER TABLE events DROP COLUMN repository_id;
ALTER TABLE events DROP COLUMN source;

DROP TABLE repositories;

ALTER TABLE projects RENAME COLUMN organization_id TO org_id;
ALTER TABLE organization_members RENAME COLUMN organization_id TO org_id;
ALTER TABLE organization_members RENAME TO org_members;
ALTER TABLE organizations RENAME TO orgs;

COMMIT;
