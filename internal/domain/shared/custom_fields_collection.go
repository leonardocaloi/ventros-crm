package shared

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CustomFieldsCollection manages a collection of custom fields
type CustomFieldsCollection struct {
	fields map[string]*CustomField
}

// NewCustomFieldsCollection creates a new empty collection
func NewCustomFieldsCollection() *CustomFieldsCollection {
	return &CustomFieldsCollection{
		fields: make(map[string]*CustomField),
	}
}

// Add adds or updates a custom field in the collection
func (c *CustomFieldsCollection) Add(field *CustomField) error {
	if field == nil {
		return ErrCustomFieldNil
	}
	c.fields[field.Key()] = field
	return nil
}

// Get retrieves a custom field by key
func (c *CustomFieldsCollection) Get(key string) (*CustomField, bool) {
	field, exists := c.fields[key]
	return field, exists
}

// Remove removes a custom field from the collection
func (c *CustomFieldsCollection) Remove(key string) error {
	if _, exists := c.fields[key]; !exists {
		return ErrCustomFieldNotFound
	}
	delete(c.fields, key)
	return nil
}

// Has checks if a custom field exists
func (c *CustomFieldsCollection) Has(key string) bool {
	_, exists := c.fields[key]
	return exists
}

// All returns all custom fields
func (c *CustomFieldsCollection) All() []*CustomField {
	fields := make([]*CustomField, 0, len(c.fields))
	for _, field := range c.fields {
		fields = append(fields, field)
	}
	return fields
}

// Count returns the number of custom fields
func (c *CustomFieldsCollection) Count() int {
	return len(c.fields)
}

// GetByType returns all custom fields of a specific type
func (c *CustomFieldsCollection) GetByType(fieldType FieldType) []*CustomField {
	fields := make([]*CustomField, 0)
	for _, field := range c.fields {
		if field.Type() == fieldType {
			fields = append(fields, field)
		}
	}
	return fields
}

// ToMap converts the collection to a map[string]interface{} for JSON serialization
func (c *CustomFieldsCollection) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for key, field := range c.fields {
		result[key] = map[string]interface{}{
			"type":  field.Type().String(),
			"value": field.Value(),
		}
	}
	return result
}

// FromMap creates a collection from a map[string]interface{}
func FromMap(data map[string]interface{}) (*CustomFieldsCollection, error) {
	collection := NewCustomFieldsCollection()

	for key, value := range data {
		// Handle different formats
		var fieldType FieldType
		var fieldValue interface{}

		switch v := value.(type) {
		case map[string]interface{}:
			// Format: {"type": "text", "value": "..."}
			typeStr, ok := v["type"].(string)
			if !ok {
				return nil, fmt.Errorf("missing or invalid type for field %s", key)
			}
			fieldType = FieldType(typeStr)
			fieldValue = v["value"]
		default:
			// Simple format: just the value, infer type
			fieldType, fieldValue = inferTypeAndValue(v)
		}

		field, err := NewCustomField(key, fieldType, fieldValue)
		if err != nil {
			return nil, fmt.Errorf("failed to create field %s: %w", key, err)
		}

		if err := collection.Add(field); err != nil {
			return nil, err
		}
	}

	return collection, nil
}

// inferTypeAndValue infers the field type from the value
func inferTypeAndValue(value interface{}) (FieldType, interface{}) {
	switch v := value.(type) {
	case string:
		return FieldTypeText, v
	case float64, int, int64:
		return FieldTypeNumber, v
	case bool:
		return FieldTypeBoolean, v
	case []string:
		return FieldTypeMultiSelect, v
	case []interface{}:
		// Try to convert to []string
		strSlice := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				strSlice = append(strSlice, str)
			}
		}
		return FieldTypeMultiSelect, strSlice
	case map[string]interface{}:
		return FieldTypeJSON, v
	default:
		return FieldTypeText, fmt.Sprintf("%v", v)
	}
}

// MarshalJSON implements json.Marshaler
func (c *CustomFieldsCollection) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.ToMap())
}

// UnmarshalJSON implements json.Unmarshaler
func (c *CustomFieldsCollection) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	collection, err := FromMap(m)
	if err != nil {
		return err
	}

	c.fields = collection.fields
	return nil
}

// Merge merges another collection into this one, overwriting existing fields
func (c *CustomFieldsCollection) Merge(other *CustomFieldsCollection) {
	if other == nil {
		return
	}

	for key, field := range other.fields {
		c.fields[key] = field
	}
}

// Clone creates a deep copy of the collection
func (c *CustomFieldsCollection) Clone() *CustomFieldsCollection {
	clone := NewCustomFieldsCollection()
	for key, field := range c.fields {
		clonedField, _ := NewCustomField(field.Key(), field.Type(), field.Value())
		clone.fields[key] = clonedField
	}
	return clone
}

// Validate validates all custom fields against their definitions
func (c *CustomFieldsCollection) Validate(definitions []*CustomFieldDefinition) error {
	// Check required fields
	for _, def := range definitions {
		if def.Required {
			if !c.Has(def.Key) {
				return fmt.Errorf("required custom field missing: %s", def.Key)
			}
		}
	}

	// Validate field types and values
	for key, field := range c.fields {
		def := findDefinition(key, definitions)
		if def != nil {
			if field.Type() != def.Type {
				return fmt.Errorf("custom field %s has wrong type: expected %s, got %s",
					key, def.Type, field.Type())
			}
		}
	}

	return nil
}

// CustomFieldDefinition defines metadata for a custom field
type CustomFieldDefinition struct {
	Key          string
	Type         FieldType
	Required     bool
	Fixed        bool   // Cannot be removed
	System       bool   // System-managed field
	Description  string
	DefaultValue interface{}
	Options      []string // For select/multi_select types
}

func findDefinition(key string, definitions []*CustomFieldDefinition) *CustomFieldDefinition {
	for _, def := range definitions {
		if def.Key == key {
			return def
		}
	}
	return nil
}

// SetValue sets the value of a custom field by key, creating it if it doesn't exist
func (c *CustomFieldsCollection) SetValue(key string, fieldType FieldType, value interface{}) error {
	field, err := NewCustomField(key, fieldType, value)
	if err != nil {
		return err
	}
	return c.Add(field)
}

// GetValue gets the value of a custom field by key
func (c *CustomFieldsCollection) GetValue(key string) (interface{}, error) {
	field, exists := c.Get(key)
	if !exists {
		return nil, errors.New("custom field not found: " + key)
	}
	return field.Value(), nil
}

// GetStringValue gets a string value from a custom field
func (c *CustomFieldsCollection) GetStringValue(key string) (string, error) {
	field, exists := c.Get(key)
	if !exists {
		return "", errors.New("custom field not found: " + key)
	}
	return field.AsText()
}

// GetNumberValue gets a number value from a custom field
func (c *CustomFieldsCollection) GetNumberValue(key string) (float64, error) {
	field, exists := c.Get(key)
	if !exists {
		return 0, errors.New("custom field not found: " + key)
	}
	return field.AsNumber()
}

// GetBooleanValue gets a boolean value from a custom field
func (c *CustomFieldsCollection) GetBooleanValue(key string) (bool, error) {
	field, exists := c.Get(key)
	if !exists {
		return false, errors.New("custom field not found: " + key)
	}
	return field.AsBoolean()
}

// GetStringSliceValue gets a string slice value from a custom field (for labels, multi_select)
func (c *CustomFieldsCollection) GetStringSliceValue(key string) ([]string, error) {
	field, exists := c.Get(key)
	if !exists {
		return nil, errors.New("custom field not found: " + key)
	}
	return field.AsStringSlice()
}
