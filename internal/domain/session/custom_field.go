package session

import (
	"errors"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// SessionCustomField é uma entidade que persiste campos customizados da sessão.
// É uma entidade (não Value Object) porque tem ID próprio e ciclo de vida independente.
type SessionCustomField struct {
	id          uuid.UUID
	sessionID   uuid.UUID
	tenantID    string
	customField *shared.CustomField // Value Object encapsulado
	createdAt   time.Time
	updatedAt   time.Time
}

// NewSessionCustomField cria um novo campo customizado para uma sessão.
func NewSessionCustomField(
	sessionID uuid.UUID,
	tenantID string,
	customField *shared.CustomField,
) (*SessionCustomField, error) {
	if sessionID == uuid.Nil {
		return nil, errors.New("sessionID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if customField == nil {
		return nil, errors.New("customField cannot be nil")
	}

	now := time.Now()
	return &SessionCustomField{
		id:          uuid.New(),
		sessionID:   sessionID,
		tenantID:    tenantID,
		customField: customField,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructSessionCustomField reconstrói a partir de dados persistidos.
func ReconstructSessionCustomField(
	id uuid.UUID,
	sessionID uuid.UUID,
	tenantID string,
	customField *shared.CustomField,
	createdAt time.Time,
	updatedAt time.Time,
) *SessionCustomField {
	return &SessionCustomField{
		id:          id,
		sessionID:   sessionID,
		tenantID:    tenantID,
		customField: customField,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// UpdateValue atualiza o valor do campo.
func (scf *SessionCustomField) UpdateValue(newField *shared.CustomField) error {
	if newField == nil {
		return errors.New("new field cannot be nil")
	}
	
	// Validar que mantém a mesma chave e tipo
	if scf.customField.Key() != newField.Key() {
		return errors.New("cannot change field key")
	}
	if scf.customField.Type() != newField.Type() {
		return errors.New("cannot change field type")
	}

	scf.customField = newField
	scf.updatedAt = time.Now()
	return nil
}

// Getters
func (scf *SessionCustomField) ID() uuid.UUID                 { return scf.id }
func (scf *SessionCustomField) SessionID() uuid.UUID          { return scf.sessionID }
func (scf *SessionCustomField) TenantID() string              { return scf.tenantID }
func (scf *SessionCustomField) CustomField() *shared.CustomField { return scf.customField }
func (scf *SessionCustomField) FieldKey() string              { return scf.customField.Key() }
func (scf *SessionCustomField) FieldType() shared.FieldType   { return scf.customField.Type() }
func (scf *SessionCustomField) FieldValue() interface{}       { return scf.customField.Value() }
func (scf *SessionCustomField) CreatedAt() time.Time          { return scf.createdAt }
func (scf *SessionCustomField) UpdatedAt() time.Time          { return scf.updatedAt }
