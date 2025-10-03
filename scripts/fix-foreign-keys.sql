-- Fix foreign key from customers to users in projects table

-- 1. Drop old foreign key if exists
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'projects_customers_projects' 
        AND table_name = 'projects'
    ) THEN
        ALTER TABLE projects DROP CONSTRAINT projects_customers_projects;
        RAISE NOTICE 'Dropped old constraint: projects_customers_projects';
    END IF;
END $$;

-- 2. Drop old foreign key if exists (alternative name)
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_projects_customer' 
        AND table_name = 'projects'
    ) THEN
        ALTER TABLE projects DROP CONSTRAINT fk_projects_customer;
        RAISE NOTICE 'Dropped old constraint: fk_projects_customer';
    END IF;
END $$;

-- 3. Create new foreign key to users table
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_projects_user' 
        AND table_name = 'projects'
    ) THEN
        ALTER TABLE projects 
        ADD CONSTRAINT fk_projects_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE;
        RAISE NOTICE 'Created new constraint: fk_projects_user';
    END IF;
END $$;

-- 4. Verify the constraint
SELECT 
    tc.constraint_name, 
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
AND tc.table_name='projects'
AND kcu.column_name = 'user_id';

-- Success message
DO $$ 
BEGIN
    RAISE NOTICE '✅ Foreign key migration completed successfully!';
    RAISE NOTICE '📋 Projects table now references users table correctly';
END $$;
