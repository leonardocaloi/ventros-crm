-- =====================================================
-- RLS (Row Level Security) Setup for Ventros CRM
-- =====================================================

-- Criar role para aplicação
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_user') THEN
        CREATE ROLE app_user;
    END IF;
END
$$;

-- Garantir que o usuário ventros pode assumir o role app_user
GRANT app_user TO ventros;

-- =====================================================
-- HABILITAR RLS NAS TABELAS
-- =====================================================

-- Users (não precisa de RLS - cada user vê apenas seu próprio perfil via app logic)
-- ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Projects - usuário só vê seus próprios projetos
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- Pipelines - usuário só vê pipelines de seus projetos
ALTER TABLE pipelines ENABLE ROW LEVEL SECURITY;

-- Pipeline Statuses - usuário só vê status de pipelines de seus projetos
ALTER TABLE pipeline_statuses ENABLE ROW LEVEL SECURITY;

-- Contacts - usuário só vê contatos de seus projetos
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;

-- Contact Pipeline Statuses - usuário só vê status de seus contatos
ALTER TABLE contact_pipeline_statuses ENABLE ROW LEVEL SECURITY;

-- Contact Status History - usuário só vê histórico de seus contatos
ALTER TABLE contact_status_histories ENABLE ROW LEVEL SECURITY;

-- Messages - usuário só vê mensagens de seus contatos
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;

-- Sessions - usuário só vê sessões de seus contatos
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

-- Webhook Subscriptions - usuário só vê seus próprios webhooks
ALTER TABLE webhook_subscriptions ENABLE ROW LEVEL SECURITY;

-- Channels - usuário só vê seus próprios canais
ALTER TABLE channels ENABLE ROW LEVEL SECURITY;

-- =====================================================
-- POLÍTICAS RLS
-- =====================================================

-- Projects: usuário só vê seus próprios projetos
DROP POLICY IF EXISTS user_projects_policy ON projects;
CREATE POLICY user_projects_policy ON projects
    FOR ALL TO app_user
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- Pipelines: usuário só vê pipelines de seus projetos
DROP POLICY IF EXISTS user_pipelines_policy ON pipelines;
CREATE POLICY user_pipelines_policy ON pipelines
    FOR ALL TO app_user
    USING (
        project_id IN (
            SELECT id FROM projects 
            WHERE user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Pipeline Statuses: usuário só vê status de pipelines de seus projetos
DROP POLICY IF EXISTS user_pipeline_statuses_policy ON pipeline_statuses;
CREATE POLICY user_pipeline_statuses_policy ON pipeline_statuses
    FOR ALL TO app_user
    USING (
        pipeline_id IN (
            SELECT p.id FROM pipelines p
            JOIN projects pr ON p.project_id = pr.id
            WHERE pr.user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Contacts: usuário só vê contatos de seus projetos
DROP POLICY IF EXISTS user_contacts_policy ON contacts;
CREATE POLICY user_contacts_policy ON contacts
    FOR ALL TO app_user
    USING (
        project_id IN (
            SELECT id FROM projects 
            WHERE user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Contact Pipeline Statuses: usuário só vê status de seus contatos
DROP POLICY IF EXISTS user_contact_pipeline_statuses_policy ON contact_pipeline_statuses;
CREATE POLICY user_contact_pipeline_statuses_policy ON contact_pipeline_statuses
    FOR ALL TO app_user
    USING (
        contact_id IN (
            SELECT c.id FROM contacts c
            JOIN projects p ON c.project_id = p.id
            WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Contact Status History: usuário só vê histórico de seus contatos
DROP POLICY IF EXISTS user_contact_status_history_policy ON contact_status_histories;
CREATE POLICY user_contact_status_history_policy ON contact_status_histories
    FOR ALL TO app_user
    USING (
        contact_id IN (
            SELECT c.id FROM contacts c
            JOIN projects p ON c.project_id = p.id
            WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Messages: usuário só vê mensagens de seus contatos
DROP POLICY IF EXISTS user_messages_policy ON messages;
CREATE POLICY user_messages_policy ON messages
    FOR ALL TO app_user
    USING (
        contact_id IN (
            SELECT c.id FROM contacts c
            JOIN projects p ON c.project_id = p.id
            WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Sessions: usuário só vê sessões de seus contatos
DROP POLICY IF EXISTS user_sessions_policy ON sessions;
CREATE POLICY user_sessions_policy ON sessions
    FOR ALL TO app_user
    USING (
        contact_id IN (
            SELECT c.id FROM contacts c
            JOIN projects p ON c.project_id = p.id
            WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
        )
    );

-- Webhook Subscriptions: usuário só vê seus próprios webhooks
DROP POLICY IF EXISTS user_webhook_subscriptions_policy ON webhook_subscriptions;
CREATE POLICY user_webhook_subscriptions_policy ON webhook_subscriptions
    FOR ALL TO app_user
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- Channels: usuário só vê seus próprios canais
DROP POLICY IF EXISTS user_channels_policy ON channels;
CREATE POLICY user_channels_policy ON channels
    FOR ALL TO app_user
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- =====================================================
-- PERMISSÕES PARA APP_USER
-- =====================================================

-- Conceder permissões básicas para app_user
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- =====================================================
-- FUNÇÕES HELPER PARA RLS
-- =====================================================

-- Função para definir o usuário atual na sessão
CREATE OR REPLACE FUNCTION set_current_user_id(user_uuid uuid)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_user_id', user_uuid::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para obter o usuário atual
CREATE OR REPLACE FUNCTION get_current_user_id()
RETURNS uuid AS $$
BEGIN
    RETURN current_setting('app.current_user_id', true)::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- ÍNDICES PARA PERFORMANCE DO RLS
-- =====================================================

-- Índices para otimizar as consultas RLS
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_project_id ON pipelines(project_id);
CREATE INDEX IF NOT EXISTS idx_contacts_project_id ON contacts(project_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_statuses_pipeline_id ON pipeline_statuses(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_messages_contact_id ON messages(contact_id);
CREATE INDEX IF NOT EXISTS idx_sessions_contact_id ON sessions(contact_id);

-- =====================================================
-- LOGS E CONFIRMAÇÃO
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '✅ RLS (Row Level Security) configurado com sucesso!';
    RAISE NOTICE '📋 Tabelas com RLS habilitado:';
    RAISE NOTICE '   - projects, pipelines, pipeline_statuses';
    RAISE NOTICE '   - contacts, contact_pipeline_statuses, contact_status_histories';
    RAISE NOTICE '   - messages, sessions, webhook_subscriptions, channels';
    RAISE NOTICE '🔒 Políticas criadas para isolamento por user_id';
    RAISE NOTICE '⚡ Índices criados para performance';
END
$$;
