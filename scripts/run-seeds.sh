#!/bin/bash
set -e

# Script para executar seeds do banco de dados
# Uso: ./scripts/run-seeds.sh

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-ventros}
DB_PASSWORD=${DB_PASSWORD:-ventros123}
DB_NAME=${DB_NAME:-ventros_crm}

echo "🌱 Running database seeds..."
echo "📍 Host: $DB_HOST:$DB_PORT"
echo "📦 Database: $DB_NAME"
echo ""

# Executa cada arquivo SQL na ordem
for seed_file in deployments/docker/seeds/*.sql; do
    if [ -f "$seed_file" ]; then
        echo "▶️  Executing: $(basename $seed_file)"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$seed_file"
        echo "✅ Done"
        echo ""
    fi
done

echo "🎉 All seeds executed successfully!"
