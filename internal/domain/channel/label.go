package channel

import (
	"fmt"
	"regexp"
)

// Label represents a WhatsApp label/tag for chat organization
// Labels are managed at the channel level and can be applied to chats
type Label struct {
	ID       string `json:"id"`        // Label ID from WAHA
	Name     string `json:"name"`      // Label name (e.g., "VIP", "Support", "Sales")
	Color    int    `json:"color"`     // Color as number (0-19 on WhatsApp)
	ColorHex string `json:"colorHex"`  // Color as hex string (e.g., "#FF5733")
}

// NewLabel creates a new label with validation
func NewLabel(id, name string, color int, colorHex string) (*Label, error) {
	if id == "" {
		return nil, fmt.Errorf("label ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("label name is required")
	}
	if color < 0 || color > 19 {
		return nil, fmt.Errorf("label color must be between 0 and 19")
	}
	if !isValidHexColor(colorHex) {
		return nil, fmt.Errorf("invalid hex color format: %s", colorHex)
	}

	return &Label{
		ID:       id,
		Name:     name,
		Color:    color,
		ColorHex: colorHex,
	}, nil
}

// isValidHexColor validates hex color format (#RRGGBB or #RGB)
func isValidHexColor(hex string) bool {
	if hex == "" {
		return false
	}
	match, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, hex)
	return match
}

// LabelCollection manages a collection of labels for a channel
type LabelCollection struct {
	labels map[string]*Label // Map of label ID to Label
}

// NewLabelCollection creates a new empty label collection
func NewLabelCollection() *LabelCollection {
	return &LabelCollection{
		labels: make(map[string]*Label),
	}
}

// ReconstructLabelCollection reconstructs a label collection from persistence
func ReconstructLabelCollection(labels []*Label) *LabelCollection {
	collection := NewLabelCollection()
	for _, label := range labels {
		collection.labels[label.ID] = label
	}
	return collection
}

// Add adds or updates a label in the collection
func (lc *LabelCollection) Add(label *Label) {
	lc.labels[label.ID] = label
}

// Remove removes a label from the collection
func (lc *LabelCollection) Remove(labelID string) {
	delete(lc.labels, labelID)
}

// Get retrieves a label by ID
func (lc *LabelCollection) Get(labelID string) (*Label, bool) {
	label, exists := lc.labels[labelID]
	return label, exists
}

// GetByName retrieves a label by name
func (lc *LabelCollection) GetByName(name string) (*Label, bool) {
	for _, label := range lc.labels {
		if label.Name == name {
			return label, true
		}
	}
	return nil, false
}

// Has checks if a label exists
func (lc *LabelCollection) Has(labelID string) bool {
	_, exists := lc.labels[labelID]
	return exists
}

// All returns all labels
func (lc *LabelCollection) All() []*Label {
	labels := make([]*Label, 0, len(lc.labels))
	for _, label := range lc.labels {
		labels = append(labels, label)
	}
	return labels
}

// Count returns the number of labels
func (lc *LabelCollection) Count() int {
	return len(lc.labels)
}

// Clear removes all labels
func (lc *LabelCollection) Clear() {
	lc.labels = make(map[string]*Label)
}

// ToMap returns the internal map (for serialization)
func (lc *LabelCollection) ToMap() map[string]*Label {
	// Return copy to prevent external modification
	copy := make(map[string]*Label, len(lc.labels))
	for k, v := range lc.labels {
		copy[k] = v
	}
	return copy
}

// ToSlice returns labels as a slice (for serialization)
func (lc *LabelCollection) ToSlice() []*Label {
	return lc.All()
}
