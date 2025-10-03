-- Seed inicial de Channel Types
-- Executado automaticamente após migrations

INSERT INTO channel_types (id, name, description, active, created_at, updated_at) 
VALUES 
    (1, 'waha', 'WAHA - WhatsApp HTTP API (Multi-device)', true, NOW(), NOW()),
    (2, 'whatsapp', 'WhatsApp Business API Official', true, NOW(), NOW()),
    (3, 'direct_ig', 'Instagram Direct Messages', true, NOW(), NOW()),
    (4, 'messenger', 'Facebook Messenger', true, NOW(), NOW()),
    (5, 'telegram', 'Telegram Bot API', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Reset sequence
SELECT setval('channel_types_id_seq', (SELECT MAX(id) FROM channel_types));
