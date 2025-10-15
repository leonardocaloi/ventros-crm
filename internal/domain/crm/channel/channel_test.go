package channel

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChannel_Valid(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	name := "My WhatsApp Channel"

	tests := []struct {
		name        string
		channelType ChannelType
	}{
		{
			name:        "WAHA channel",
			channelType: TypeWAHA,
		},
		{
			name:        "WhatsApp channel",
			channelType: TypeWhatsApp,
		},
		{
			name:        "Telegram channel",
			channelType: TypeTelegram,
		},
		{
			name:        "Messenger channel",
			channelType: TypeMessenger,
		},
		{
			name:        "Instagram channel",
			channelType: TypeInstagram,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := NewChannel(userID, projectID, tenantID, name, tt.channelType)
			require.NoError(t, err)
			assert.NotNil(t, ch)
			assert.NotEqual(t, uuid.Nil, ch.ID)
			assert.Equal(t, userID, ch.UserID)
			assert.Equal(t, projectID, ch.ProjectID)
			assert.Equal(t, tenantID, ch.TenantID)
			assert.Equal(t, name, ch.Name)
			assert.Equal(t, tt.channelType, ch.Type)
			assert.Equal(t, StatusInactive, ch.Status)
			assert.NotNil(t, ch.Config)
			assert.NotZero(t, ch.CreatedAt)
			assert.NotZero(t, ch.UpdatedAt)

			// Verifica evento de criação
			events := ch.DomainEvents()
			assert.Len(t, events, 1)
			createdEvent, ok := events[0].(ChannelCreatedEvent)
			assert.True(t, ok)
			assert.Equal(t, ch.ID, createdEvent.ChannelID)
		})
	}
}

func TestNewChannel_Invalid(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	tests := []struct {
		name        string
		channelName string
		channelType ChannelType
		expectErr   bool
		errContains string
	}{
		{
			name:        "empty name",
			channelName: "",
			channelType: TypeWAHA,
			expectErr:   true,
			errContains: "name is required",
		},
		{
			name:        "invalid channel type",
			channelName: "Test Channel",
			channelType: ChannelType("invalid"),
			expectErr:   true,
			errContains: "invalid channel type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := NewChannel(userID, projectID, tenantID, tt.channelName, tt.channelType)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, ch)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ch)
			}
		})
	}
}

func TestNewWAHAChannel(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	name := "WAHA Channel"

	t.Run("valid WAHA channel", func(t *testing.T) {
		config := WAHAConfig{
			BaseURL: "https://waha.example.com",
			Auth: WAHAAuth{
				APIKey: "test-api-key",
			},
			SessionID:      "session-123",
			WebhookURL:     "https://app.example.com/webhooks/waha",
			ImportStrategy: WAHAImportNewOnly,
		}

		ch, err := NewWAHAChannel(userID, projectID, tenantID, name, config)
		require.NoError(t, err)
		assert.NotNil(t, ch)
		assert.Equal(t, TypeWAHA, ch.Type)
		assert.Equal(t, "session-123", ch.ExternalID)

		wahaConfig, err := ch.GetWAHAConfig()
		require.NoError(t, err)
		assert.Equal(t, config.BaseURL, wahaConfig.BaseURL)
		assert.Equal(t, config.Auth.APIKey, wahaConfig.Auth.APIKey)
		assert.Equal(t, config.SessionID, wahaConfig.SessionID)
	})

	t.Run("invalid WAHA config", func(t *testing.T) {
		config := WAHAConfig{
			BaseURL: "", // Missing
			Auth: WAHAAuth{
				APIKey: "test-api-key",
			},
			SessionID: "session-123",
		}

		ch, err := NewWAHAChannel(userID, projectID, tenantID, name, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base URL is required")
		assert.Nil(t, ch)
	})
}

func TestNewWhatsAppChannel(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	name := "WhatsApp Channel"

	t.Run("valid WhatsApp channel", func(t *testing.T) {
		config := WhatsAppConfig{
			AccessToken:   "test-token",
			PhoneNumberID: "1234567890",
			BusinessID:    "business-123",
			WebhookURL:    "https://app.example.com/webhooks/whatsapp",
			VerifyToken:   "verify-token",
		}

		ch, err := NewWhatsAppChannel(userID, projectID, tenantID, name, config)
		require.NoError(t, err)
		assert.NotNil(t, ch)
		assert.Equal(t, TypeWhatsApp, ch.Type)
		assert.Equal(t, "1234567890", ch.ExternalID)
	})
}

func TestNewTelegramChannel(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	name := "Telegram Channel"

	t.Run("valid Telegram channel", func(t *testing.T) {
		config := TelegramConfig{
			BotToken:   "bot-token-123",
			BotID:      "bot-123",
			WebhookURL: "https://app.example.com/webhooks/telegram",
		}

		ch, err := NewTelegramChannel(userID, projectID, tenantID, name, config)
		require.NoError(t, err)
		assert.NotNil(t, ch)
		assert.Equal(t, TypeTelegram, ch.Type)
		assert.Equal(t, "bot-123", ch.ExternalID)
	})
}

func TestChannel_Activate(t *testing.T) {
	ch := createTestChannel(t)

	ch.Activate()

	assert.Equal(t, StatusActive, ch.Status)
	assert.True(t, ch.IsActive())

	events := ch.DomainEvents()
	assert.Len(t, events, 2) // Created + Activated
	activatedEvent, ok := events[1].(ChannelActivatedEvent)
	assert.True(t, ok)
	assert.Equal(t, ch.ID, activatedEvent.ChannelID)
}

func TestChannel_Deactivate(t *testing.T) {
	ch := createTestChannel(t)
	ch.Activate()
	ch.ClearEvents()

	ch.Deactivate()

	assert.Equal(t, StatusInactive, ch.Status)
	assert.False(t, ch.IsActive())

	events := ch.DomainEvents()
	assert.Len(t, events, 1)
	deactivatedEvent, ok := events[0].(ChannelDeactivatedEvent)
	assert.True(t, ok)
	assert.Equal(t, ch.ID, deactivatedEvent.ChannelID)
}

func TestChannel_SetConnecting(t *testing.T) {
	ch := createTestChannel(t)

	ch.SetConnecting()

	assert.Equal(t, StatusConnecting, ch.Status)
}

func TestChannel_SetError(t *testing.T) {
	ch := createTestChannel(t)
	errorMsg := "Connection failed"

	ch.SetError(errorMsg)

	assert.Equal(t, StatusError, ch.Status)
	assert.Equal(t, errorMsg, ch.LastError)
	assert.NotNil(t, ch.LastErrorAt)
}

func TestChannel_IncrementMessagesReceived(t *testing.T) {
	ch := createTestChannel(t)
	assert.Equal(t, 0, ch.MessagesReceived)
	assert.Nil(t, ch.LastMessageAt)

	ch.IncrementMessagesReceived()

	assert.Equal(t, 1, ch.MessagesReceived)
	assert.NotNil(t, ch.LastMessageAt)

	ch.IncrementMessagesReceived()
	assert.Equal(t, 2, ch.MessagesReceived)
}

func TestChannel_IncrementMessagesSent(t *testing.T) {
	ch := createTestChannel(t)
	assert.Equal(t, 0, ch.MessagesSent)

	ch.IncrementMessagesSent()

	assert.Equal(t, 1, ch.MessagesSent)
}

func TestChannel_WAHASessionStatus(t *testing.T) {
	ch := createWAHATestChannel(t)

	tests := []struct {
		name         string
		status       WAHASessionStatus
		expectActive bool
	}{
		{
			name:         "STARTING status",
			status:       WAHASessionStatusStarting,
			expectActive: false,
		},
		{
			name:         "SCAN_QR_CODE status",
			status:       WAHASessionStatusScanQR,
			expectActive: false,
		},
		{
			name:         "WORKING status activates channel",
			status:       WAHASessionStatusWorking,
			expectActive: true,
		},
		{
			name:         "FAILED status deactivates channel",
			status:       WAHASessionStatusFailed,
			expectActive: false,
		},
		{
			name:         "STOPPED status deactivates channel",
			status:       WAHASessionStatusStopped,
			expectActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch.SetWAHASessionStatus(tt.status)

			assert.Equal(t, tt.status, ch.GetWAHASessionStatus())

			if tt.expectActive {
				assert.True(t, ch.IsActive())
			} else {
				assert.False(t, ch.IsActive())
			}
		})
	}
}

func TestChannel_WAHAQRCode(t *testing.T) {
	ch := createWAHATestChannel(t)

	t.Run("set and get QR code", func(t *testing.T) {
		qrCode := "QR_CODE_DATA_HERE"
		ch.SetWAHAQRCode(qrCode)

		assert.Equal(t, qrCode, ch.GetWAHAQRCode())
		assert.NotNil(t, ch.Config["qr_generated_at"])
	})

	t.Run("QR code validity", func(t *testing.T) {
		qrCode := "QR_CODE_DATA_HERE"
		ch.SetWAHASessionStatus(WAHASessionStatusScanQR)
		ch.SetWAHAQRCode(qrCode)

		// Should be valid immediately after setting
		assert.True(t, ch.IsWAHAQRCodeValid())

		// Should be invalid if already working
		ch.SetWAHASessionStatus(WAHASessionStatusWorking)
		assert.False(t, ch.IsWAHAQRCodeValid())
	})

	t.Run("clear QR code", func(t *testing.T) {
		qrCode := "QR_CODE_DATA_HERE"
		ch.SetWAHAQRCode(qrCode)

		ch.ClearWAHAQRCode()

		assert.Empty(t, ch.GetWAHAQRCode())
		assert.Nil(t, ch.Config["qr_generated_at"])
	})

	t.Run("update QR code increments count", func(t *testing.T) {
		ch.UpdateWAHAQRCode("QR1")
		assert.Equal(t, 1, ch.GetWAHAQRCodeCount())

		ch.UpdateWAHAQRCode("QR2")
		assert.Equal(t, 2, ch.GetWAHAQRCodeCount())

		ch.UpdateWAHAQRCode("QR3")
		assert.Equal(t, 3, ch.GetWAHAQRCodeCount())
	})

	t.Run("needs new QR code", func(t *testing.T) {
		// Clear everything first
		ch.ClearWAHAQRCode()

		// Needs QR code when status is SCAN_QR and no valid QR
		ch.SetWAHASessionStatus(WAHASessionStatusScanQR)
		assert.True(t, ch.NeedsNewQRCode())

		// Doesn't need when has valid QR
		ch.UpdateWAHAQRCode("NEW_QR")
		assert.False(t, ch.NeedsNewQRCode())

		// Doesn't need when status is WORKING
		ch.SetWAHASessionStatus(WAHASessionStatusWorking)
		assert.False(t, ch.NeedsNewQRCode())
	})
}

func TestChannel_WAHAImport(t *testing.T) {
	ch := createWAHATestChannel(t)

	t.Run("import strategy", func(t *testing.T) {
		config := WAHAConfig{
			BaseURL: "https://waha.example.com",
			Auth: WAHAAuth{
				APIKey: "test-key",
			},
			SessionID:      "session-123",
			ImportStrategy: WAHAImportAll,
		}

		err := ch.SetWAHAConfig(config)
		require.NoError(t, err)

		assert.Equal(t, WAHAImportAll, ch.GetWAHAImportStrategy())
	})

	t.Run("import completion", func(t *testing.T) {
		assert.False(t, ch.IsWAHAImportCompleted())
		assert.True(t, ch.NeedsHistoryImport())

		ch.SetWAHAImportCompleted()

		assert.True(t, ch.IsWAHAImportCompleted())
		assert.False(t, ch.NeedsHistoryImport())
	})

	t.Run("no import needed for 'none' strategy", func(t *testing.T) {
		config := WAHAConfig{
			BaseURL: "https://waha.example.com",
			Auth: WAHAAuth{
				APIKey: "test-key",
			},
			SessionID:      "session-123",
			ImportStrategy: WAHAImportNone,
		}

		err := ch.SetWAHAConfig(config)
		require.NoError(t, err)

		assert.False(t, ch.NeedsHistoryImport())
	})
}

func TestChannel_AssociatePipeline(t *testing.T) {
	ch := createTestChannel(t)
	pipelineID := uuid.New()

	t.Run("associate pipeline", func(t *testing.T) {
		err := ch.AssociatePipeline(pipelineID)
		require.NoError(t, err)

		assert.True(t, ch.HasPipeline())
		assert.NotNil(t, ch.PipelineID)
		assert.Equal(t, pipelineID, *ch.PipelineID)

		events := ch.DomainEvents()
		found := false
		for _, event := range events {
			if assocEvent, ok := event.(ChannelPipelineAssociatedEvent); ok {
				found = true
				assert.Equal(t, ch.ID, assocEvent.ChannelID)
				assert.Equal(t, pipelineID, assocEvent.PipelineID)
			}
		}
		assert.True(t, found, "ChannelPipelineAssociatedEvent should be emitted")
	})

	t.Run("cannot associate nil pipeline", func(t *testing.T) {
		err := ch.AssociatePipeline(uuid.Nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestChannel_DisassociatePipeline(t *testing.T) {
	ch := createTestChannel(t)
	pipelineID := uuid.New()

	err := ch.AssociatePipeline(pipelineID)
	require.NoError(t, err)
	ch.ClearEvents()

	ch.DisassociatePipeline()

	assert.False(t, ch.HasPipeline())
	assert.Nil(t, ch.PipelineID)

	events := ch.DomainEvents()
	assert.Len(t, events, 1)
	disassocEvent, ok := events[0].(ChannelPipelineDisassociatedEvent)
	assert.True(t, ok)
	assert.Equal(t, ch.ID, disassocEvent.ChannelID)
	assert.Equal(t, pipelineID, disassocEvent.PipelineID)
}

func TestChannel_SetDefaultTimeout(t *testing.T) {
	ch := createTestChannel(t)

	t.Run("valid timeout", func(t *testing.T) {
		err := ch.SetDefaultTimeout(60)
		require.NoError(t, err)
		assert.Equal(t, 60, ch.DefaultSessionTimeoutMinutes)
	})

	t.Run("timeout too low", func(t *testing.T) {
		err := ch.SetDefaultTimeout(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "greater than 0")
	})

	t.Run("timeout too high", func(t *testing.T) {
		err := ch.SetDefaultTimeout(1500) // More than 24 hours
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed")
	})
}

func TestChannel_AIFeatures(t *testing.T) {
	ch := createTestChannel(t)

	t.Run("AI processing disabled by default", func(t *testing.T) {
		assert.False(t, ch.ShouldProcessAI())
		assert.False(t, ch.AIEnabled)
		assert.False(t, ch.AIAgentsEnabled)
	})

	t.Run("enable AI processing", func(t *testing.T) {
		ch.AIEnabled = true
		assert.True(t, ch.ShouldProcessAI())
	})

	t.Run("enable AI agents", func(t *testing.T) {
		ch.AIEnabled = true
		ch.AIAgentsEnabled = true
		assert.True(t, ch.AIAgentsEnabled)
	})
}

func TestChannel_EventManagement(t *testing.T) {
	ch := createTestChannel(t)

	t.Run("domain events are collected", func(t *testing.T) {
		events := ch.DomainEvents()
		assert.Len(t, events, 1) // ChannelCreatedEvent
	})

	t.Run("clear events", func(t *testing.T) {
		ch.ClearEvents()
		events := ch.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestChannel_IsWAHA(t *testing.T) {
	wahaChannel := createWAHATestChannel(t)
	whatsappChannel, _ := NewChannel(uuid.New(), uuid.New(), "tenant", "WhatsApp", TypeWhatsApp)

	assert.True(t, wahaChannel.IsWAHA())
	assert.False(t, whatsappChannel.IsWAHA())
}

func TestSetWAHAConfig_ValidationErrors(t *testing.T) {
	ch := createWAHATestChannel(t)

	tests := []struct {
		name        string
		config      WAHAConfig
		errContains string
	}{
		{
			name: "missing base URL",
			config: WAHAConfig{
				BaseURL: "",
				Auth: WAHAAuth{
					APIKey: "key",
				},
				SessionID: "session",
			},
			errContains: "base URL is required",
		},
		{
			name: "missing auth",
			config: WAHAConfig{
				BaseURL: "https://waha.example.com",
				Auth: WAHAAuth{
					APIKey: "",
					Token:  "",
				},
				SessionID: "session",
			},
			errContains: "authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ch.SetWAHAConfig(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestSetWhatsAppConfig_ValidationErrors(t *testing.T) {
	ch, _ := NewChannel(uuid.New(), uuid.New(), "tenant", "WhatsApp", TypeWhatsApp)

	tests := []struct {
		name        string
		config      WhatsAppConfig
		errContains string
	}{
		{
			name: "missing access token",
			config: WhatsAppConfig{
				AccessToken:   "",
				PhoneNumberID: "123",
			},
			errContains: "access token is required",
		},
		{
			name: "missing phone number ID",
			config: WhatsAppConfig{
				AccessToken:   "token",
				PhoneNumberID: "",
			},
			errContains: "phone number ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ch.SetWhatsAppConfig(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestSetTelegramConfig_ValidationErrors(t *testing.T) {
	ch, _ := NewChannel(uuid.New(), uuid.New(), "tenant", "Telegram", TypeTelegram)

	tests := []struct {
		name        string
		config      TelegramConfig
		errContains string
	}{
		{
			name: "missing bot token",
			config: TelegramConfig{
				BotToken: "",
				BotID:    "123",
			},
			errContains: "bot token is required",
		},
		{
			name: "missing bot ID",
			config: TelegramConfig{
				BotToken: "token",
				BotID:    "",
			},
			errContains: "bot ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ch.SetTelegramConfig(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestChannel_ConfigTypeMismatch(t *testing.T) {
	t.Run("set WAHA config on non-WAHA channel", func(t *testing.T) {
		ch, _ := NewChannel(uuid.New(), uuid.New(), "tenant", "WhatsApp", TypeWhatsApp)
		config := WAHAConfig{
			BaseURL: "https://waha.example.com",
			Auth:    WAHAAuth{APIKey: "key"},
		}

		err := ch.SetWAHAConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not WAHA type")
	})

	t.Run("get WAHA config on non-WAHA channel", func(t *testing.T) {
		ch, _ := NewChannel(uuid.New(), uuid.New(), "tenant", "WhatsApp", TypeWhatsApp)

		config, err := ch.GetWAHAConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

// Helper functions
func createTestChannel(t *testing.T) *Channel {
	ch, err := NewChannel(
		uuid.New(),
		uuid.New(),
		"tenant-123",
		"Test Channel",
		TypeWAHA,
	)
	require.NoError(t, err)
	return ch
}

func createWAHATestChannel(t *testing.T) *Channel {
	config := WAHAConfig{
		BaseURL: "https://waha.example.com",
		Auth: WAHAAuth{
			APIKey: "test-api-key",
		},
		SessionID:      "session-123",
		WebhookURL:     "https://app.example.com/webhooks/waha",
		ImportStrategy: WAHAImportNewOnly,
	}

	ch, err := NewWAHAChannel(
		uuid.New(),
		uuid.New(),
		"tenant-123",
		"WAHA Test Channel",
		config,
	)
	require.NoError(t, err)
	return ch
}
