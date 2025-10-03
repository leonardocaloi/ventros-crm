#!/bin/bash
set -e

echo "🚀 Ventros CRM - Starting..."

# Wait for database to be ready
echo "⏳ Waiting for database..."
until psql "$DATABASE_URL" -c '\q' 2>/dev/null; do
  >&2 echo "Database is unavailable - sleeping"
  sleep 1
done
echo "✅ Database is ready!"

# Run migrations
echo "📦 Running database migrations..."
atlas migrate apply \
  --dir "file://migrations" \
  --url "${DATABASE_URL}" \
  --allow-dirty

echo "✅ Migrations completed!"

# Start application
echo "🎯 Starting API server..."
exec "$@"
