package user

// ResourceType representa os tipos de recursos no sistema
type ResourceType string

const (
	ResourceProject    ResourceType = "project"
	ResourceContact    ResourceType = "contact"
	ResourceMessage    ResourceType = "message"
	ResourceWebhook    ResourceType = "webhook"
	ResourcePipeline   ResourceType = "pipeline"
	ResourceUser       ResourceType = "user"
	ResourceAnalytics  ResourceType = "analytics"
)

// Operation representa as operações possíveis
type Operation string

const (
	OperationCreate Operation = "create"
	OperationRead   Operation = "read"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
	OperationList   Operation = "list"
	OperationExport Operation = "export"
)

// Permission representa uma permissão específica
type Permission struct {
	Resource  ResourceType
	Operation Operation
}

// NewPermission cria uma nova permissão
func NewPermission(resource ResourceType, operation Operation) Permission {
	return Permission{
		Resource:  resource,
		Operation: operation,
	}
}

// IsManagerAllowed verifica se managers podem executar esta operação
func (p Permission) IsManagerAllowed() bool {
	switch p.Resource {
	case ResourceUser:
		// Managers não podem gerenciar usuários
		return p.Operation == OperationRead || p.Operation == OperationList
	case ResourceAnalytics:
		// Managers podem ver analytics de sua equipe
		return p.Operation == OperationRead || p.Operation == OperationList
	default:
		// Managers podem fazer tudo com recursos de negócio
		return true
	}
}

// IsUserAllowed verifica se usuários padrão podem executar esta operação
func (p Permission) IsUserAllowed() bool {
	switch p.Resource {
	case ResourceUser:
		// Usuários só podem ver/editar próprio perfil
		return p.Operation == OperationRead || p.Operation == OperationUpdate
	case ResourceAnalytics:
		// Usuários podem ver apenas suas próprias analytics
		return p.Operation == OperationRead
	default:
		// Usuários podem fazer tudo com seus próprios recursos
		return true
	}
}

// IsReadOnlyAllowed verifica se usuários readonly podem executar esta operação
func (p Permission) IsReadOnlyAllowed() bool {
	// ReadOnly só pode ler
	return p.Operation == OperationRead || p.Operation == OperationList
}

// String implementa fmt.Stringer
func (p Permission) String() string {
	return string(p.Resource) + ":" + string(p.Operation)
}

// PermissionMatrix define a matriz completa de permissões
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
		ResourceUser:      {OperationRead, OperationList}, // Só pode ver usuários
		ResourceAnalytics: {OperationRead, OperationList, OperationExport},
	},
	RoleUser: {
		ResourceProject:   {OperationCreate, OperationRead, OperationUpdate, OperationList},
		ResourceContact:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourceMessage:   {OperationCreate, OperationRead, OperationUpdate, OperationList},
		ResourceWebhook:   {OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList},
		ResourcePipeline:  {OperationCreate, OperationRead, OperationUpdate, OperationList},
		ResourceUser:      {OperationRead, OperationUpdate}, // Só próprio perfil
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
