package agent

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name      string
		projectID uuid.UUID
		tenantID  string
		agentName string
		agentType AgentType
		userID    *uuid.UUID
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid human agent",
			projectID: projectID,
			tenantID:  "tenant-123",
			agentName: "John Doe",
			agentType: AgentTypeHuman,
			userID:    &userID,
			wantErr:   false,
		},
		{
			name:      "valid AI agent without userID",
			projectID: projectID,
			tenantID:  "tenant-123",
			agentName: "AI Assistant",
			agentType: AgentTypeAI,
			userID:    nil,
			wantErr:   false,
		},
		{
			name:      "nil projectID",
			projectID: uuid.Nil,
			tenantID:  "tenant-123",
			agentName: "Agent",
			agentType: AgentTypeHuman,
			userID:    &userID,
			wantErr:   true,
			errMsg:    "projectID cannot be nil",
		},
		{
			name:      "empty tenantID",
			projectID: projectID,
			tenantID:  "",
			agentName: "Agent",
			agentType: AgentTypeHuman,
			userID:    &userID,
			wantErr:   true,
			errMsg:    "tenantID cannot be empty",
		},
		{
			name:      "empty name",
			projectID: projectID,
			tenantID:  "tenant-123",
			agentName: "",
			agentType: AgentTypeHuman,
			userID:    &userID,
			wantErr:   true,
			errMsg:    "name cannot be empty",
		},
		{
			name:      "human agent without userID",
			projectID: projectID,
			tenantID:  "tenant-123",
			agentName: "Agent",
			agentType: AgentTypeHuman,
			userID:    nil,
			wantErr:   true,
			errMsg:    "human agent requires a valid userID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewAgent(tt.projectID, tt.tenantID, tt.agentName, tt.agentType, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, agent)
			} else {
				require.NoError(t, err)
				require.NotNil(t, agent)

				assert.NotEqual(t, uuid.Nil, agent.ID())
				assert.Equal(t, tt.projectID, agent.ProjectID())
				assert.Equal(t, tt.tenantID, agent.TenantID())
				assert.Equal(t, tt.agentName, agent.Name())
				assert.Equal(t, tt.agentType, agent.Type())
				assert.Equal(t, AgentStatusOffline, agent.Status())
				assert.True(t, agent.IsActive())
				assert.NotNil(t, agent.Config())
				assert.NotNil(t, agent.Permissions())
				assert.NotNil(t, agent.Settings())
				assert.Equal(t, 0, agent.SessionsHandled())
				assert.Equal(t, 0, agent.AverageResponseMs())
				assert.Nil(t, agent.LastActivityAt())

				// Check domain event
				events := agent.DomainEvents()
				require.Len(t, events, 1)
				_, ok := events[0].(AgentCreatedEvent)
				require.True(t, ok)
			}
		})
	}
}

func TestAgent_UpdateProfile(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Original Name", AgentTypeHuman, &userID)
	require.NoError(t, err)
	agent.ClearEvents()

	t.Run("successful profile update", func(t *testing.T) {
		err := agent.UpdateProfile("New Name", "new.email@example.com")
		require.NoError(t, err)

		assert.Equal(t, "New Name", agent.Name())
		assert.Equal(t, "new.email@example.com", agent.Email())

		events := agent.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(AgentUpdatedEvent)
		require.True(t, ok)
		assert.NotNil(t, event.Changes)
	})

	t.Run("empty name", func(t *testing.T) {
		err := agent.UpdateProfile("", "email@example.com")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("empty email", func(t *testing.T) {
		err := agent.UpdateProfile("Name", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email cannot be empty")
	})

	t.Run("no changes", func(t *testing.T) {
		agent.ClearEvents()
		currentName := agent.Name()
		currentEmail := agent.Email()

		err := agent.UpdateProfile(currentName, currentEmail)
		require.NoError(t, err)

		// Should not generate event when nothing changed
		events := agent.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestAgent_ActivateDeactivate(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeHuman, &userID)
	require.NoError(t, err)
	agent.ClearEvents()

	t.Run("deactivate active agent", func(t *testing.T) {
		assert.True(t, agent.IsActive())

		err := agent.Deactivate()
		require.NoError(t, err)
		assert.False(t, agent.IsActive())

		events := agent.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(AgentDeactivatedEvent)
		require.True(t, ok)
	})

	t.Run("deactivate already inactive agent", func(t *testing.T) {
		err := agent.Deactivate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already inactive")
	})

	t.Run("activate inactive agent", func(t *testing.T) {
		agent.ClearEvents()
		err := agent.Activate()
		require.NoError(t, err)
		assert.True(t, agent.IsActive())

		events := agent.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(AgentActivatedEvent)
		require.True(t, ok)
	})

	t.Run("activate already active agent", func(t *testing.T) {
		err := agent.Activate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already active")
	})
}

func TestAgent_Permissions(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeHuman, &userID)
	require.NoError(t, err)
	agent.ClearEvents()

	t.Run("grant permission", func(t *testing.T) {
		err := agent.GrantPermission("read:contacts")
		require.NoError(t, err)

		assert.True(t, agent.HasPermission("read:contacts"))

		events := agent.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(AgentPermissionGrantedEvent)
		require.True(t, ok)
		assert.Equal(t, "read:contacts", event.Permission)
	})

	t.Run("grant empty permission", func(t *testing.T) {
		err := agent.GrantPermission("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission cannot be empty")
	})

	t.Run("grant duplicate permission", func(t *testing.T) {
		agent.ClearEvents()
		err := agent.GrantPermission("read:contacts")
		require.NoError(t, err)

		// Should succeed but not generate event
		events := agent.DomainEvents()
		assert.Len(t, events, 0)
	})

	t.Run("revoke permission", func(t *testing.T) {
		agent.ClearEvents()
		err := agent.RevokePermission("read:contacts")
		require.NoError(t, err)

		assert.False(t, agent.HasPermission("read:contacts"))

		events := agent.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(AgentPermissionRevokedEvent)
		require.True(t, ok)
		assert.Equal(t, "read:contacts", event.Permission)
	})

	t.Run("revoke empty permission", func(t *testing.T) {
		err := agent.RevokePermission("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission cannot be empty")
	})

	t.Run("revoke non-existent permission", func(t *testing.T) {
		agent.ClearEvents()
		err := agent.RevokePermission("write:contacts")
		require.NoError(t, err)

		// Should succeed but not generate event
		events := agent.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestAgent_StatusManagement(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeHuman, &userID)
	require.NoError(t, err)

	t.Run("set status", func(t *testing.T) {
		agent.SetStatus(AgentStatusAvailable)
		assert.Equal(t, AgentStatusAvailable, agent.Status())
		assert.NotNil(t, agent.LastActivityAt())
	})

	t.Run("set same status", func(t *testing.T) {
		oldActivity := agent.LastActivityAt()
		time.Sleep(10 * time.Millisecond)

		agent.SetStatus(AgentStatusAvailable)

		// LastActivityAt should not change if status is the same
		assert.Equal(t, oldActivity, agent.LastActivityAt())
	})

	t.Run("set different status", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond)
		agent.SetStatus(AgentStatusBusy)

		assert.Equal(t, AgentStatusBusy, agent.Status())
	})
}

func TestAgent_SessionHandling(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeHuman, &userID)
	require.NoError(t, err)

	t.Run("record first session", func(t *testing.T) {
		agent.RecordSessionHandled(1000)

		assert.Equal(t, 1, agent.SessionsHandled())
		assert.Equal(t, 1000, agent.AverageResponseMs())
		assert.NotNil(t, agent.LastActivityAt())
	})

	t.Run("record multiple sessions calculates average", func(t *testing.T) {
		agent.RecordSessionHandled(2000)

		assert.Equal(t, 2, agent.SessionsHandled())
		// Average should be (1000 + 2000) / 2 = 1500
		assert.Equal(t, 1500, agent.AverageResponseMs())
	})

	t.Run("record session with different response time", func(t *testing.T) {
		agent.RecordSessionHandled(500)

		assert.Equal(t, 3, agent.SessionsHandled())
		// Average should be (1500 + 500) / 2 = 1000
		assert.Equal(t, 1000, agent.AverageResponseMs())
	})
}

func TestAgent_ConfigAndSettings(t *testing.T) {
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeAI, nil)
	require.NoError(t, err)

	t.Run("set config", func(t *testing.T) {
		config := map[string]interface{}{
			"provider":    "openai",
			"model":       "gpt-4",
			"temperature": 0.7,
		}

		agent.SetConfig(config)

		agentConfig := agent.Config()
		assert.Equal(t, "openai", agentConfig["provider"])
		assert.Equal(t, "gpt-4", agentConfig["model"])
		assert.Equal(t, 0.7, agentConfig["temperature"])
	})

	t.Run("update settings", func(t *testing.T) {
		settings := map[string]interface{}{
			"notification_enabled": true,
			"auto_reply":           false,
		}

		agent.UpdateSettings(settings)

		agentSettings := agent.Settings()
		assert.Equal(t, true, agentSettings["notification_enabled"])
		assert.Equal(t, false, agentSettings["auto_reply"])
	})
}

func TestAgent_RecordLogin(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeHuman, &userID)
	require.NoError(t, err)
	agent.ClearEvents()

	t.Run("record login", func(t *testing.T) {
		beforeLogin := time.Now()
		time.Sleep(10 * time.Millisecond)

		agent.RecordLogin()

		assert.NotNil(t, agent.LastLoginAt())
		assert.True(t, agent.LastLoginAt().After(beforeLogin))

		events := agent.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(AgentLoggedInEvent)
		require.True(t, ok)
	})
}

func TestReconstructAgent(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()
	lastLogin := time.Now().Add(-1 * time.Hour)

	permissions := map[string]bool{
		"read:contacts":  true,
		"write:contacts": true,
	}

	settings := map[string]interface{}{
		"theme": "dark",
	}

	t.Run("reconstruct agent", func(t *testing.T) {
		projectID := uuid.New()
		userID := uuid.New()
		agent := ReconstructAgent(
			id,
			1, // version
			projectID,
			&userID,
			"tenant-123",
			"Reconstructed Agent",
			"agent@example.com",
			AgentTypeHuman,
			AgentStatusAvailable,
			RoleHumanAgent,
			true,
			map[string]interface{}{}, // config
			permissions,
			settings,
			0,   // sessionsHandled
			0,   // averageResponseMs
			nil, // lastActivityAt
			nil, // virtualMetadata
			createdAt,
			updatedAt,
			&lastLogin,
		)

		assert.Equal(t, id, agent.ID())
		assert.Equal(t, "tenant-123", agent.TenantID())
		assert.Equal(t, "Reconstructed Agent", agent.Name())
		assert.Equal(t, "agent@example.com", agent.Email())
		assert.Equal(t, RoleHumanAgent, agent.Role())
		assert.True(t, agent.IsActive())
		assert.Equal(t, createdAt, agent.CreatedAt())
		assert.Equal(t, updatedAt, agent.UpdatedAt())
		assert.Equal(t, &lastLogin, agent.LastLoginAt())
		assert.Len(t, agent.DomainEvents(), 0) // No events on reconstruction

		// Check permissions
		assert.True(t, agent.HasPermission("read:contacts"))
		assert.True(t, agent.HasPermission("write:contacts"))
		assert.False(t, agent.HasPermission("delete:contacts"))

		// Check settings
		agentSettings := agent.Settings()
		assert.Equal(t, "dark", agentSettings["theme"])
	})
}

func TestAgent_AgentTypes(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name      string
		agentType AgentType
		userID    *uuid.UUID
	}{
		{"human agent", AgentTypeHuman, &userID},
		{"AI agent", AgentTypeAI, nil},
		{"bot agent", AgentTypeBot, nil},
		{"channel agent", AgentTypeChannel, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewAgent(projectID, "tenant-123", "Test "+string(tt.agentType), tt.agentType, tt.userID)
			require.NoError(t, err)

			assert.Equal(t, tt.agentType, agent.Type())

			if tt.agentType == AgentTypeHuman {
				assert.NotNil(t, agent.UserID())
			} else {
				// Non-human agents may or may not have userID
				if agent.UserID() != nil {
					assert.NotEqual(t, uuid.Nil, *agent.UserID())
				}
			}
		})
	}
}

func TestAgent_EventManagement(t *testing.T) {
	userID := uuid.New()
	agent, err := NewAgent(uuid.New(), "tenant-123", "Test Agent", AgentTypeHuman, &userID)
	require.NoError(t, err)

	t.Run("clear events", func(t *testing.T) {
		assert.Len(t, agent.DomainEvents(), 1) // Creation event

		agent.ClearEvents()
		assert.Len(t, agent.DomainEvents(), 0)
	})

	t.Run("multiple operations generate events", func(t *testing.T) {
		agent.ClearEvents()

		_ = agent.UpdateProfile("New Name", "new@example.com")
		_ = agent.GrantPermission("read:contacts")
		_ = agent.Deactivate()

		events := agent.DomainEvents()
		assert.Len(t, events, 3)
	})
}
