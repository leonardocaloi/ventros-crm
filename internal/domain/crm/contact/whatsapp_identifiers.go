package contact

import (
	"errors"
	"strings"
)

// WhatsAppIdentifiers representa os identificadores normalizados do WhatsApp/WAHA
//
// WID (WhatsApp ID): ID principal do contato no WhatsApp (ex: 5511999999999@c.us)
// LID (Local ID): ID alternativo/local quando disponível
// JID (Jabber ID): ID do protocolo Jabber/XMPP do WhatsApp
type WhatsAppIdentifiers struct {
	wid string  // WhatsApp ID (obrigatório)
	lid *string // Local ID (opcional)
	jid *string // Jabber ID (opcional)
}

var (
	ErrEmptyWID         = errors.New("WID cannot be empty")
	ErrInvalidWID       = errors.New("WID format is invalid")
	ErrInvalidLID       = errors.New("LID format is invalid")
	ErrInvalidJID       = errors.New("JID format is invalid")
	ErrWIDNotNormalized = errors.New("WID must be normalized (no suffixes)")
)

// NewWhatsAppIdentifiers cria identificadores do WhatsApp com validação
func NewWhatsAppIdentifiers(wid string, lid, jid *string) (*WhatsAppIdentifiers, error) {
	// WID é obrigatório
	if wid == "" {
		return nil, ErrEmptyWID
	}

	// Normaliza e valida WID
	normalizedWID, err := NormalizeWhatsAppID(wid)
	if err != nil {
		return nil, err
	}

	// Valida LID se fornecido
	var normalizedLID *string
	if lid != nil && *lid != "" {
		normalized, err := NormalizeWhatsAppID(*lid)
		if err != nil {
			return nil, err
		}
		normalizedLID = &normalized
	}

	// Valida JID se fornecido
	var normalizedJID *string
	if jid != nil && *jid != "" {
		normalized, err := NormalizeWhatsAppID(*jid)
		if err != nil {
			return nil, err
		}
		normalizedJID = &normalized
	}

	return &WhatsAppIdentifiers{
		wid: normalizedWID,
		lid: normalizedLID,
		jid: normalizedJID,
	}, nil
}

// NormalizeWhatsAppID remove sufixos do WhatsApp e normaliza o ID
//
// Sufixos conhecidos:
// - @c.us (contatos individuais)
// - @s.whatsapp.net (formato alternativo)
// - @lid (Local ID)
// - @g.us (grupos - não usado para contatos)
func NormalizeWhatsAppID(id string) (string, error) {
	if id == "" {
		return "", errors.New("ID cannot be empty")
	}

	// Remove espaços
	id = strings.TrimSpace(id)

	// Lista de sufixos conhecidos do WhatsApp
	suffixes := []string{
		"@c.us",
		"@s.whatsapp.net",
		"@lid",
		"@g.us",
	}

	// Remove sufixo se encontrado
	normalized := id
	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			break
		}
	}

	// Valida que o ID normalizado contém apenas números
	if !isNumericString(normalized) {
		return "", errors.New("normalized ID must contain only numbers")
	}

	return normalized, nil
}

// isNumericString verifica se a string contém apenas dígitos
func isNumericString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// WID retorna o WhatsApp ID normalizado
func (w *WhatsAppIdentifiers) WID() string {
	return w.wid
}

// LID retorna o Local ID se disponível
func (w *WhatsAppIdentifiers) LID() *string {
	return w.lid
}

// JID retorna o Jabber ID se disponível
func (w *WhatsAppIdentifiers) JID() *string {
	return w.jid
}

// HasLID verifica se possui Local ID
func (w *WhatsAppIdentifiers) HasLID() bool {
	return w.lid != nil && *w.lid != ""
}

// HasJID verifica se possui Jabber ID
func (w *WhatsAppIdentifiers) HasJID() bool {
	return w.jid != nil && *w.jid != ""
}

// ToCustomFields converte os identificadores para custom fields do domínio
func (w *WhatsAppIdentifiers) ToCustomFields() map[string]string {
	fields := map[string]string{
		"waha_wid": w.wid, // Campo obrigatório
	}

	if w.HasLID() {
		fields["waha_lid"] = *w.lid
	}

	if w.HasJID() {
		fields["waha_jid"] = *w.jid
	}

	return fields
}

// Equals compara dois WhatsAppIdentifiers
func (w *WhatsAppIdentifiers) Equals(other *WhatsAppIdentifiers) bool {
	if other == nil {
		return false
	}

	if w.wid != other.wid {
		return false
	}

	// Compara LID
	if w.HasLID() != other.HasLID() {
		return false
	}
	if w.HasLID() && *w.lid != *other.lid {
		return false
	}

	// Compara JID
	if w.HasJID() != other.HasJID() {
		return false
	}
	if w.HasJID() && *w.jid != *other.jid {
		return false
	}

	return true
}

// String retorna representação em string
func (w *WhatsAppIdentifiers) String() string {
	s := "WID: " + w.wid
	if w.HasLID() {
		s += ", LID: " + *w.lid
	}
	if w.HasJID() {
		s += ", JID: " + *w.jid
	}
	return s
}
