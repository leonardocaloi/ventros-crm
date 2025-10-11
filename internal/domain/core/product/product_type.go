package product

import "errors"

// ProductType represents available products in Ventros platform
type ProductType string

const (
	// ProductTypeCRM is the Customer Relationship Management product
	ProductTypeCRM ProductType = "crm"

	// ProductTypeBI is the Business Intelligence product
	ProductTypeBI ProductType = "bi"

	// ProductTypeAutomation is the Automation/Integration product
	ProductTypeAutomation ProductType = "automation"
)

// All returns all valid product types
func All() []ProductType {
	return []ProductType{
		ProductTypeCRM,
		ProductTypeBI,
		ProductTypeAutomation,
	}
}

// IsValid checks if product type is valid
func (p ProductType) IsValid() bool {
	switch p {
	case ProductTypeCRM, ProductTypeBI, ProductTypeAutomation:
		return true
	default:
		return false
	}
}

// String returns string representation
func (p ProductType) String() string {
	return string(p)
}

// ParseProductType parses string to ProductType
func ParseProductType(s string) (ProductType, error) {
	p := ProductType(s)
	if !p.IsValid() {
		return "", errors.New("invalid product type")
	}
	return p, nil
}
