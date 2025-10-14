package contact_list

import (
	"github.com/ventros/crm/internal/domain/core/shared"
)

func parseFieldType(fieldTypeStr string) shared.FieldType {
	return shared.FieldType(fieldTypeStr)
}
