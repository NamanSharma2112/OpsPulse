-- Projections M2 analytics needs that the earlier milestones did not keep.
--
-- Issues were only stored when they were labelled as incidents, so "how many
-- issues are open" had no answer. Pushes were not projected at all, so commit
-- frequency had nothing to count. Both are derived from the event log and can
-- be rebuilt by replaying it.
BEGIN;

CREATE TABLE issues (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    repository_id uuid REFERENCES repositories (id) ON DELETE CASCADE,
    number        integer NOT NULL,
    title         text NOT NULL,
    author        text NOT NULL DEFAULT '',
    state         text NOT NULL CHECK (state IN ('open', 'closed')),
    url           text,
    opened_at     timestamptz NOT NULL,
    closed_at     timestamptz,
    updated_at    timestamptz NOT NULL DEFAULT now(),

    UNIQUE (project_id, number),
    -- A closed issue has a closing time; an open one does not.
    CONSTRAINT issues_closed_has_timestamp
        CHECK ((state = 'closed') = (closed_at IS NOT NULL))
);

-- The open-issue count is the hot read, so it leads the index.
CREATE INDEX issues_project_state_idx ON issues (project_id, state, opened_at DESC);

CREATE TABLE commits (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    repository_id uuid REFERENCES repositories (id) ON DELETE CASCADE,
    sha           text NOT NULL,
    message       text NOT NULL DEFAULT '',
    author        text NOT NULL DEFAULT '',
    -- The branch the push landed on, without the refs/heads/ prefix.
    branch        text NOT NULL DEFAULT '',
    committed_at  timestamptz NOT NULL,

    -- One commit can arrive twice: pushed to a branch, then to the default
    -- branch on merge. The sha is what makes the second delivery a no-op.
    UNIQUE (project_id, sha)
);

CREATE INDEX commits_project_time_idx ON commits (project_id, committed_at DESC);

COMMIT;
