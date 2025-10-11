-- Add unique constraint to contact_custom_fields to ensure one value per field_key per contact
-- This also enables upsert operations using ON CONFLICT

-- First, remove any duplicate records (keep the most recent one)
DELETE FROM contact_custom_fields a
USING contact_custom_fields b
WHERE a.id < b.id
  AND a.contact_id = b.contact_id
  AND a.field_key = b.field_key;

-- Add unique constraint
ALTER TABLE contact_custom_fields
ADD CONSTRAINT uq_contact_custom_fields_contact_key UNIQUE (contact_id, field_key);
