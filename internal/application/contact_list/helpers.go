package contact_list

import (
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

func parseFieldType(fieldTypeStr string) shared.FieldType {
	return shared.FieldType(fieldTypeStr)
}
