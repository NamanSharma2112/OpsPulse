-- GitHub OAuth: store the user's access token so OpsPulse can act on their
-- behalf (list repositories, install webhooks), and record the webhook we
-- installed on each repository.
BEGIN;

ALTER TABLE users
    -- Sealed with AES-256-GCM before it reaches the database, so a dump on
    -- its own does not yield a token that can reach someone's repositories.
    ADD COLUMN github_access_token   text,
    ADD COLUMN github_token_scopes   text,
    ADD COLUMN github_connected_at   timestamptz;

ALTER TABLE repositories
    -- GitHub's own hook id, set when OpsPulse installed the webhook itself.
    -- Null means the webhook was configured by hand.
    ADD COLUMN webhook_external_id bigint,
    ADD COLUMN webhook_installed_at timestamptz,
    -- The user whose token installed it, so it can be removed later with a
    -- token that has the rights to do so.
    ADD COLUMN connected_by uuid REFERENCES users (id) ON DELETE SET NULL;

COMMIT;
