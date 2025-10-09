-- Add session timeout configuration to projects
ALTER TABLE projects
ADD COLUMN session_timeout_minutes INT NOT NULL DEFAULT 30;

-- Create index for queries
CREATE INDEX idx_projects_session_timeout ON projects(session_timeout_minutes);

-- Add comment
COMMENT ON COLUMN projects.session_timeout_minutes IS 'Default session timeout in minutes for all sessions in this project. Can be customized per project in CRM settings.';
