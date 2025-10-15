package docs

// Este arquivo contém definições Swagger para documentação da API
// Seguindo as melhores práticas de documentação de API

// @title Ventros CRM API
// @version 2.0
// @description API completa do sistema Ventros CRM com suporte a multi-tenancy, mensageria avançada e automação de workflows
// @termsOfService https://ventros.cloud/terms

// @contact.name Ventros Support
// @contact.url https://ventros.cloud/support
// @contact.email support@ventros.cloud

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite 'Bearer' seguido do seu token JWT

// @securityDefinitions.apikey TenantHeader
// @in header
// @name X-Tenant-ID
// @description ID do tenant para multi-tenancy

// @tag.name Health
// @tag.description Endpoints de monitoramento e saúde do sistema

// @tag.name Authentication
// @tag.description Endpoints de autenticação e autorização

// @tag.name Users
// @tag.description Gerenciamento de usuários e perfis

// @tag.name Projects
// @tag.description Gerenciamento de projetos e configurações

// @tag.name Billing
// @tag.description Gerenciamento de cobrança, planos e faturas

// @tag.name Contacts
// @tag.description Gerenciamento de contatos e clientes

// @tag.name Pipelines
// @tag.description Gerenciamento de pipelines e status de vendas

// @tag.name Channels
// @tag.description Configuração e gerenciamento de canais de comunicação

// @tag.name ChannelTypes
// @tag.description Tipos de canais disponíveis (WhatsApp, Telegram, Email, etc.)

// @tag.name Sessions
// @tag.description Gerenciamento de sessões de atendimento

// @tag.name Messages
// @tag.description Sistema de mensageria avançado com suporte a múltiplos canais

// @tag.name Contact Lists
// @tag.description Gerenciamento de listas de contatos com filtros personalizados

// @tag.name Agents
// @tag.description Gerenciamento de agentes e operadores

// @tag.name Webhooks
// @tag.description Sistema de webhooks para integração externa

// @tag.name Analytics
// @tag.description Relatórios e análises de performance

// ErrorResponse representa uma resposta de erro padrão
type ErrorResponse struct {
	Error     string                 `json:"error" example:"Invalid request"`
	Message   string                 `json:"message" example:"The provided data is invalid"`
	Code      string                 `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	RequestID string                 `json:"request_id,omitempty" example:"req_123456789"`
}

// SuccessResponse representa uma resposta de sucesso padrão
type SuccessResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	RequestID string      `json:"request_id,omitempty" example:"req_123456789"`
}

// PaginationResponse representa metadados de paginação
type PaginationResponse struct {
	Page       int  `json:"page" example:"1"`
	Limit      int  `json:"limit" example:"20"`
	Total      int  `json:"total" example:"150"`
	TotalPages int  `json:"total_pages" example:"8"`
	HasNext    bool `json:"has_next" example:"true"`
	HasPrev    bool `json:"has_prev" example:"false"`
}

// ListResponse representa uma resposta de lista com paginação
type ListResponse struct {
	Data       interface{}         `json:"data"`
	Pagination *PaginationResponse `json:"pagination"`
	Filters    map[string]string   `json:"filters,omitempty"`
	Sort       string              `json:"sort,omitempty" example:"created_at:desc"`
}

// HealthCheckResponse representa o status de saúde do sistema
type HealthCheckResponse struct {
	Status    string                 `json:"status" example:"healthy"`
	Version   string                 `json:"version" example:"2.0.0"`
	Timestamp string                 `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Uptime    string                 `json:"uptime" example:"72h30m15s"`
	Services  map[string]interface{} `json:"services"`
}

// ServiceStatus representa o status de um serviço
type ServiceStatus struct {
	Status       string  `json:"status" example:"healthy"`
	ResponseTime string  `json:"response_time" example:"15ms"`
	LastCheck    string  `json:"last_check" example:"2024-01-01T12:00:00Z"`
	Uptime       float64 `json:"uptime" example:"99.9"`
}

// MessageTypeEnum define os tipos de mensagem suportados
type MessageTypeEnum string

const (
	MessageTypeText     MessageTypeEnum = "text"
	MessageTypeImage    MessageTypeEnum = "image"
	MessageTypeAudio    MessageTypeEnum = "audio"
	MessageTypeVideo    MessageTypeEnum = "video"
	MessageTypeDocument MessageTypeEnum = "document"
	MessageTypeLocation MessageTypeEnum = "location"
	MessageTypeContact  MessageTypeEnum = "contact"
	MessageTypeTemplate MessageTypeEnum = "template"
)

// MessagePriorityEnum define as prioridades de mensagem
type MessagePriorityEnum string

const (
	PriorityLow    MessagePriorityEnum = "low"
	PriorityNormal MessagePriorityEnum = "normal"
	PriorityHigh   MessagePriorityEnum = "high"
	PriorityUrgent MessagePriorityEnum = "urgent"
)

// MessageStatusEnum define os status de mensagem
type MessageStatusEnum string

const (
	StatusPending   MessageStatusEnum = "pending"
	StatusQueued    MessageStatusEnum = "queued"
	StatusSending   MessageStatusEnum = "sending"
	StatusSent      MessageStatusEnum = "sent"
	StatusDelivered MessageStatusEnum = "delivered"
	StatusRead      MessageStatusEnum = "read"
	StatusFailed    MessageStatusEnum = "failed"
	StatusExpired   MessageStatusEnum = "expired"
	StatusCancelled MessageStatusEnum = "cancelled"
)

// ChannelTypeEnum define os tipos de canal suportados
type ChannelTypeEnum string

const (
	ChannelTypeWAHA      ChannelTypeEnum = "waha"
	ChannelTypeTelegram  ChannelTypeEnum = "telegram"
	ChannelTypeEmail     ChannelTypeEnum = "email"
	ChannelTypeSMS       ChannelTypeEnum = "sms"
	ChannelTypeWebChat   ChannelTypeEnum = "webchat"
	ChannelTypeFacebook  ChannelTypeEnum = "facebook"
	ChannelTypeInstagram ChannelTypeEnum = "instagram"
)

// SessionStatusEnum define os status de sessão
type SessionStatusEnum string

const (
	SessionStatusActive   SessionStatusEnum = "active"
	SessionStatusInactive SessionStatusEnum = "inactive"
	SessionStatusClosed   SessionStatusEnum = "closed"
	SessionStatusTimeout  SessionStatusEnum = "timeout"
	SessionStatusTransfer SessionStatusEnum = "transfer"
)

// AgentStatusEnum define os status de agente
type AgentStatusEnum string

const (
	AgentStatusOnline      AgentStatusEnum = "online"
	AgentStatusOffline     AgentStatusEnum = "offline"
	AgentStatusBusy        AgentStatusEnum = "busy"
	AgentStatusAway        AgentStatusEnum = "away"
	AgentStatusUnavailable AgentStatusEnum = "unavailable"
)

// WebhookEventEnum define os tipos de eventos de webhook
type WebhookEventEnum string

const (
	WebhookEventMessageReceived WebhookEventEnum = "message.received"
	WebhookEventMessageSent     WebhookEventEnum = "message.sent"
	WebhookEventMessageFailed   WebhookEventEnum = "message.failed"
	WebhookEventSessionStarted  WebhookEventEnum = "session.started"
	WebhookEventSessionClosed   WebhookEventEnum = "session.closed"
	WebhookEventAgentAssigned   WebhookEventEnum = "agent.assigned"
	WebhookEventContactCreated  WebhookEventEnum = "contact.created"
	WebhookEventContactUpdated  WebhookEventEnum = "contact.updated"
)

// FilterOperatorEnum define operadores de filtro
type FilterOperatorEnum string

const (
	FilterOperatorEquals      FilterOperatorEnum = "eq"
	FilterOperatorNotEquals   FilterOperatorEnum = "ne"
	FilterOperatorGreaterThan FilterOperatorEnum = "gt"
	FilterOperatorLessThan    FilterOperatorEnum = "lt"
	FilterOperatorContains    FilterOperatorEnum = "contains"
	FilterOperatorStartsWith  FilterOperatorEnum = "starts_with"
	FilterOperatorEndsWith    FilterOperatorEnum = "ends_with"
	FilterOperatorIn          FilterOperatorEnum = "in"
	FilterOperatorNotIn       FilterOperatorEnum = "not_in"
)

// SortOrderEnum define ordens de classificação
type SortOrderEnum string

const (
	SortOrderAsc  SortOrderEnum = "asc"
	SortOrderDesc SortOrderEnum = "desc"
)

// ContentTypeEnum define tipos de conteúdo
type ContentTypeEnum string

const (
	ContentTypeJSON ContentTypeEnum = "application/json"
	ContentTypeXML  ContentTypeEnum = "application/xml"
	ContentTypeForm ContentTypeEnum = "application/x-www-form-urlencoded"
	ContentTypeText ContentTypeEnum = "text/plain"
)

// RateLimitResponse representa informações de rate limiting
type RateLimitResponse struct {
	Limit     int    `json:"limit" example:"1000"`
	Remaining int    `json:"remaining" example:"999"`
	Reset     string `json:"reset" example:"2024-01-01T13:00:00Z"`
	Window    string `json:"window" example:"1h"`
}

// BatchOperationResponse representa resultado de operação em lote
type BatchOperationResponse struct {
	Total     int           `json:"total" example:"100"`
	Success   int           `json:"success" example:"95"`
	Failed    int           `json:"failed" example:"5"`
	Errors    []BatchError  `json:"errors,omitempty"`
	Results   []interface{} `json:"results,omitempty"`
	Duration  string        `json:"duration" example:"1.5s"`
	RequestID string        `json:"request_id" example:"batch_123456789"`
}

// BatchError representa um erro em operação em lote
type BatchError struct {
	Index   int    `json:"index" example:"5"`
	Error   string `json:"error" example:"Validation failed"`
	Message string `json:"message" example:"Invalid email format"`
	Code    string `json:"code,omitempty" example:"VALIDATION_ERROR"`
}

// WebhookDeliveryResponse representa resultado de entrega de webhook
type WebhookDeliveryResponse struct {
	WebhookID    string `json:"webhook_id" example:"wh_123456789"`
	EventType    string `json:"event_type" example:"message.received"`
	Status       string `json:"status" example:"delivered"`
	Attempts     int    `json:"attempts" example:"1"`
	LastAttempt  string `json:"last_attempt" example:"2024-01-01T12:00:00Z"`
	NextRetry    string `json:"next_retry,omitempty" example:"2024-01-01T12:05:00Z"`
	ResponseCode int    `json:"response_code,omitempty" example:"200"`
	ResponseTime string `json:"response_time,omitempty" example:"150ms"`
	Error        string `json:"error,omitempty"`
}

// AnalyticsTimeRange define intervalos de tempo para analytics
type AnalyticsTimeRange string

const (
	TimeRangeHour  AnalyticsTimeRange = "1h"
	TimeRangeDay   AnalyticsTimeRange = "1d"
	TimeRangeWeek  AnalyticsTimeRange = "7d"
	TimeRangeMonth AnalyticsTimeRange = "30d"
	TimeRangeYear  AnalyticsTimeRange = "365d"
)

// MetricType define tipos de métricas
type MetricType string

const (
	MetricTypeCount   MetricType = "count"
	MetricTypeSum     MetricType = "sum"
	MetricTypeAverage MetricType = "avg"
	MetricTypeMin     MetricType = "min"
	MetricTypeMax     MetricType = "max"
)

// AnalyticsResponse representa resposta de analytics
type AnalyticsResponse struct {
	Metric    string                 `json:"metric" example:"messages_sent"`
	TimeRange string                 `json:"time_range" example:"7d"`
	Data      []AnalyticsDataPoint   `json:"data"`
	Summary   AnalyticsSummary       `json:"summary"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
}

// AnalyticsDataPoint representa um ponto de dados
type AnalyticsDataPoint struct {
	Timestamp string  `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Value     float64 `json:"value" example:"150.5"`
	Label     string  `json:"label,omitempty" example:"WhatsApp"`
}

// AnalyticsSummary representa resumo de analytics
type AnalyticsSummary struct {
	Total   float64 `json:"total" example:"1500"`
	Average float64 `json:"average" example:"214.3"`
	Min     float64 `json:"min" example:"50"`
	Max     float64 `json:"max" example:"500"`
	Change  float64 `json:"change" example:"15.5"`
}

// ExportFormat define formatos de exportação
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatXML  ExportFormat = "xml"
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatXLSX ExportFormat = "xlsx"
)

// ExportRequest representa requisição de exportação
type ExportRequest struct {
	Format    ExportFormat           `json:"format" example:"csv"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Fields    []string               `json:"fields,omitempty" example:"id,name,email,created_at"`
	DateRange *DateRange             `json:"date_range,omitempty"`
	Sort      string                 `json:"sort,omitempty" example:"created_at:desc"`
}

// DateRange representa um intervalo de datas
type DateRange struct {
	From string `json:"from" example:"2024-01-01T00:00:00Z"`
	To   string `json:"to" example:"2024-01-31T23:59:59Z"`
}

// ExportResponse representa resposta de exportação
type ExportResponse struct {
	ExportID    string `json:"export_id" example:"exp_123456789"`
	Status      string `json:"status" example:"processing"`
	Format      string `json:"format" example:"csv"`
	FileSize    int64  `json:"file_size,omitempty" example:"1048576"`
	RecordCount int    `json:"record_count,omitempty" example:"1000"`
	DownloadURL string `json:"download_url,omitempty" example:"https://api.ventros.cloud/exports/exp_123456789/download"`
	ExpiresAt   string `json:"expires_at,omitempty" example:"2024-01-02T12:00:00Z"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T12:00:00Z"`
}

// ValidationError representa erro de validação detalhado
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Invalid email format"`
	Code    string `json:"code" example:"INVALID_FORMAT"`
	Value   string `json:"value,omitempty" example:"invalid-email"`
}

// BulkValidationResponse representa resposta de validação em lote
type BulkValidationResponse struct {
	Valid   int               `json:"valid" example:"95"`
	Invalid int               `json:"invalid" example:"5"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Details []BulkItemResult  `json:"details,omitempty"`
}

// BulkItemResult representa resultado de um item em operação em lote
type BulkItemResult struct {
	Index  int    `json:"index" example:"0"`
	Status string `json:"status" example:"success"`
	ID     string `json:"id,omitempty" example:"123456789"`
	Error  string `json:"error,omitempty"`
}

// SystemInfo representa informações do sistema
type SystemInfo struct {
	Version     string          `json:"version" example:"2.0.0"`
	BuildDate   string          `json:"build_date" example:"2024-01-01T12:00:00Z"`
	GitCommit   string          `json:"git_commit" example:"abc123def456"`
	Environment string          `json:"environment" example:"production"`
	Features    map[string]bool `json:"features"`
	Limits      map[string]int  `json:"limits"`
	Endpoints   []EndpointInfo  `json:"endpoints,omitempty"`
}

// EndpointInfo representa informações de um endpoint
type EndpointInfo struct {
	Path        string   `json:"path" example:"/api/v1/messages"`
	Methods     []string `json:"methods" example:"GET,POST"`
	Description string   `json:"description" example:"Message management endpoints"`
	Version     string   `json:"version" example:"2.0"`
	Deprecated  bool     `json:"deprecated" example:"false"`
}

// ConfigurationResponse representa configuração do sistema
type ConfigurationResponse struct {
	Features     map[string]bool      `json:"features"`
	Limits       map[string]int       `json:"limits"`
	Channels     []string             `json:"channels" example:"waha,telegram,email"`
	Integrations []IntegrationInfo    `json:"integrations"`
	Webhooks     WebhookConfiguration `json:"webhooks"`
}

// IntegrationInfo representa informações de integração
type IntegrationInfo struct {
	Name         string            `json:"name" example:"WAHA"`
	Type         string            `json:"type" example:"messaging"`
	Status       string            `json:"status" example:"active"`
	Version      string            `json:"version" example:"1.0.0"`
	Capabilities []string          `json:"capabilities" example:"send_message,receive_message"`
	Config       map[string]string `json:"config,omitempty"`
}

// WebhookConfiguration representa configuração de webhooks
type WebhookConfiguration struct {
	MaxRetries    int      `json:"max_retries" example:"3"`
	RetryInterval string   `json:"retry_interval" example:"5m"`
	Timeout       string   `json:"timeout" example:"30s"`
	Events        []string `json:"events" example:"message.received,session.started"`
}

// SearchRequest representa requisição de busca
type SearchRequest struct {
	Query     string                 `json:"query" example:"customer support"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Sort      []SortField            `json:"sort,omitempty"`
	Page      int                    `json:"page" example:"1"`
	Limit     int                    `json:"limit" example:"20"`
	Highlight bool                   `json:"highlight" example:"true"`
}

// SortField representa campo de ordenação
type SortField struct {
	Field string `json:"field" example:"created_at"`
	Order string `json:"order" example:"desc"`
}

// SearchResponse representa resposta de busca
type SearchResponse struct {
	Query    string             `json:"query" example:"customer support"`
	Results  []SearchResult     `json:"results"`
	Total    int                `json:"total" example:"150"`
	Page     int                `json:"page" example:"1"`
	Limit    int                `json:"limit" example:"20"`
	Duration string             `json:"duration" example:"15ms"`
	Facets   map[string][]Facet `json:"facets,omitempty"`
}

// SearchResult representa resultado de busca
type SearchResult struct {
	ID        string                 `json:"id" example:"123456789"`
	Type      string                 `json:"type" example:"message"`
	Score     float64                `json:"score" example:"0.95"`
	Data      map[string]interface{} `json:"data"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
}

// Facet representa faceta de busca
type Facet struct {
	Value string `json:"value" example:"WhatsApp"`
	Count int    `json:"count" example:"50"`
}
