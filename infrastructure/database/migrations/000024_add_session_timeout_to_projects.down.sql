-- Remove session timeout configuration from projects
DROP INDEX IF EXISTS idx_projects_session_timeout;

ALTER TABLE projects
DROP COLUMN IF EXISTS session_timeout_minutes;
