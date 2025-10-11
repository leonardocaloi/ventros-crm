-- Remove unique constraint from contact_custom_fields
ALTER TABLE contact_custom_fields
DROP CONSTRAINT IF EXISTS uq_contact_custom_fields_contact_key;
