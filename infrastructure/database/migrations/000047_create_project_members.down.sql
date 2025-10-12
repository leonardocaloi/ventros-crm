-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- Rollback: Drop Project Members Table
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DROP INDEX IF EXISTS idx_project_members_deleted_at;
DROP INDEX IF EXISTS idx_project_members_project_role;
DROP INDEX IF EXISTS idx_project_members_agent;
DROP INDEX IF EXISTS idx_project_members_project;

DROP TABLE IF EXISTS project_members;
