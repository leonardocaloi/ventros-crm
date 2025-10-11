package shared

// CustomFieldType representa os tipos de custom fields suportados
type CustomFieldType string

const (
	// Tipos básicos
	CustomFieldTypeText        CustomFieldType = "text"
	CustomFieldTypeNumber      CustomFieldType = "number"
	CustomFieldTypeDate        CustomFieldType = "date"
	CustomFieldTypeBoolean     CustomFieldType = "boolean"
	CustomFieldTypeSelect      CustomFieldType = "select"
	CustomFieldTypeMultiSelect CustomFieldType = "multi_select"

	// Tipo especial para labels/tags
	CustomFieldTypeLabel CustomFieldType = "label"

	// Tipos avançados
	CustomFieldTypeURL   CustomFieldType = "url"
	CustomFieldTypeEmail CustomFieldType = "email"
	CustomFieldTypePhone CustomFieldType = "phone"
	CustomFieldTypeJSON  CustomFieldType = "json"
)

// IsValid verifica se o tipo de custom field é válido
func (t CustomFieldType) IsValid() bool {
	validTypes := []CustomFieldType{
		CustomFieldTypeText,
		CustomFieldTypeNumber,
		CustomFieldTypeDate,
		CustomFieldTypeBoolean,
		CustomFieldTypeSelect,
		CustomFieldTypeMultiSelect,
		CustomFieldTypeLabel,
		CustomFieldTypeURL,
		CustomFieldTypeEmail,
		CustomFieldTypePhone,
		CustomFieldTypeJSON,
	}

	for _, valid := range validTypes {
		if t == valid {
			return true
		}
	}

	return false
}

// String retorna a representação em string do tipo
func (t CustomFieldType) String() string {
	return string(t)
}

// AcceptsValue verifica se o tipo aceita um determinado valor
func (t CustomFieldType) AcceptsValue(value interface{}) bool {
	if value == nil {
		return true // nil é válido para campos opcionais
	}

	switch t {
	case CustomFieldTypeText, CustomFieldTypeURL, CustomFieldTypeEmail, CustomFieldTypePhone:
		_, ok := value.(string)
		return ok

	case CustomFieldTypeNumber:
		switch value.(type) {
		case int, int64, float64:
			return true
		default:
			return false
		}

	case CustomFieldTypeBoolean:
		_, ok := value.(bool)
		return ok

	case CustomFieldTypeDate:
		_, ok := value.(string) // ISO 8601 format
		return ok

	case CustomFieldTypeSelect:
		_, ok := value.(string)
		return ok

	case CustomFieldTypeMultiSelect, CustomFieldTypeLabel:
		_, ok := value.([]string)
		if !ok {
			// Tenta converter de []interface{} para []string
			if arr, ok := value.([]interface{}); ok {
				for _, item := range arr {
					if _, ok := item.(string); !ok {
						return false
					}
				}
				return true
			}
			return false
		}
		return true

	case CustomFieldTypeJSON:
		switch value.(type) {
		case map[string]interface{}, []interface{}:
			return true
		default:
			return false
		}

	default:
		return false
	}
}

// DefaultValue retorna o valor padrão para o tipo
func (t CustomFieldType) DefaultValue() interface{} {
	switch t {
	case CustomFieldTypeText, CustomFieldTypeURL, CustomFieldTypeEmail, CustomFieldTypePhone, CustomFieldTypeSelect:
		return ""
	case CustomFieldTypeNumber:
		return 0
	case CustomFieldTypeBoolean:
		return false
	case CustomFieldTypeDate:
		return ""
	case CustomFieldTypeMultiSelect, CustomFieldTypeLabel:
		return []string{}
	case CustomFieldTypeJSON:
		return map[string]interface{}{}
	default:
		return nil
	}
}
