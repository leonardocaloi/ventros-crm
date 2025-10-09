package contact

import (
	"errors"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// ContactCustomField é uma entidade que persiste campos customizados do contato.
// É uma entidade (não Value Object) porque tem ID próprio e ciclo de vida independente.
type ContactCustomField struct {
	id          uuid.UUID
	contactID   uuid.UUID
	tenantID    string
	customField *shared.CustomField // Value Object encapsulado
	createdAt   time.Time
	updatedAt   time.Time
}

// NewContactCustomField cria um novo campo customizado para um contato.
func NewContactCustomField(
	contactID uuid.UUID,
	tenantID string,
	customField *shared.CustomField,
) (*ContactCustomField, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if customField == nil {
		return nil, errors.New("customField cannot be nil")
	}

	now := time.Now()
	return &ContactCustomField{
		id:          uuid.New(),
		contactID:   contactID,
		tenantID:    tenantID,
		customField: customField,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructContactCustomField reconstrói a partir de dados persistidos.
func ReconstructContactCustomField(
	id uuid.UUID,
	contactID uuid.UUID,
	tenantID string,
	customField *shared.CustomField,
	createdAt time.Time,
	updatedAt time.Time,
) *ContactCustomField {
	return &ContactCustomField{
		id:          id,
		contactID:   contactID,
		tenantID:    tenantID,
		customField: customField,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// UpdateValue atualiza o valor do campo.
func (ccf *ContactCustomField) UpdateValue(newField *shared.CustomField) error {
	if newField == nil {
		return errors.New("new field cannot be nil")
	}

	// Validar que mantém a mesma chave e tipo
	if ccf.customField.Key() != newField.Key() {
		return errors.New("cannot change field key")
	}
	if ccf.customField.Type() != newField.Type() {
		return errors.New("cannot change field type")
	}

	ccf.customField = newField
	ccf.updatedAt = time.Now()
	return nil
}

// Getters
func (ccf *ContactCustomField) ID() uuid.UUID                    { return ccf.id }
func (ccf *ContactCustomField) ContactID() uuid.UUID             { return ccf.contactID }
func (ccf *ContactCustomField) TenantID() string                 { return ccf.tenantID }
func (ccf *ContactCustomField) CustomField() *shared.CustomField { return ccf.customField }
func (ccf *ContactCustomField) FieldKey() string                 { return ccf.customField.Key() }
func (ccf *ContactCustomField) FieldType() shared.FieldType      { return ccf.customField.Type() }
func (ccf *ContactCustomField) FieldValue() interface{}          { return ccf.customField.Value() }
func (ccf *ContactCustomField) CreatedAt() time.Time             { return ccf.createdAt }
func (ccf *ContactCustomField) UpdatedAt() time.Time             { return ccf.updatedAt }
