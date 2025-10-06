-- Script para limpar banco de dados mantendo apenas usuário admin
-- Remove todos os dados de usuários comuns, mas preserva admin

BEGIN;

-- 1. Buscar IDs dos usuários admin (para preservar)
-- Assumindo que admin tem email específico ou role 'admin'
-- Ajuste conforme sua lógica de identificação do admin

-- 2. Deletar dados relacionados de usuários não-admin (em ordem de dependências)

-- Deletar webhook subscriptions
DELETE FROM webhook_subscriptions 
WHERE user_id NOT IN (
    SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
);

-- Deletar contact custom fields
DELETE FROM contact_custom_fields 
WHERE contact_id IN (
    SELECT id FROM contacts WHERE project_id IN (
        SELECT id FROM projects WHERE user_id NOT IN (
            SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
        )
    )
);

-- Deletar session custom fields
DELETE FROM session_custom_fields 
WHERE session_id IN (
    SELECT id FROM sessions WHERE contact_id IN (
        SELECT id FROM contacts WHERE project_id IN (
            SELECT id FROM projects WHERE user_id NOT IN (
                SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
            )
        )
    )
);

-- Deletar contact pipeline statuses
DELETE FROM contact_pipeline_statuses 
WHERE contact_id IN (
    SELECT id FROM contacts WHERE project_id IN (
        SELECT id FROM projects WHERE user_id NOT IN (
            SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
        )
    )
);

-- Deletar mensagens (incluindo de admins de teste)
DELETE FROM messages 
WHERE project_id IN (
    SELECT id FROM projects WHERE user_id NOT IN (
        SELECT id FROM users 
        WHERE (email LIKE '%admin%' OR role = 'admin')
        AND email NOT LIKE '%.e2e@%'
    )
);

-- Deletar contact events
DELETE FROM contact_events 
WHERE contact_id IN (
    SELECT id FROM contacts WHERE project_id IN (
        SELECT id FROM projects WHERE user_id NOT IN (
            SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
        )
    )
);

-- Deletar sessões (incluindo de admins de teste)
DELETE FROM sessions 
WHERE contact_id IN (
    SELECT id FROM contacts WHERE project_id IN (
        SELECT id FROM projects WHERE user_id NOT IN (
            SELECT id FROM users 
            WHERE (email LIKE '%admin%' OR role = 'admin')
            AND email NOT LIKE '%.e2e@%'
        )
    )
);

-- Deletar agent sessions
DELETE FROM agent_sessions 
WHERE session_id IN (
    SELECT id FROM sessions WHERE contact_id IN (
        SELECT id FROM contacts WHERE project_id IN (
            SELECT id FROM projects WHERE user_id NOT IN (
                SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
            )
        )
    )
);

-- Deletar contatos (incluindo de admins de teste com email *.e2e@*)
DELETE FROM contacts 
WHERE project_id IN (
    SELECT id FROM projects WHERE user_id NOT IN (
        SELECT id FROM users 
        WHERE (email LIKE '%admin%' OR role = 'admin')
        AND email NOT LIKE '%.e2e@%'  -- Remove admins de teste
    )
);

-- Deletar pipeline statuses
DELETE FROM pipeline_statuses 
WHERE pipeline_id IN (
    SELECT id FROM pipelines WHERE project_id IN (
        SELECT id FROM projects WHERE user_id NOT IN (
            SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
        )
    )
);

-- Deletar pipelines
DELETE FROM pipelines 
WHERE project_id IN (
    SELECT id FROM projects WHERE user_id NOT IN (
        SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
    )
);

-- Deletar canais (relacionados com projects de usuários não-admin)
DELETE FROM channels 
WHERE project_id IN (
    SELECT id FROM projects WHERE user_id NOT IN (
        SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
    )
);

-- Deletar agentes (se existir relação com user)
DELETE FROM agents 
WHERE EXISTS (
    SELECT 1 FROM users u 
    WHERE agents.id IS NOT NULL 
    AND u.email NOT LIKE '%admin%' 
    AND (u.role IS NULL OR u.role != 'admin')
);

-- Deletar projetos
DELETE FROM projects 
WHERE user_id NOT IN (
    SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
);

-- Deletar API keys dos usuários não-admin
DELETE FROM user_api_keys 
WHERE user_id NOT IN (
    SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
);

-- Deletar usuários não-admin (incluindo soft deleted)
DELETE FROM users 
WHERE email NOT LIKE '%admin%' 
  AND role != 'admin'
  AND id NOT IN (
    SELECT id FROM users WHERE email LIKE '%admin%' OR role = 'admin'
  );

-- Limpar usuários soft deleted
UPDATE users SET deleted_at = NULL WHERE deleted_at IS NOT NULL;

COMMIT;

-- Mostrar estatísticas
SELECT 
    'users' as table_name, 
    COUNT(*) as remaining_records 
FROM users
UNION ALL
SELECT 'projects', COUNT(*) FROM projects
UNION ALL
SELECT 'contacts', COUNT(*) FROM contacts
UNION ALL
SELECT 'sessions', COUNT(*) FROM sessions
UNION ALL
SELECT 'messages', COUNT(*) FROM messages
UNION ALL
SELECT 'channels', COUNT(*) FROM channels;
