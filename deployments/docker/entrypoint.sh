#!/bin/sh
set -e

echo "🔄 Running database migrations..."
./migrate-auth

echo "🚀 Starting API server..."
exec ./main "$@"
