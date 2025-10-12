package project_member

// Permission representa uma permissão granular no sistema
type Permission string

const (
	// Session & Messages
	PermissionViewSessions    Permission = "sessions.view"
	PermissionManageSessions  Permission = "sessions.manage"
	PermissionSendMessages    Permission = "messages.send"
	PermissionViewMessages    Permission = "messages.view"

	// Contacts
	PermissionViewContacts    Permission = "contacts.view"
	PermissionManageContacts  Permission = "contacts.manage"
	PermissionExportContacts  Permission = "contacts.export"

	// Pipelines
	PermissionViewPipelines   Permission = "pipelines.view"
	PermissionManagePipelines Permission = "pipelines.manage"

	// Campaigns & Sequences
	PermissionViewCampaigns    Permission = "campaigns.view"
	PermissionManageCampaigns  Permission = "campaigns.manage"
	PermissionViewSequences    Permission = "sequences.view"
	PermissionManageSequences  Permission = "sequences.manage"

	// Analytics
	PermissionViewAnalytics   Permission = "analytics.view"
	PermissionExportAnalytics Permission = "analytics.export"

	// Agents & Members
	PermissionViewMembers   Permission = "members.view"
	PermissionManageMembers Permission = "members.manage"

	// Channels
	PermissionViewChannels   Permission = "channels.view"
	PermissionManageChannels Permission = "channels.manage"

	// Billing (apenas Customer, não Project-level)
	PermissionViewBilling   Permission = "billing.view"
	PermissionManageBilling Permission = "billing.manage"

	// Settings
	PermissionViewSettings   Permission = "settings.view"
	PermissionManageSettings Permission = "settings.manage"
)

// rolePermissions mapeia cada role para suas permissões
var rolePermissions = map[ProjectMemberRole][]Permission{
	// Admin - Full access
	RoleAdmin: {
		PermissionViewSessions,
		PermissionManageSessions,
		PermissionSendMessages,
		PermissionViewMessages,
		PermissionViewContacts,
		PermissionManageContacts,
		PermissionExportContacts,
		PermissionViewPipelines,
		PermissionManagePipelines,
		PermissionViewCampaigns,
		PermissionManageCampaigns,
		PermissionViewSequences,
		PermissionManageSequences,
		PermissionViewAnalytics,
		PermissionExportAnalytics,
		PermissionViewMembers,
		PermissionManageMembers,
		PermissionViewChannels,
		PermissionManageChannels,
		PermissionViewSettings,
		PermissionManageSettings,
	},

	// Supervisor - Gerenciamento operacional, analytics, campaigns
	RoleSupervisor: {
		PermissionViewSessions,
		PermissionManageSessions,
		PermissionSendMessages,
		PermissionViewMessages,
		PermissionViewContacts,
		PermissionManageContacts,
		PermissionExportContacts,
		PermissionViewPipelines,
		PermissionManagePipelines,
		PermissionViewCampaigns,
		PermissionManageCampaigns,
		PermissionViewSequences,
		PermissionManageSequences,
		PermissionViewAnalytics,
		PermissionExportAnalytics,
		PermissionViewMembers, // Pode VER membros, mas não gerenciar
		PermissionViewChannels,
		PermissionViewSettings,
	},

	// Agent - Atendimento e interação com clientes
	RoleAgent: {
		PermissionViewSessions,
		PermissionManageSessions,
		PermissionSendMessages,
		PermissionViewMessages,
		PermissionViewContacts,
		PermissionManageContacts, // Pode criar/editar contacts
		PermissionViewPipelines,
		PermissionViewCampaigns,
		PermissionViewSequences,
		PermissionViewMembers, // Pode ver outros membros
		PermissionViewChannels,
	},

	// Viewer - Read-only access
	RoleViewer: {
		PermissionViewSessions,
		PermissionViewMessages,
		PermissionViewContacts,
		PermissionViewPipelines,
		PermissionViewCampaigns,
		PermissionViewSequences,
		PermissionViewAnalytics,
		PermissionViewMembers,
		PermissionViewChannels,
		PermissionViewSettings,
	},
}

// RoleHasPermission verifica se um role tem uma permissão específica
func RoleHasPermission(role ProjectMemberRole, permission Permission) bool {
	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetRolePermissions retorna todas as permissões de um role
func GetRolePermissions(role ProjectMemberRole) []Permission {
	permissions, exists := rolePermissions[role]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// AllPermissions retorna todas as permissões disponíveis no sistema
func AllPermissions() []Permission {
	return []Permission{
		PermissionViewSessions,
		PermissionManageSessions,
		PermissionSendMessages,
		PermissionViewMessages,
		PermissionViewContacts,
		PermissionManageContacts,
		PermissionExportContacts,
		PermissionViewPipelines,
		PermissionManagePipelines,
		PermissionViewCampaigns,
		PermissionManageCampaigns,
		PermissionViewSequences,
		PermissionManageSequences,
		PermissionViewAnalytics,
		PermissionExportAnalytics,
		PermissionViewMembers,
		PermissionManageMembers,
		PermissionViewChannels,
		PermissionManageChannels,
		PermissionViewBilling,
		PermissionManageBilling,
		PermissionViewSettings,
		PermissionManageSettings,
	}
}

// PermissionDescription retorna descrição legível de uma permissão
func PermissionDescription(p Permission) string {
	descriptions := map[Permission]string{
		PermissionViewSessions:     "View chat sessions",
		PermissionManageSessions:   "Manage chat sessions (assign, close, transfer)",
		PermissionSendMessages:     "Send messages to contacts",
		PermissionViewMessages:     "View message history",
		PermissionViewContacts:     "View contacts",
		PermissionManageContacts:   "Create, edit, and delete contacts",
		PermissionExportContacts:   "Export contact lists",
		PermissionViewPipelines:    "View pipelines and deals",
		PermissionManagePipelines:  "Create and manage pipelines",
		PermissionViewCampaigns:    "View campaigns",
		PermissionManageCampaigns:  "Create and manage campaigns",
		PermissionViewSequences:    "View sequences",
		PermissionManageSequences:  "Create and manage sequences",
		PermissionViewAnalytics:    "View analytics and reports",
		PermissionExportAnalytics:  "Export analytics data",
		PermissionViewMembers:      "View project members",
		PermissionManageMembers:    "Invite, remove, and manage member roles",
		PermissionViewChannels:     "View communication channels",
		PermissionManageChannels:   "Create and configure channels",
		PermissionViewBilling:      "View billing information",
		PermissionManageBilling:    "Manage billing and subscriptions",
		PermissionViewSettings:     "View project settings",
		PermissionManageSettings:   "Modify project settings",
	}

	if desc, exists := descriptions[p]; exists {
		return desc
	}
	return string(p)
}
