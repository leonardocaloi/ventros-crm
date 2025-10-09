#!/bin/bash

# Script de refatoração: Follow-up Rules → Automation Rules
# Respeita DDD: Domain → Application → Infrastructure

echo "🔄 Iniciando refatoração Follow-up → Automation..."

# Função para substituir em arquivo
replace_in_file() {
    file=$1
    sed -i 's/FollowUpRule/AutomationRule/g' "$file"
    sed -i 's/FollowUpTrigger/AutomationTrigger/g' "$file"
    sed -i 's/FollowUpAction/AutomationAction/g' "$file"
    sed -i 's/follow_up_rule/automation_rule/g' "$file"
    sed -i 's/follow-up-rule/automation-rule/g' "$file"
    sed -i 's/follow_up/automation/g' "$file"
    sed -i 's/followup/automation/g' "$file"
    sed -i 's/FollowUp/Automation/g' "$file"
    echo "  ✅ $file"
}

# 1. DOMAIN LAYER
echo ""
echo "📦 Refatorando Domain Layer..."
for file in internal/domain/pipeline/*.go; do
    if grep -q "FollowUp\|follow_up" "$file" 2>/dev/null; then
        replace_in_file "$file"
    fi
done

# Renomear arquivo scheduled_rule.go
if [ -f "internal/domain/pipeline/scheduled_rule.go" ]; then
    mv internal/domain/pipeline/scheduled_rule.go internal/domain/pipeline/scheduled_automation.go
    echo "  📝 Renamed: scheduled_rule.go → scheduled_automation.go"
fi

# 2. APPLICATION LAYER
echo ""
echo "🎯 Refatorando Application Layer..."

# Renomear arquivos
if [ -f "internal/application/pipeline/follow_up_engine.go" ]; then
    mv internal/application/pipeline/follow_up_engine.go internal/application/pipeline/automation_engine.go
    echo "  📝 Renamed: follow_up_engine.go → automation_engine.go"
fi

if [ -f "internal/application/pipeline/follow_up_action_executor.go" ]; then
    mv internal/application/pipeline/follow_up_action_executor.go internal/application/pipeline/automation_action_executor.go
    echo "  📝 Renamed: follow_up_action_executor.go → automation_action_executor.go"
fi

if [ -f "internal/application/pipeline/follow_up_rule_manager.go" ]; then
    mv internal/application/pipeline/follow_up_rule_manager.go internal/application/pipeline/automation_rule_manager.go
    echo "  📝 Renamed: follow_up_rule_manager.go → automation_rule_manager.go"
fi

if [ -f "internal/application/pipeline/follow_up_integration.go" ]; then
    mv internal/application/pipeline/follow_up_integration.go internal/application/pipeline/automation_integration.go
    echo "  📝 Renamed: follow_up_integration.go → automation_integration.go"
fi

# Substituir conteúdo
for file in internal/application/pipeline/*.go; do
    if grep -q "FollowUp\|follow_up" "$file" 2>/dev/null; then
        replace_in_file "$file"
    fi
done

# 3. INFRASTRUCTURE LAYER
echo ""
echo "🏗️ Refatorando Infrastructure Layer..."

# Entities
if [ -f "infrastructure/persistence/entities/follow_up_rule.go" ]; then
    mv infrastructure/persistence/entities/follow_up_rule.go infrastructure/persistence/entities/automation_rule.go
    echo "  📝 Renamed: follow_up_rule.go → automation_rule.go"
fi

for file in infrastructure/persistence/entities/*.go; do
    if grep -q "FollowUp\|follow_up" "$file" 2>/dev/null; then
        replace_in_file "$file"
    fi
done

# Repositories
if [ -f "infrastructure/persistence/gorm_follow_up_rule_repository.go" ]; then
    mv infrastructure/persistence/gorm_follow_up_rule_repository.go infrastructure/persistence/gorm_automation_rule_repository.go
    echo "  📝 Renamed: gorm_follow_up_rule_repository.go → gorm_automation_rule_repository.go"
fi

for file in infrastructure/persistence/*.go; do
    if grep -q "FollowUp\|follow_up" "$file" 2>/dev/null; then
        replace_in_file "$file"
    fi
done

# Workers
if [ -f "infrastructure/workflow/scheduled_rules_worker.go" ]; then
    mv infrastructure/workflow/scheduled_rules_worker.go infrastructure/workflow/scheduled_automation_worker.go
    echo "  📝 Renamed: scheduled_rules_worker.go → scheduled_automation_worker.go"
fi

for file in infrastructure/workflow/*.go; do
    if grep -q "FollowUp\|follow_up\|followup" "$file" 2>/dev/null; then
        replace_in_file "$file"
    fi
done

# Migrations - renomear
if [ -f "infrastructure/database/migrations/000019_create_follow_up_rules_table.up.sql" ]; then
    mv infrastructure/database/migrations/000019_create_follow_up_rules_table.up.sql infrastructure/database/migrations/000019_create_automation_rules_table.up.sql
    echo "  📝 Renamed: 000019_create_follow_up_rules_table.up.sql → 000019_create_automation_rules_table.up.sql"
fi

if [ -f "infrastructure/database/migrations/000019_create_follow_up_rules_table.down.sql" ]; then
    mv infrastructure/database/migrations/000019_create_follow_up_rules_table.down.sql infrastructure/database/migrations/000019_create_automation_rules_table.down.sql
    echo "  📝 Renamed: 000019_create_follow_up_rules_table.down.sql → 000019_create_automation_rules_table.down.sql"
fi

# Migrations - conteúdo
for file in infrastructure/database/migrations/*.sql; do
    if grep -q "follow_up\|followup" "$file" 2>/dev/null; then
        replace_in_file "$file"
    fi
done

# 4. DOCUMENTAÇÃO
echo ""
echo "📚 Refatorando Documentação..."
for file in docs/*.md; do
    if grep -q "Follow-up\|follow-up\|FollowUp\|follow_up" "$file" 2>/dev/null; then
        replace_in_file "$file"
        # Ajustes específicos de docs
        sed -i 's/Automation-up/Follow-up/g' "$file"  # reverter casos específicos
    fi
done

echo ""
echo "✅ Refatoração concluída!"
echo ""
echo "📊 Resumo:"
echo "  - Domain Layer: FollowUpRule → AutomationRule"
echo "  - Application Layer: FollowUpEngine → AutomationEngine"
echo "  - Infrastructure Layer: follow_up_rules → automation_rules"
echo "  - Migrations: Tabela renomeada"
echo ""
echo "⚠️ Próximos passos manuais:"
echo "  1. Revisar código gerado"
echo "  2. Rodar: go build ./..."
echo "  3. Rodar: go test ./..."
echo "  4. Atualizar migration se necessário"
