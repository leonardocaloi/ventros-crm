#!/bin/bash
# PostgreSQL Initialization Script
# Creates additional databases needed by the system
#
# This script runs automatically when PostgreSQL container starts

set -e

# Create Keycloak database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname="$POSTGRES_DB" <<-EOSQL
    -- Create Keycloak database if it doesn't exist
    SELECT 'CREATE DATABASE keycloak'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak')\gexec

    -- Grant privileges to ventros user
    GRANT ALL PRIVILEGES ON DATABASE keycloak TO $POSTGRES_USER;
EOSQL

echo "✓ Keycloak database created successfully"
