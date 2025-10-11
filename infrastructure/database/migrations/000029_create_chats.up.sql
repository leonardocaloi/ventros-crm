-- Create chats table to support individual, group, and channel conversations
CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    chat_type VARCHAR(50) NOT NULL CHECK (chat_type IN ('individual', 'group', 'channel')),
    subject VARCHAR(255), -- Group/channel name (NULL for individual chats)
    description TEXT, -- Group/channel description
    participants JSONB NOT NULL DEFAULT '[]'::jsonb, -- Array of participant objects with id, type, joined_at, left_at, is_admin
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'closed')),
    metadata JSONB DEFAULT '{}'::jsonb, -- Flexible metadata storage
    last_message_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    CONSTRAINT fk_chats_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_chats_project ON chats(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_tenant ON chats(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_tenant_status ON chats(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_tenant_type ON chats(tenant_id, chat_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_type ON chats(chat_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_status ON chats(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_last_message ON chats(last_message_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_created ON chats(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_updated ON chats(updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_deleted ON chats(deleted_at);

-- GIN indexes for JSONB queries
CREATE INDEX idx_chats_participants ON chats USING GIN (participants) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_metadata ON chats USING GIN (metadata) WHERE deleted_at IS NULL;

-- Subject search index (for group/channel search)
CREATE INDEX idx_chats_subject ON chats(subject) WHERE deleted_at IS NULL AND subject IS NOT NULL;
