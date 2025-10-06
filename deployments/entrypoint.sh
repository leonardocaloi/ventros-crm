#!/bin/sh
set -e

echo "🚀 Ventros CRM - Starting..."

# Start application directly
# Migrations are handled by GORM auto-migrate in the application
echo "🎯 Starting API server..."
exec ./main
