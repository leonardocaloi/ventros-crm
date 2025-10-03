-- Seed de projeto e customer default para desenvolvimento
-- Em produção, isso seria criado via API

-- Default Customer
INSERT INTO customers (id, name, email, status, created_at, updated_at)
VALUES 
    ('00000000-0000-0000-0000-000000000001'::uuid, 'Default Customer', 'default@ventros.dev', 'active', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Default Project
INSERT INTO projects (id, customer_id, tenant_id, name, active, created_at, updated_at)
VALUES 
    ('00000000-0000-0000-0000-000000000002'::uuid, '00000000-0000-0000-0000-000000000001'::uuid, 'default-tenant', 'Default Project', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
