package shared

import (
	"errors"
	"fmt"
	"time"
)

// FieldType representa os tipos suportados de campos customizados.
type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeDate    FieldType = "date"
	FieldTypeJSON    FieldType = "json"
	FieldTypeURL     FieldType = "url"
	FieldTypeEmail   FieldType = "email"
	FieldTypePhone   FieldType = "phone"
)

// IsValid verifica se o tipo de campo é válido.
func (ft FieldType) IsValid() bool {
	switch ft {
	case FieldTypeText, FieldTypeNumber, FieldTypeBoolean,
		FieldTypeDate, FieldTypeJSON, FieldTypeURL,
		FieldTypeEmail, FieldTypePhone:
		return true
	default:
		return false
	}
}

// String retorna a representação em string do tipo.
func (ft FieldType) String() string {
	return string(ft)
}

// CustomField representa um campo customizado com tipo e valor.
// É um Value Object imutável.
type CustomField struct {
	key       string
	fieldType FieldType
	value     interface{}
}

// NewCustomField cria um novo campo customizado.
func NewCustomField(key string, fieldType FieldType, value interface{}) (*CustomField, error) {
	if key == "" {
		return nil, errors.New("field key cannot be empty")
	}
	if !fieldType.IsValid() {
		return nil, fmt.Errorf("invalid field type: %s", fieldType)
	}

	// Validar tipo do valor
	if err := validateFieldValue(fieldType, value); err != nil {
		return nil, err
	}

	return &CustomField{
		key:       key,
		fieldType: fieldType,
		value:     value,
	}, nil
}

// NewTextField cria um campo de texto.
func NewTextField(key, value string) (*CustomField, error) {
	return NewCustomField(key, FieldTypeText, value)
}

// NewNumberField cria um campo numérico.
func NewNumberField(key string, value float64) (*CustomField, error) {
	return NewCustomField(key, FieldTypeNumber, value)
}

// NewBooleanField cria um campo booleano.
func NewBooleanField(key string, value bool) (*CustomField, error) {
	return NewCustomField(key, FieldTypeBoolean, value)
}

// NewDateField cria um campo de data.
func NewDateField(key string, value time.Time) (*CustomField, error) {
	return NewCustomField(key, FieldTypeDate, value)
}

// NewJSONField cria um campo JSON.
func NewJSONField(key string, value map[string]interface{}) (*CustomField, error) {
	return NewCustomField(key, FieldTypeJSON, value)
}

// validateFieldValue valida se o valor corresponde ao tipo declarado.
func validateFieldValue(fieldType FieldType, value interface{}) error {
	if value == nil {
		return errors.New("field value cannot be nil")
	}

	switch fieldType {
	case FieldTypeText, FieldTypeURL, FieldTypeEmail, FieldTypePhone:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string for type %s, got %T", fieldType, value)
		}
	case FieldTypeNumber:
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// OK
		default:
			return fmt.Errorf("expected number for type %s, got %T", fieldType, value)
		}
	case FieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool for type %s, got %T", fieldType, value)
		}
	case FieldTypeDate:
		if _, ok := value.(time.Time); !ok {
			return fmt.Errorf("expected time.Time for type %s, got %T", fieldType, value)
		}
	case FieldTypeJSON:
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected map[string]interface{} for type %s, got %T", fieldType, value)
		}
	}

	return nil
}

// Key retorna a chave do campo.
func (cf *CustomField) Key() string {
	return cf.key
}

// Type retorna o tipo do campo.
func (cf *CustomField) Type() FieldType {
	return cf.fieldType
}

// Value retorna o valor do campo.
func (cf *CustomField) Value() interface{} {
	return cf.value
}

// AsText retorna o valor como string (se for do tipo text/url/email/phone).
func (cf *CustomField) AsText() (string, error) {
	switch cf.fieldType {
	case FieldTypeText, FieldTypeURL, FieldTypeEmail, FieldTypePhone:
		if str, ok := cf.value.(string); ok {
			return str, nil
		}
		return "", errors.New("value is not a string")
	default:
		return "", fmt.Errorf("field type is %s, not text-based", cf.fieldType)
	}
}

// AsNumber retorna o valor como float64 (se for do tipo number).
func (cf *CustomField) AsNumber() (float64, error) {
	if cf.fieldType != FieldTypeNumber {
		return 0, fmt.Errorf("field type is %s, not number", cf.fieldType)
	}

	switch v := cf.value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, errors.New("value is not a number")
	}
}

// AsBoolean retorna o valor como bool (se for do tipo boolean).
func (cf *CustomField) AsBoolean() (bool, error) {
	if cf.fieldType != FieldTypeBoolean {
		return false, fmt.Errorf("field type is %s, not boolean", cf.fieldType)
	}

	if b, ok := cf.value.(bool); ok {
		return b, nil
	}
	return false, errors.New("value is not a boolean")
}

// AsDate retorna o valor como time.Time (se for do tipo date).
func (cf *CustomField) AsDate() (time.Time, error) {
	if cf.fieldType != FieldTypeDate {
		return time.Time{}, fmt.Errorf("field type is %s, not date", cf.fieldType)
	}

	if t, ok := cf.value.(time.Time); ok {
		return t, nil
	}
	return time.Time{}, errors.New("value is not a time.Time")
}

// AsJSON retorna o valor como map[string]interface{} (se for do tipo json).
func (cf *CustomField) AsJSON() (map[string]interface{}, error) {
	if cf.fieldType != FieldTypeJSON {
		return nil, fmt.Errorf("field type is %s, not json", cf.fieldType)
	}

	if m, ok := cf.value.(map[string]interface{}); ok {
		// Return copy
		copy := make(map[string]interface{})
		for k, v := range m {
			copy[k] = v
		}
		return copy, nil
	}
	return nil, errors.New("value is not a map[string]interface{}")
}

// Equals compara dois CustomFields por valor.
func (cf *CustomField) Equals(other *CustomField) bool {
	if other == nil {
		return false
	}

	if cf.key != other.key || cf.fieldType != other.fieldType {
		return false
	}

	// Comparação de valores depende do tipo
	switch cf.fieldType {
	case FieldTypeText, FieldTypeURL, FieldTypeEmail, FieldTypePhone:
		v1, _ := cf.AsText()
		v2, _ := other.AsText()
		return v1 == v2
	case FieldTypeNumber:
		v1, _ := cf.AsNumber()
		v2, _ := other.AsNumber()
		return v1 == v2
	case FieldTypeBoolean:
		v1, _ := cf.AsBoolean()
		v2, _ := other.AsBoolean()
		return v1 == v2
	case FieldTypeDate:
		v1, _ := cf.AsDate()
		v2, _ := other.AsDate()
		return v1.Equal(v2)
	case FieldTypeJSON:
		// JSON comparison is complex, just compare pointer for now
		return false
	}

	return false
}

// String retorna uma representação em string do campo.
func (cf *CustomField) String() string {
	return fmt.Sprintf("%s (%s): %v", cf.key, cf.fieldType, cf.value)
}
