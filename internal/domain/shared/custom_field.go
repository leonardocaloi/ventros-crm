package shared

import (
	"errors"
	"fmt"
	"time"
)

type FieldType string

const (
	FieldTypeText        FieldType = "text"
	FieldTypeNumber      FieldType = "number"
	FieldTypeBoolean     FieldType = "boolean"
	FieldTypeDate        FieldType = "date"
	FieldTypeJSON        FieldType = "json"
	FieldTypeURL         FieldType = "url"
	FieldTypeEmail       FieldType = "email"
	FieldTypePhone       FieldType = "phone"
	FieldTypeSelect      FieldType = "select"       // Seleção única
	FieldTypeMultiSelect FieldType = "multi_select" // Seleção múltipla
	FieldTypeLabel       FieldType = "label"        // Labels/Tags especiais
)

func (ft FieldType) IsValid() bool {
	switch ft {
	case FieldTypeText, FieldTypeNumber, FieldTypeBoolean,
		FieldTypeDate, FieldTypeJSON, FieldTypeURL,
		FieldTypeEmail, FieldTypePhone, FieldTypeSelect,
		FieldTypeMultiSelect, FieldTypeLabel:
		return true
	default:
		return false
	}
}

func (ft FieldType) String() string {
	return string(ft)
}

type CustomField struct {
	key       string
	fieldType FieldType
	value     interface{}
}

func NewCustomField(key string, fieldType FieldType, value interface{}) (*CustomField, error) {
	if key == "" {
		return nil, errors.New("field key cannot be empty")
	}
	if !fieldType.IsValid() {
		return nil, fmt.Errorf("invalid field type: %s", fieldType)
	}

	if err := validateFieldValue(fieldType, value); err != nil {
		return nil, err
	}

	return &CustomField{
		key:       key,
		fieldType: fieldType,
		value:     value,
	}, nil
}

func NewTextField(key, value string) (*CustomField, error) {
	return NewCustomField(key, FieldTypeText, value)
}

func NewNumberField(key string, value float64) (*CustomField, error) {
	return NewCustomField(key, FieldTypeNumber, value)
}

func NewBooleanField(key string, value bool) (*CustomField, error) {
	return NewCustomField(key, FieldTypeBoolean, value)
}

func NewDateField(key string, value time.Time) (*CustomField, error) {
	return NewCustomField(key, FieldTypeDate, value)
}

func NewJSONField(key string, value map[string]interface{}) (*CustomField, error) {
	return NewCustomField(key, FieldTypeJSON, value)
}

func NewLabelField(key string, labelIDs []string) (*CustomField, error) {
	if labelIDs == nil {
		labelIDs = []string{}
	}
	return NewCustomField(key, FieldTypeLabel, labelIDs)
}

func NewSelectField(key string, value string) (*CustomField, error) {
	return NewCustomField(key, FieldTypeSelect, value)
}

func NewMultiSelectField(key string, values []string) (*CustomField, error) {
	if values == nil {
		values = []string{}
	}
	return NewCustomField(key, FieldTypeMultiSelect, values)
}

func validateFieldValue(fieldType FieldType, value interface{}) error {
	if value == nil {
		return errors.New("field value cannot be nil")
	}

	switch fieldType {
	case FieldTypeText, FieldTypeURL, FieldTypeEmail, FieldTypePhone, FieldTypeSelect:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string for type %s, got %T", fieldType, value)
		}
	case FieldTypeNumber:
		switch value.(type) {
		case float64, float32, int, int32, int64:

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
	case FieldTypeLabel, FieldTypeMultiSelect:
		// Aceita []string ou []interface{}
		if _, ok := value.([]string); ok {
			return nil
		}
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if _, ok := item.(string); !ok {
					return fmt.Errorf("expected []string for type %s, but array contains %T", fieldType, item)
				}
			}
			return nil
		}
		return fmt.Errorf("expected []string for type %s, got %T", fieldType, value)
	}

	return nil
}

func (cf *CustomField) Key() string {
	return cf.key
}

func (cf *CustomField) Type() FieldType {
	return cf.fieldType
}

func (cf *CustomField) Value() interface{} {
	return cf.value
}

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

func (cf *CustomField) AsBoolean() (bool, error) {
	if cf.fieldType != FieldTypeBoolean {
		return false, fmt.Errorf("field type is %s, not boolean", cf.fieldType)
	}

	if b, ok := cf.value.(bool); ok {
		return b, nil
	}
	return false, errors.New("value is not a boolean")
}

func (cf *CustomField) AsDate() (time.Time, error) {
	if cf.fieldType != FieldTypeDate {
		return time.Time{}, fmt.Errorf("field type is %s, not date", cf.fieldType)
	}

	if t, ok := cf.value.(time.Time); ok {
		return t, nil
	}
	return time.Time{}, errors.New("value is not a time.Time")
}

func (cf *CustomField) AsJSON() (map[string]interface{}, error) {
	if cf.fieldType != FieldTypeJSON {
		return nil, fmt.Errorf("field type is %s, not json", cf.fieldType)
	}

	if m, ok := cf.value.(map[string]interface{}); ok {
		copy := make(map[string]interface{})
		for k, v := range m {
			copy[k] = v
		}
		return copy, nil
	}
	return nil, errors.New("value is not a map[string]interface{}")
}

func (cf *CustomField) Equals(other *CustomField) bool {
	if other == nil {
		return false
	}

	if cf.key != other.key || cf.fieldType != other.fieldType {
		return false
	}

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
		return false
	}

	return false
}

func (cf *CustomField) String() string {
	return fmt.Sprintf("%s (%s): %v", cf.key, cf.fieldType, cf.value)
}

// AsStringSlice retorna o valor como []string (para labels, multi_select)
func (cf *CustomField) AsStringSlice() ([]string, error) {
	switch cf.fieldType {
	case FieldTypeLabel, FieldTypeMultiSelect:
		if cf.value == nil {
			return []string{}, nil
		}

		// Tenta converter de []string
		if slice, ok := cf.value.([]string); ok {
			return slice, nil
		}

		// Tenta converter de []interface{}
		if arr, ok := cf.value.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				} else {
					return nil, fmt.Errorf("custom field contains non-string values")
				}
			}
			return result, nil
		}

		return nil, errors.New("value is not a string slice")

	default:
		return nil, fmt.Errorf("field type is %s, not label or multi_select", cf.fieldType)
	}
}
