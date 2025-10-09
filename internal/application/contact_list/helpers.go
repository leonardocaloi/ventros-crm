package contact_list

import (
	"github.com/caloi/ventros-crm/internal/domain/shared"
)

func parseFieldType(fieldTypeStr string) shared.FieldType {
	return shared.FieldType(fieldTypeStr)
}
