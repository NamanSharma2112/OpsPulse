BEGIN;

ALTER TABLE repositories
    DROP COLUMN connected_by,
    DROP COLUMN webhook_installed_at,
    DROP COLUMN webhook_external_id;

ALTER TABLE users
    DROP COLUMN github_connected_at,
    DROP COLUMN github_token_scopes,
    DROP COLUMN github_access_token;

COMMIT;
