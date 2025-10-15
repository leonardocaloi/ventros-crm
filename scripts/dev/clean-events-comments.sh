#!/bin/bash
# Script to clean Portuguese comments from event files
# Converts to English following Google Go Style Guide

set -euo pipefail

echo "🧹 Cleaning event files - Removing Portuguese, adding English Google Go comments..."
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter
TOTAL=0
SUCCESS=0
SKIP=0

# Function to clean a file
clean_file() {
    local file=$1
    local filename=$(basename "$file")
    
    echo -e "${YELLOW}Processing:${NC} $file"
    
    # Check if file exists
    if [ ! -f "$file" ]; then
        echo -e "${RED}❌ File not found${NC}"
        return 1
    fi
    
    # Create backup
    cp "$file" "${file}.backup"
    
    # Common Portuguese → English replacements
    sed -i 's|// DomainEvent é um alias para shared.DomainEvent (compatibilidade retroativa).|type DomainEvent = shared.DomainEvent|g' "$file"
    sed -i 's|DomainEvent é a interface base para eventos de domínio|DomainEvent interface for domain events|g' "$file"
    sed -i 's|DomainEvent é a interface para todos os eventos de domínio|DomainEvent interface for domain events|g' "$file"
    sed -i 's|DomainEvent é a interface para eventos de domínio|DomainEvent interface for domain events|g' "$file"
    
    # Event patterns
    sed -i 's| - Pipeline criado| is emitted when a pipeline is created.|g' "$file"
    sed -i 's| - Pipeline atualizado| is emitted when a pipeline is updated.|g' "$file"
    sed -i 's| - Pipeline ativado| is emitted when a pipeline is activated.|g' "$file"
    sed -i 's| - Pipeline desativado| is emitted when a pipeline is deactivated.|g' "$file"
    
    sed -i 's| - Status criado| is emitted when a status is created.|g' "$file"
    sed -i 's| - Status atualizado| is emitted when a status is updated.|g' "$file"
    sed -i 's| - Status ativado| is emitted when a status is activated.|g' "$file"
    sed -i 's| - Status desativado| is emitted when a status is deactivated.|g' "$file"
    
    sed -i 's| - Status adicionado ao pipeline| is emitted when a status is added to pipeline.|g' "$file"
    sed -i 's| - Status removido do pipeline| is emitted when a status is removed from pipeline.|g' "$file"
    
    sed -i 's| - Status do contato alterado| is emitted when contact status changes.|g' "$file"
    sed -i 's| - Contato entrou no pipeline| is emitted when a contact enters pipeline.|g' "$file"
    sed -i 's| - Contato saiu do pipeline| is emitted when a contact exits pipeline.|g' "$file"
    
    sed -i 's| - Regra de follow-up criada| is emitted when an automation rule is created.|g' "$file"
    sed -i 's| - Automação ativada| is emitted when automation is enabled.|g' "$file"
    sed -i 's| - Automação desativada| is emitted when automation is disabled.|g' "$file"
    sed -i 's| - Regra de follow-up disparada| is emitted when automation rule is triggered.|g' "$file"
    sed -i 's| - Regra de follow-up executada com sucesso| is emitted when automation rule is executed.|g' "$file"
    sed -i 's| - Regra de follow-up falhou ao executar| is emitted when automation rule fails.|g' "$file"
    
    # Channel events
    sed -i 's| - Canal criado no sistema.| is emitted when a channel is created.|g' "$file"
    sed -i 's| - Canal ativado.| is emitted when a channel is activated.|g' "$file"
    sed -i 's| - Canal desativado.| is emitted when a channel is deactivated.|g' "$file"
    sed -i 's| - Canal deletado.| is emitted when a channel is deleted.|g' "$file"
    sed -i 's| - Pipeline associado ao canal.| is emitted when a pipeline is associated to channel.|g' "$file"
    sed -i 's| - Pipeline desassociado do canal.| is emitted when a pipeline is disassociated from channel.|g' "$file"
    
    # Agent events  
    sed -i 's| - Agente criado no sistema.| is emitted when an agent is created.|g' "$file"
    sed -i 's| - Informações do agente atualizadas.| is emitted when agent information is updated.|g' "$file"
    sed -i 's| - Agente ativado.| is emitted when an agent is activated.|g' "$file"
    sed -i 's| - Agente desativado.| is emitted when an agent is deactivated.|g' "$file"
    sed -i 's| - Agente fez login.| is emitted when an agent logs in.|g' "$file"
    sed -i 's| - Permissão concedida ao agente.| is emitted when permission is granted to agent.|g' "$file"
    sed -i 's| - Permissão revogada do agente.| is emitted when permission is revoked from agent.|g' "$file"
    
    # Billing events
    sed -i 's|é disparado quando uma conta de faturamento é criada|is emitted when a billing account is created|g' "$file"
    sed -i 's|é disparado quando um método de pagamento é ativado|is emitted when a payment method is activated|g' "$file"
    sed -i 's|é disparado quando uma conta é suspensa|is emitted when an account is suspended|g' "$file"
    sed -i 's|é disparado quando uma conta é reativada|is emitted when an account is reactivated|g' "$file"
    sed -i 's|é disparado quando uma conta é cancelada|is emitted when an account is canceled|g' "$file"
    
    # Tracking events
    sed -i 's| - Tracking criado no sistema.| is emitted when tracking is created.|g' "$file"
    sed -i 's| - Tracking enriquecido com dados adicionais.| is emitted when tracking is enriched with additional data.|g' "$file"
    sed -i 's|creates a new evento de tracking criado|creates a new tracking created event|g' "$file"
    sed -i 's|creates a new evento de tracking enriquecido|creates a new tracking enriched event|g' "$file"
    
    # Note events
    sed -i 's|é disparado quando uma nota é adicionada|is emitted when a note is added|g' "$file"
    sed -i 's|é disparado quando uma nota é atualizada|is emitted when a note is updated|g' "$file"
    sed -i 's|é disparado quando uma nota é deletada|is emitted when a note is deleted|g' "$file"
    sed -i 's|é disparado quando uma nota é fixada|is emitted when a note is pinned|g' "$file"
    
    # Credential events
    sed -i 's| - Credencial criada| is emitted when a credential is created.|g' "$file"
    sed -i 's| - Credencial atualizada| is emitted when a credential is updated.|g' "$file"
    sed -i 's| - Token OAuth renovado| is emitted when OAuth token is refreshed.|g' "$file"
    sed -i 's| - Credencial ativada| is emitted when a credential is activated.|g' "$file"
    sed -i 's| - Credencial desativada| is emitted when a credential is deactivated.|g' "$file"
    sed -i 's| - Credencial foi usada| is emitted when a credential is used.|g' "$file"
    sed -i 's| - Credencial expirou| is emitted when a credential expires.|g' "$file"
    
    # Contact list events
    sed -i 's|evento disparado quando uma listas é criada|is emitted when a contact list is created|g' "$file"
    sed -i 's|evento disparado quando uma listas é atualizada|is emitted when a contact list is updated|g' "$file"
    sed -i 's|evento disparado quando uma listas é deletada|is emitted when a contact list is deleted|g' "$file"
    sed -i 's|evento disparado quando uma regra de filtro é adicionada|is emitted when a filter rule is added|g' "$file"
    sed -i 's|evento disparado quando uma regra de filtro é removida|is emitted when a filter rule is removed|g' "$file"
    sed -i 's|evento disparado quando todas as regras são removidas|is emitted when all filter rules are cleared|g' "$file"
    sed -i 's|evento disparado quando a lista é recalculada|is emitted when the list is recalculated|g' "$file"
    sed -i 's|evento disparado quando um contato é adicionado à lista (listas estáticas)|is emitted when a contact is added to list|g' "$file"
    sed -i 's|evento disparado quando um contato é removido da lista (listas estáticas)|is emitted when a contact is removed from list|g' "$file"
    
    # Agent session events
    sed -i 's|é emitido quando um agente entra em uma sessão.|is emitted when an agent joins a session.|g' "$file"
    sed -i 's|é emitido quando um agente sai de uma sessão.|is emitted when an agent leaves a session.|g' "$file"
    sed -i 's|é emitido quando o papel do agente muda na sessão.|is emitted when agent role changes in session.|g' "$file"
    
    # Remove inline Portuguese comments
    sed -i 's| // ID da sessão||g' "$file"
    sed -i 's| // ID do contato||g' "$file"
    sed -i 's| // ID do tenant||g' "$file"
    sed -i 's| // ID do tipo de canal (opcional)||g' "$file"
    sed -i 's| // Momento em que iniciou||g' "$file"
    sed -i 's| // segundos||g' "$file"
    sed -i 's| // Lista de IDs das mensagens (ordenado por timestamp)||g' "$file"
    sed -i 's| // ID da primeira mensagem que iniciou a sessão||g' "$file"
    sed -i 's| // Resumo de eventos: {"message.created": 5, "tracking.captured": 1}||g' "$file"
    sed -i 's| // Métricas da sessão||g' "$file"
    sed -i 's| // Total de mensagens||g' "$file"
    sed -i 's| // Mensagens recebidas do contato||g' "$file"
    sed -i 's| // Mensagens enviadas pelo sistema/agente||g' "$file"
    sed -i 's| // Timestamp da primeira mensagem||g' "$file"
    sed -i 's| // Timestamp da última mensagem||g' "$file"
    
    # Remove section dividers
    sed -i '/^\/\/ ============/d' "$file"
    sed -i '/^\/\/ AI Processing Events - Disparam workflows do Temporal/d' "$file"
    sed -i '/^\/\/ Contexto completo da sessão (adicionado para webhook enrichment)/d' "$file"
    
    echo -e "${GREEN}✅ Cleaned${NC}"
    ((SUCCESS++))
}

# Files to clean
FILES=(
    "internal/domain/pipeline/events.go"
    "internal/domain/channel/events.go"
    "internal/domain/agent/events.go"
    "internal/domain/billing/events.go"
    "internal/domain/tracking/events.go"
    "internal/domain/note/events.go"
    "internal/domain/credential/events.go"
    "internal/domain/contact_list/events.go"
    "internal/domain/agent_session/events.go"
)

# Process each file
for file in "${FILES[@]}"; do
    ((TOTAL++))
    if [ -f "$file" ]; then
        clean_file "$file" || ((SKIP++))
    else
        echo -e "${RED}❌ Not found:${NC} $file"
        ((SKIP++))
    fi
    echo ""
done

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ Summary${NC}"
echo "Total files:    $TOTAL"
echo "Cleaned:        $SUCCESS"
echo "Skipped/Error:  $SKIP"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🔍 To verify: grep -r 'é emitido\|é disparado\|criado\|atualizado\|// segundos' internal/domain/*/events.go"
echo "♻️  Backups saved as: *.backup"
echo ""
echo -e "${GREEN}Done!${NC}"
