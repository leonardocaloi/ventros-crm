package waha

import (
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"go.uber.org/zap"
)

// IdentifierExtractor extrai e normaliza identificadores do WhatsApp de eventos WAHA
type IdentifierExtractor struct {
	logger *zap.Logger
}

// NewIdentifierExtractor cria um novo extrator de identificadores
func NewIdentifierExtractor(logger *zap.Logger) *IdentifierExtractor {
	return &IdentifierExtractor{
		logger: logger,
	}
}

// ExtractFromMessageEvent extrai identificadores do WhatsApp de um evento de mensagem
// Implementa lógica robusta de busca em múltiplos campos candidatos (inspirado no padrão JS)
func (e *IdentifierExtractor) ExtractFromMessageEvent(event WAHAMessageEvent) (*contact.WhatsAppIdentifiers, error) {
	// Lista de candidatos para buscar IDs (ordem de prioridade)
	candidates := []string{
		event.Payload.From,             // Campo principal (quem enviou)
		event.Payload.Data.Info.Chat,   // Info.Chat alternativo
		event.Payload.Data.Info.Sender, // Info.Sender alternativo
		event.Me.ID,                    // ID do "me" (pode ter formato diferente)
		event.Me.LID,                   // LID do "me"
		event.Me.JID,                   // JID do "me"
	}

	// Remove vazios e filtra candidatos válidos
	validCandidates := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c != "" {
			validCandidates = append(validCandidates, c)
		}
	}

	// Extração dos IDs por formato
	jid := e.findByFormat(validCandidates, "@s.whatsapp.net")
	lid := e.findByFormat(validCandidates, "@lid")

	// Para WHID, busca @c.us ou constrói a partir do JID
	whid := e.findByFormat(validCandidates, "@c.us")
	if whid == "" && jid != "" {
		// Constrói WHID a partir do JID (remove sufixo e adiciona @c.us)
		phoneNumber := e.extractPhoneNumber(jid)
		if phoneNumber != "" {
			whid = phoneNumber + "@c.us"
		}
	}

	// Se não encontrou WHID, usa o primeiro candidato (fallback)
	if whid == "" && len(validCandidates) > 0 {
		whid = validCandidates[0]
	}

	// Prepara ponteiros opcionais
	var lidPtr *string
	if lid != "" {
		lidPtr = &lid
	}

	var jidPtr *string
	if jid != "" {
		jidPtr = &jid
	}

	// Cria identificadores normalizados
	identifiers, err := contact.NewWhatsAppIdentifiers(whid, lidPtr, jidPtr)
	if err != nil {
		e.logger.Warn("Failed to create WhatsApp identifiers",
			zap.String("whid", whid),
			zap.Int("candidates_count", len(validCandidates)),
			zap.Error(err))
		return nil, err
	}

	// Log detalhado do que foi extraído
	idType := "UNKNOWN"
	if lid != "" && jid != "" {
		idType = "BOTH_AVAILABLE"
	} else if lid != "" {
		idType = "LID_ONLY"
	} else if jid != "" {
		idType = "PHONE_ONLY"
	}

	e.logger.Debug("Extracted WhatsApp identifiers",
		zap.String("wid", identifiers.WID()),
		zap.Bool("has_lid", identifiers.HasLID()),
		zap.Bool("has_jid", identifiers.HasJID()),
		zap.String("id_type", idType),
		zap.Int("candidates_count", len(validCandidates)))

	return identifiers, nil
}

// findByFormat busca o primeiro candidato que contém o formato especificado
func (e *IdentifierExtractor) findByFormat(candidates []string, format string) string {
	for _, candidate := range candidates {
		if candidate != "" && e.containsFormat(candidate, format) {
			return candidate
		}
	}
	return ""
}

// containsFormat verifica se o identificador contém o formato especificado
func (e *IdentifierExtractor) containsFormat(identifier, format string) bool {
	if identifier == "" || format == "" {
		return false
	}
	// Verifica se termina com o formato (mais preciso que contains)
	if len(identifier) >= len(format) {
		return identifier[len(identifier)-len(format):] == format
	}
	return false
}

// extractPhoneNumber extrai apenas o número do telefone (remove sufixos)
func (e *IdentifierExtractor) extractPhoneNumber(identifier string) string {
	if identifier == "" {
		return ""
	}

	// Remove todos os sufixos conhecidos
	suffixes := []string{"@c.us", "@s.whatsapp.net", "@lid", "@g.us"}
	result := identifier

	for _, suffix := range suffixes {
		if len(result) >= len(suffix) && result[len(result)-len(suffix):] == suffix {
			result = result[:len(result)-len(suffix)]
			break
		}
	}

	return result
}

// ExtractFromContact extrai identificadores de um WAHAContact (usado em listas de contatos)
func (e *IdentifierExtractor) ExtractFromContact(wahaContact WAHAContact) (*contact.WhatsAppIdentifiers, error) {
	// WID principal vem do ID do contato
	wid := wahaContact.ID

	// Não temos LID/JID neste contexto, podem ser nil
	identifiers, err := contact.NewWhatsAppIdentifiers(wid, nil, nil)
	if err != nil {
		e.logger.Warn("Failed to create WhatsApp identifiers from contact",
			zap.String("contact_id", wahaContact.ID),
			zap.Error(err))
		return nil, err
	}

	e.logger.Debug("Extracted WhatsApp identifiers from contact",
		zap.String("wid", identifiers.WID()))

	return identifiers, nil
}

// ToCustomFieldsMap converte identificadores para map de custom fields
// Retorna map[key]value pronto para ser salvo como custom fields do contato
func (e *IdentifierExtractor) ToCustomFieldsMap(identifiers *contact.WhatsAppIdentifiers) map[string]string {
	if identifiers == nil {
		return make(map[string]string)
	}

	return identifiers.ToCustomFields()
}
