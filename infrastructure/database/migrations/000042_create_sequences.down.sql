-- Drop indexes for sequence_enrollments
DROP INDEX IF EXISTS idx_enrollments_sequence_contact_unique;
DROP INDEX IF EXISTS idx_enrollments_next_scheduled;
DROP INDEX IF EXISTS idx_enrollments_status;
DROP INDEX IF EXISTS idx_enrollments_contact_id;
DROP INDEX IF EXISTS idx_enrollments_sequence_id;

-- Drop sequence_enrollments table
DROP TABLE IF EXISTS sequence_enrollments;

-- Drop indexes for sequence_steps
DROP INDEX IF EXISTS idx_sequence_steps_sequence_order;
DROP INDEX IF EXISTS idx_sequence_steps_sequence_id;

-- Drop sequence_steps table
DROP TABLE IF EXISTS sequence_steps;

-- Drop indexes for sequences
DROP INDEX IF EXISTS idx_sequences_trigger_type;
DROP INDEX IF EXISTS idx_sequences_status;
DROP INDEX IF EXISTS idx_sequences_tenant;

-- Drop sequences table
DROP TABLE IF EXISTS sequences;
