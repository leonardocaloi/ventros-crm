package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTriggerRegistry(t *testing.T) {
	registry := NewTriggerRegistry()

	assert.NotNil(t, registry)

	// Should have system triggers registered
	systemTriggers := registry.ListSystemTriggers()
	assert.NotEmpty(t, systemTriggers)
	assert.GreaterOrEqual(t, len(systemTriggers), 10)
}

func TestTriggerRegistry_RegisterSystemTriggers(t *testing.T) {
	registry := NewTriggerRegistry()

	// Verify all expected system triggers are registered
	expectedTriggers := []AutomationTrigger{
		TriggerSessionEnded,
		TriggerSessionTimeout,
		TriggerSessionResolved,
		TriggerSessionEscalated,
		TriggerNoResponse,
		TriggerMessageReceived,
		TriggerStatusChanged,
		TriggerStageCompleted,
		TriggerAfterDelay,
		TriggerScheduled,
		TriggerPurchaseCompleted,
		TriggerPaymentReceived,
		TriggerRefundIssued,
		TriggerCartAbandoned,
		TriggerOrderShipped,
		TriggerPageVisited,
		TriggerFormSubmitted,
		TriggerFileDownloaded,
	}

	for _, trigger := range expectedTriggers {
		code := string(trigger)
		assert.True(t, registry.IsValidTrigger(code), "Expected %s to be valid", code)

		metadata, err := registry.GetTrigger(code)
		require.NoError(t, err)
		assert.Equal(t, code, metadata.Code)
		assert.NotEmpty(t, metadata.Name)
		assert.NotEmpty(t, metadata.Description)
		assert.True(t, metadata.IsSystem)
	}
}

func TestTriggerRegistry_RegisterCustomTrigger(t *testing.T) {
	registry := NewTriggerRegistry()

	t.Run("register valid custom trigger", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code:        "custom.loyalty_points_reached",
			Name:        "Pontos de Fidelidade Atingidos",
			Description: "Disparado quando cliente atinge X pontos",
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "points", Type: "int"},
			},
		}

		err := registry.RegisterCustomTrigger(customTrigger)
		require.NoError(t, err)

		// Verify it was registered
		assert.True(t, registry.IsValidTrigger("custom.loyalty_points_reached"))

		metadata, err := registry.GetTrigger("custom.loyalty_points_reached")
		require.NoError(t, err)
		assert.Equal(t, "custom.loyalty_points_reached", metadata.Code)
		assert.False(t, metadata.IsSystem)
		assert.Equal(t, CategoryCustom, metadata.Category)
	})

	t.Run("empty code", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: "",
			Name: "Test",
		}

		err := registry.RegisterCustomTrigger(customTrigger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "code cannot be empty")
	})

	t.Run("override system trigger", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: string(TriggerSessionEnded),
			Name: "My Custom Trigger",
		}

		err := registry.RegisterCustomTrigger(customTrigger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot override system trigger")
	})

	t.Run("missing custom prefix", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: "my_trigger",
			Name: "My Trigger",
		}

		err := registry.RegisterCustomTrigger(customTrigger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must start with 'custom.' prefix")
	})
}

func TestTriggerRegistry_UnregisterCustomTrigger(t *testing.T) {
	registry := NewTriggerRegistry()

	// Register a custom trigger
	customTrigger := TriggerMetadata{
		Code: "custom.test_trigger",
		Name: "Test Trigger",
	}
	err := registry.RegisterCustomTrigger(customTrigger)
	require.NoError(t, err)

	t.Run("unregister custom trigger", func(t *testing.T) {
		err := registry.UnregisterCustomTrigger("custom.test_trigger")
		require.NoError(t, err)

		// Verify it was removed
		assert.False(t, registry.IsValidTrigger("custom.test_trigger"))
	})

	t.Run("cannot unregister system trigger", func(t *testing.T) {
		err := registry.UnregisterCustomTrigger(string(TriggerSessionEnded))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot unregister system trigger")
	})

	t.Run("unregister non-existent trigger", func(t *testing.T) {
		// Should not error, just silently remove from map
		err := registry.UnregisterCustomTrigger("custom.non_existent")
		require.NoError(t, err)
	})
}

func TestTriggerRegistry_IsValidTrigger(t *testing.T) {
	registry := NewTriggerRegistry()

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"system trigger", string(TriggerSessionEnded), true},
		{"another system trigger", string(TriggerStatusChanged), true},
		{"invalid trigger", "invalid.trigger", false},
		{"empty code", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsValidTrigger(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("custom trigger after registration", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: "custom.my_trigger",
			Name: "My Trigger",
		}
		registry.RegisterCustomTrigger(customTrigger)

		assert.True(t, registry.IsValidTrigger("custom.my_trigger"))
	})
}

func TestTriggerRegistry_GetTrigger(t *testing.T) {
	registry := NewTriggerRegistry()

	t.Run("get system trigger", func(t *testing.T) {
		metadata, err := registry.GetTrigger(string(TriggerSessionEnded))
		require.NoError(t, err)
		assert.Equal(t, string(TriggerSessionEnded), metadata.Code)
		assert.Equal(t, "Sessão Encerrada", metadata.Name)
		assert.True(t, metadata.IsSystem)
		assert.Equal(t, CategorySession, metadata.Category)
		assert.NotEmpty(t, metadata.Parameters)
	})

	t.Run("get custom trigger", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: "custom.vip_customer",
			Name: "VIP Customer",
		}
		registry.RegisterCustomTrigger(customTrigger)

		metadata, err := registry.GetTrigger("custom.vip_customer")
		require.NoError(t, err)
		assert.Equal(t, "custom.vip_customer", metadata.Code)
		assert.False(t, metadata.IsSystem)
	})

	t.Run("trigger not found", func(t *testing.T) {
		_, err := registry.GetTrigger("non_existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "trigger not found")
	})
}

func TestTriggerRegistry_ListSystemTriggers(t *testing.T) {
	registry := NewTriggerRegistry()

	triggers := registry.ListSystemTriggers()
	assert.NotEmpty(t, triggers)
	assert.GreaterOrEqual(t, len(triggers), 10)

	// Verify all are system triggers
	for _, trigger := range triggers {
		assert.True(t, trigger.IsSystem)
		assert.NotEmpty(t, trigger.Code)
		assert.NotEmpty(t, trigger.Name)
		assert.NotEmpty(t, trigger.Description)
	}
}

func TestTriggerRegistry_ListCustomTriggers(t *testing.T) {
	registry := NewTriggerRegistry()

	// Initially empty
	triggers := registry.ListCustomTriggers()
	assert.Empty(t, triggers)

	// Register some custom triggers
	custom1 := TriggerMetadata{Code: "custom.trigger1", Name: "Trigger 1"}
	custom2 := TriggerMetadata{Code: "custom.trigger2", Name: "Trigger 2"}

	registry.RegisterCustomTrigger(custom1)
	registry.RegisterCustomTrigger(custom2)

	triggers = registry.ListCustomTriggers()
	assert.Len(t, triggers, 2)

	// Verify all are custom triggers
	for _, trigger := range triggers {
		assert.False(t, trigger.IsSystem)
		assert.Equal(t, CategoryCustom, trigger.Category)
	}
}

func TestTriggerRegistry_ListAllTriggers(t *testing.T) {
	registry := NewTriggerRegistry()

	// Register some custom triggers
	custom1 := TriggerMetadata{Code: "custom.trigger1", Name: "Trigger 1"}
	custom2 := TriggerMetadata{Code: "custom.trigger2", Name: "Trigger 2"}

	registry.RegisterCustomTrigger(custom1)
	registry.RegisterCustomTrigger(custom2)

	triggers := registry.ListAllTriggers()

	// Should include both system and custom
	assert.GreaterOrEqual(t, len(triggers), 12) // at least 10 system + 2 custom

	// Count system vs custom
	systemCount := 0
	customCount := 0
	for _, trigger := range triggers {
		if trigger.IsSystem {
			systemCount++
		} else {
			customCount++
		}
	}

	assert.GreaterOrEqual(t, systemCount, 10)
	assert.Equal(t, 2, customCount)
}

func TestTriggerRegistry_ListTriggersByCategory(t *testing.T) {
	registry := NewTriggerRegistry()

	tests := []struct {
		name         string
		category     TriggerCategory
		minExpected  int
	}{
		{"session triggers", CategorySession, 4},
		{"message triggers", CategoryMessage, 2},
		{"pipeline triggers", CategoryPipeline, 2},
		{"temporal triggers", CategoryTemporal, 2},
		{"transaction triggers", CategoryTransaction, 5},
		{"behavior triggers", CategoryBehavior, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggers := registry.ListTriggersByCategory(tt.category)
			assert.GreaterOrEqual(t, len(triggers), tt.minExpected,
				"Expected at least %d triggers in category %s", tt.minExpected, tt.category)

			// Verify all belong to the correct category
			for _, trigger := range triggers {
				assert.Equal(t, tt.category, trigger.Category)
			}
		})
	}

	t.Run("custom category with custom triggers", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: "custom.test",
			Name: "Test",
		}
		registry.RegisterCustomTrigger(customTrigger)

		triggers := registry.ListTriggersByCategory(CategoryCustom)
		assert.Len(t, triggers, 1)
		assert.Equal(t, "custom.test", triggers[0].Code)
	})

	t.Run("empty category", func(t *testing.T) {
		triggers := registry.ListTriggersByCategory("nonexistent")
		assert.Empty(t, triggers)
	})
}

func TestTriggerRegistry_GetParametersForTrigger(t *testing.T) {
	registry := NewTriggerRegistry()

	t.Run("get parameters for session ended", func(t *testing.T) {
		params, err := registry.GetParametersForTrigger(string(TriggerSessionEnded))
		require.NoError(t, err)
		assert.NotEmpty(t, params)

		// Check that expected parameters are present
		paramNames := make(map[string]bool)
		for _, param := range params {
			paramNames[param.Name] = true
			assert.NotEmpty(t, param.Type)
		}

		assert.True(t, paramNames["session_id"])
		assert.True(t, paramNames["contact_id"])
		assert.True(t, paramNames["message_count"])
	})

	t.Run("get parameters for custom trigger", func(t *testing.T) {
		customTrigger := TriggerMetadata{
			Code: "custom.points",
			Name: "Points Trigger",
			Parameters: []TriggerParameter{
				{Name: "points", Type: "int", Description: "Total points"},
				{Name: "level", Type: "string", Description: "Customer level"},
			},
		}
		registry.RegisterCustomTrigger(customTrigger)

		params, err := registry.GetParametersForTrigger("custom.points")
		require.NoError(t, err)
		assert.Len(t, params, 2)
	})

	t.Run("trigger not found", func(t *testing.T) {
		_, err := registry.GetParametersForTrigger("invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "trigger not found")
	})
}

func TestTriggerRegistry_VerifyTriggerMetadata(t *testing.T) {
	registry := NewTriggerRegistry()

	// Verify specific triggers have correct metadata
	tests := []struct {
		code        string
		name        string
		category    TriggerCategory
		minParams   int
	}{
		{string(TriggerSessionEnded), "Sessão Encerrada", CategorySession, 4},
		{string(TriggerSessionTimeout), "Sessão Expirou", CategorySession, 3},
		{string(TriggerStatusChanged), "Status Mudou", CategoryPipeline, 3},
		{string(TriggerPurchaseCompleted), "Compra Concluída", CategoryTransaction, 4},
		{string(TriggerPageVisited), "Página Visitada", CategoryBehavior, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := registry.GetTrigger(tt.code)
			require.NoError(t, err)

			assert.Equal(t, tt.name, metadata.Name)
			assert.Equal(t, tt.category, metadata.Category)
			assert.GreaterOrEqual(t, len(metadata.Parameters), tt.minParams)
			assert.True(t, metadata.IsSystem)
			assert.NotEmpty(t, metadata.Description)
		})
	}
}

func TestTriggerRegistry_Concurrency(t *testing.T) {
	registry := NewTriggerRegistry()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				registry.IsValidTrigger(string(TriggerSessionEnded))
				registry.ListSystemTriggers()
				registry.GetTrigger(string(TriggerStatusChanged))
			}
			done <- true
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				customTrigger := TriggerMetadata{
					Code: "custom.concurrent_test",
					Name: "Concurrent Test",
				}
				registry.RegisterCustomTrigger(customTrigger)
				registry.UnregisterCustomTrigger("custom.concurrent_test")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Verify registry still works correctly
	assert.True(t, registry.IsValidTrigger(string(TriggerSessionEnded)))
	triggers := registry.ListSystemTriggers()
	assert.NotEmpty(t, triggers)
}

func TestTriggerCategory_Constants(t *testing.T) {
	// Verify category constants exist and have correct values
	assert.Equal(t, TriggerCategory("session"), CategorySession)
	assert.Equal(t, TriggerCategory("message"), CategoryMessage)
	assert.Equal(t, TriggerCategory("pipeline"), CategoryPipeline)
	assert.Equal(t, TriggerCategory("temporal"), CategoryTemporal)
	assert.Equal(t, TriggerCategory("transaction"), CategoryTransaction)
	assert.Equal(t, TriggerCategory("behavior"), CategoryBehavior)
	assert.Equal(t, TriggerCategory("custom"), CategoryCustom)
	assert.Equal(t, TriggerCategory("webhook"), CategoryWebhook)
}
