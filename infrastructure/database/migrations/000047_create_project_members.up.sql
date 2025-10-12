-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- Project Members Table (RBAC - Role-Based Access Control)
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- Representa a associação entre um Agent (usuário Keycloak) e um Project
-- Define o role (admin, supervisor, agent, viewer) do usuário no projeto

CREATE TABLE IF NOT EXISTS project_members (
    -- Primary Key
    id UUID PRIMARY KEY,

    -- Optimistic Locking
    version INTEGER NOT NULL DEFAULT 1,

    -- Core Fields
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,  -- Keycloak user ID (sub claim)
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'supervisor', 'agent', 'viewer')),

    -- Audit Fields
    invited_by VARCHAR(255) NOT NULL,  -- Keycloak user ID who invited
    invited_at TIMESTAMP NOT NULL,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,  -- Soft delete

    -- Unique Constraint: Um agent só pode ter UM role por projeto
    CONSTRAINT unique_project_agent UNIQUE (project_id, agent_id)
);

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- Indexes
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

-- Index: Buscar membros de um projeto
CREATE INDEX idx_project_members_project ON project_members(project_id) WHERE deleted_at IS NULL;

-- Index: Buscar projetos de um agent
CREATE INDEX idx_project_members_agent ON project_members(agent_id) WHERE deleted_at IS NULL;

-- Index: Buscar admins de um projeto
CREATE INDEX idx_project_members_project_role ON project_members(project_id, role) WHERE deleted_at IS NULL;

-- Index: Soft delete
CREATE INDEX idx_project_members_deleted_at ON project_members(deleted_at);

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- Comments
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

COMMENT ON TABLE project_members IS 'Project members with role-based permissions (RBAC)';

COMMENT ON COLUMN project_members.id IS 'Unique identifier for the project member';
COMMENT ON COLUMN project_members.version IS 'Version for optimistic locking';
COMMENT ON COLUMN project_members.project_id IS 'Reference to the project';
COMMENT ON COLUMN project_members.agent_id IS 'Keycloak user ID (sub claim from JWT)';
COMMENT ON COLUMN project_members.role IS 'Role: admin (full access), supervisor (management), agent (operations), viewer (read-only)';
COMMENT ON COLUMN project_members.invited_by IS 'Keycloak user ID who invited this member';
COMMENT ON COLUMN project_members.invited_at IS 'When the member was invited';
