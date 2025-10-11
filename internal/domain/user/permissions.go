package user

type ResourceType string

const (
	ResourceProject   ResourceType = "project"
	ResourceContact   ResourceType = "contact"
	ResourceMessage   ResourceType = "message"
	ResourceWebhook   ResourceType = "webhook"
	ResourcePipeline  ResourceType = "pipeline"
	ResourceUser      ResourceType = "user"
	ResourceAnalytics ResourceType = "analytics"
)

type Operation string

const (
	OperationCreate Operation = "create"
	OperationRead   Operation = "read"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
	OperationList   Operation = "list"
	OperationExport Operation = "export"
)

type Permission struct {
	Resource  ResourceType
	Operation Operation
}

func NewPermission(resource ResourceType, operation Operation) Permission {
	return Permission{
		Resource:  resource,
		Operation: operation,
	}
}

func (p Permission) IsManagerAllowed() bool {
	switch p.Resource {
	case ResourceUser:

		return p.Operation == OperationRead || p.Operation == OperationList
	case ResourceAnalytics:

		return p.Operation == OperationRead || p.Operation == OperationList
	default:

		return true
	}
}

func (p Permission) IsUserAllowed() bool {
	switch p.Resource {
	case ResourceUser:

		return p.Operation == OperationRead || p.Operation == OperationUpdate
	case ResourceAnalytics:

		return p.Operation == OperationRead
	default:

		return true
	}
}

func (p Permission) IsReadOnlyAllowed() bool {

	return p.Operation == OperationRead || p.Operation == OperationList
}

func (p Permission) String() string {
	return string(p.Resource) + ":" + string(p.Operation)
}

var PermissionMatrix = map[Role]map[ResourceType][]Operation{
	RoleAdmin: {
		ResourceProject:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList, OperationExport},
		ResourceContact:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList, OperationExport},
		ResourceMessage:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList, OperationExport},
		ResourceWebhook:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourcePipeline:  {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourceUser:      {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourceAnalytics: {OperationRead, OperationList, OperationExport},
	},
	RoleManager: {
		ResourceProject:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList, OperationExport},
		ResourceContact:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList, OperationExport},
		ResourceMessage:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList, OperationExport},
		ResourceWebhook:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourcePipeline:  {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourceUser:      {OperationRead, OperationList},
		ResourceAnalytics: {OperationRead, OperationList, OperationExport},
	},
	RoleUser: {
		ResourceProject:   {OperationCreate, OperationRead, OperationUpdate, OperationList},
		ResourceContact:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourceMessage:   {OperationCreate, OperationRead, OperationUpdate, OperationList},
		ResourceWebhook:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourcePipeline:  {OperationCreate, OperationRead, OperationUpdate, OperationList},
		ResourceUser:      {OperationRead, OperationUpdate},
		ResourceAnalytics: {OperationRead},
	},
	RoleReadOnly: {
		ResourceProject:   {OperationRead, OperationList},
		ResourceContact:   {OperationRead, OperationList},
		ResourceMessage:   {OperationRead, OperationList},
		ResourceWebhook:   {OperationRead, OperationList},
		ResourcePipeline:  {OperationRead, OperationList},
		ResourceUser:      {OperationRead},
		ResourceAnalytics: {OperationRead, OperationList},
	},
}
