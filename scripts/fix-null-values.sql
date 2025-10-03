-- Fix NULL values in database before migration
-- This should be run before applying Ent migrations

-- Fix contacts table
UPDATE contacts 
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, created_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;

-- Fix sessions table  
UPDATE sessions 
SET created_at = COALESCE(created_at, started_at, NOW()),
    updated_at = COALESCE(updated_at, created_at, started_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;

-- Fix projects table
UPDATE projects 
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, created_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;

-- Fix messages table (if exists)
UPDATE messages 
SET created_at = COALESCE(created_at, timestamp, NOW()),
    updated_at = COALESCE(updated_at, created_at, timestamp, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;

-- Fix pipelines table (if exists)
UPDATE pipelines 
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, created_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;

-- Fix users table (if exists)
UPDATE users 
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, created_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;
