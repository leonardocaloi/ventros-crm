-- Add connection_mode column to channels table
--
-- This migration adds support for two WAHA connection modes:
-- - manual: User provides existing WAHA credentials (session_id, base_url, token)
-- - auto: System creates and manages WAHA session, returns QR code
--
-- Channel types:
-- - "waha" -> manual mode (user manages WAHA)
-- - "whatsapp_business" -> auto mode (system manages WAHA, provides QR code)

-- 1. Add connection_mode column with default 'manual'
ALTER TABLE channels
ADD COLUMN IF NOT EXISTS connection_mode VARCHAR(20) DEFAULT 'manual';

-- 2. Create index for performance
CREATE INDEX IF NOT EXISTS idx_channels_connection_mode
ON channels(connection_mode);

-- 3. Update existing channels based on type
--    - TypeWAHA channels use manual mode (user provides credentials)
--    - TypeWhatsAppBusiness channels use auto mode (system manages session)
UPDATE channels
SET connection_mode = 'manual'
WHERE type = 'waha' AND connection_mode IS NULL;

UPDATE channels
SET connection_mode = 'auto'
WHERE type = 'whatsapp_business' AND connection_mode IS NULL;

-- 4. Add comment to document the field
COMMENT ON COLUMN channels.connection_mode IS 'Connection mode: manual (user provides WAHA credentials) or auto (system manages WAHA session with QR code)';
