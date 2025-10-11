package credential

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Encryptor for testing
type mockEncryptor struct{}

func (m *mockEncryptor) Encrypt(plaintext string) (EncryptedValue, error) {
	return NewEncryptedValue("encrypted_"+plaintext, "nonce_123"), nil
}

func (m *mockEncryptor) Decrypt(encrypted EncryptedValue) (string, error) {
	// Remove "encrypted_" prefix for decryption
	ciphertext := encrypted.Ciphertext()
	if len(ciphertext) > 10 {
		return ciphertext[10:], nil
	}
	return ciphertext, nil
}

func TestNewCredential_Valid(t *testing.T) {
	encryptor := &mockEncryptor{}
	tenantID := "tenant-123"
	name := "Meta WhatsApp Token"
	plainValue := "secret-token-value"

	tests := []struct {
		name           string
		credentialType CredentialType
	}{
		{
			name:           "Meta WhatsApp credential",
			credentialType: CredentialTypeMetaWhatsApp,
		},
		{
			name:           "Meta Ads credential",
			credentialType: CredentialTypeMetaAds,
		},
		{
			name:           "API Key credential",
			credentialType: CredentialTypeAPIKey,
		},
		{
			name:           "WAHA credential",
			credentialType: CredentialTypeWAHA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := NewCredential(tenantID, tt.credentialType, name, plainValue, encryptor)
			require.NoError(t, err)
			assert.NotNil(t, cred)
			assert.NotEqual(t, uuid.Nil, cred.ID())
			assert.Equal(t, tenantID, cred.TenantID())
			assert.Equal(t, tt.credentialType, cred.Type())
			assert.Equal(t, name, cred.Name())
			assert.True(t, cred.IsActive())
			assert.NotZero(t, cred.CreatedAt())
			assert.NotZero(t, cred.UpdatedAt())

			// Verify encryption worked
			assert.False(t, cred.EncryptedValue().IsEmpty())

			// Verify event was emitted
			events := cred.DomainEvents()
			assert.Len(t, events, 1)
			createdEvent, ok := events[0].(CredentialCreatedEvent)
			assert.True(t, ok)
			assert.Equal(t, cred.ID(), createdEvent.CredentialID)
		})
	}
}

func TestNewCredential_Invalid(t *testing.T) {
	encryptor := &mockEncryptor{}

	tests := []struct {
		name           string
		tenantID       string
		credentialType CredentialType
		credName       string
		plainValue     string
		expectErr      string
	}{
		{
			name:           "empty tenant ID",
			tenantID:       "",
			credentialType: CredentialTypeAPIKey,
			credName:       "Test",
			plainValue:     "value",
			expectErr:      "tenantID cannot be empty",
		},
		{
			name:           "empty name",
			tenantID:       "tenant-123",
			credentialType: CredentialTypeAPIKey,
			credName:       "",
			plainValue:     "value",
			expectErr:      "name cannot be empty",
		},
		{
			name:           "empty value",
			tenantID:       "tenant-123",
			credentialType: CredentialTypeAPIKey,
			credName:       "Test",
			plainValue:     "",
			expectErr:      "value cannot be empty",
		},
		{
			name:           "invalid credential type",
			tenantID:       "tenant-123",
			credentialType: CredentialType("invalid"),
			credName:       "Test",
			plainValue:     "value",
			expectErr:      "invalid credential type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := NewCredential(tt.tenantID, tt.credentialType, tt.credName, tt.plainValue, encryptor)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
			assert.Nil(t, cred)
		})
	}
}

func TestCredential_SetOAuthToken(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeMetaWhatsApp, encryptor)

	t.Run("set OAuth token on OAuth credential type", func(t *testing.T) {
		accessToken := "access_token_123"
		refreshToken := "refresh_token_123"
		expiresIn := 3600

		err := cred.SetOAuthToken(accessToken, refreshToken, expiresIn, encryptor)
		require.NoError(t, err)

		assert.NotNil(t, cred.OAuthToken())
		assert.NotNil(t, cred.ExpiresAt())

		// Verify tokens can be decrypted
		decryptedAccess, err := cred.GetAccessToken(encryptor)
		require.NoError(t, err)
		assert.Equal(t, accessToken, decryptedAccess)

		decryptedRefresh, err := cred.GetRefreshToken(encryptor)
		require.NoError(t, err)
		assert.Equal(t, refreshToken, decryptedRefresh)
	})

	t.Run("cannot set OAuth token on non-OAuth credential type", func(t *testing.T) {
		apiKeyCred := createTestCredential(t, CredentialTypeAPIKey, encryptor)

		err := apiKeyCred.SetOAuthToken("token", "refresh", 3600, encryptor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OAuth tokens only valid for OAuth credential types")
	})
}

func TestCredential_RefreshOAuthToken(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeMetaAds, encryptor)

	t.Run("refresh OAuth token", func(t *testing.T) {
		// Set initial token
		err := cred.SetOAuthToken("initial_token", "refresh_token", 3600, encryptor)
		require.NoError(t, err)
		cred.ClearEvents()

		// Refresh
		newAccessToken := "new_access_token"
		expiresIn := 7200

		err = cred.RefreshOAuthToken(newAccessToken, expiresIn, encryptor)
		require.NoError(t, err)

		// Verify new token
		decryptedAccess, err := cred.GetAccessToken(encryptor)
		require.NoError(t, err)
		assert.Equal(t, newAccessToken, decryptedAccess)

		// Verify event was emitted
		events := cred.DomainEvents()
		assert.Len(t, events, 1)
		_, ok := events[0].(OAuthTokenRefreshedEvent)
		assert.True(t, ok)
	})

	t.Run("cannot refresh without existing token", func(t *testing.T) {
		freshCred := createTestCredential(t, CredentialTypeMetaAds, encryptor)

		err := freshCred.RefreshOAuthToken("new_token", 3600, encryptor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no OAuth token to refresh")
	})
}

func TestCredential_UpdateValue(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)
	cred.ClearEvents()

	t.Run("update value", func(t *testing.T) {
		newValue := "new-secret-value"

		err := cred.UpdateValue(newValue, encryptor)
		require.NoError(t, err)

		// Verify new value
		decrypted, err := cred.Decrypt(encryptor)
		require.NoError(t, err)
		assert.Equal(t, newValue, decrypted)

		// Verify event
		events := cred.DomainEvents()
		assert.Len(t, events, 1)
		_, ok := events[0].(CredentialUpdatedEvent)
		assert.True(t, ok)
	})

	t.Run("cannot update with empty value", func(t *testing.T) {
		err := cred.UpdateValue("", encryptor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value cannot be empty")
	})
}

func TestCredential_Activate_Deactivate(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)

	t.Run("deactivate credential", func(t *testing.T) {
		assert.True(t, cred.IsActive())
		cred.ClearEvents()

		cred.Deactivate()

		assert.False(t, cred.IsActive())

		events := cred.DomainEvents()
		assert.Len(t, events, 1)
		_, ok := events[0].(CredentialDeactivatedEvent)
		assert.True(t, ok)
	})

	t.Run("activate credential", func(t *testing.T) {
		cred.ClearEvents()

		cred.Activate()

		assert.True(t, cred.IsActive())

		events := cred.DomainEvents()
		assert.Len(t, events, 1)
		_, ok := events[0].(CredentialActivatedEvent)
		assert.True(t, ok)
	})

	t.Run("deactivating twice doesn't emit duplicate events", func(t *testing.T) {
		cred.Deactivate()
		cred.ClearEvents()

		cred.Deactivate()

		events := cred.DomainEvents()
		assert.Len(t, events, 0) // No new event
	})
}

func TestCredential_IsExpired(t *testing.T) {
	encryptor := &mockEncryptor{}

	t.Run("credential without expiration never expires", func(t *testing.T) {
		cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)
		assert.False(t, cred.IsExpired())
	})

	t.Run("credential with future expiration not expired", func(t *testing.T) {
		cred := createTestCredential(t, CredentialTypeMetaWhatsApp, encryptor)
		err := cred.SetOAuthToken("token", "refresh", 3600, encryptor)
		require.NoError(t, err)

		assert.False(t, cred.IsExpired())
	})

	t.Run("credential with past expiration is expired", func(t *testing.T) {
		cred := createTestCredential(t, CredentialTypeMetaWhatsApp, encryptor)

		// Set token that already expired
		pastTime := time.Now().Add(-1 * time.Hour)
		token, _ := NewOAuthToken("token", "refresh", pastTime, encryptor)
		cred.oauthToken = token
		cred.expiresAt = &pastTime

		assert.True(t, cred.IsExpired())
	})
}

func TestCredential_NeedsRefresh(t *testing.T) {
	encryptor := &mockEncryptor{}

	t.Run("credential without OAuth token doesn't need refresh", func(t *testing.T) {
		cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)
		assert.False(t, cred.NeedsRefresh())
	})

	t.Run("token expiring soon needs refresh", func(t *testing.T) {
		cred := createTestCredential(t, CredentialTypeMetaAds, encryptor)

		// Token expiring in 20 minutes (< 30 min threshold)
		err := cred.SetOAuthToken("token", "refresh", 20*60, encryptor)
		require.NoError(t, err)

		assert.True(t, cred.NeedsRefresh())
	})

	t.Run("token with plenty of time doesn't need refresh", func(t *testing.T) {
		cred := createTestCredential(t, CredentialTypeMetaAds, encryptor)

		// Token expiring in 2 hours
		err := cred.SetOAuthToken("token", "refresh", 2*60*60, encryptor)
		require.NoError(t, err)

		assert.False(t, cred.NeedsRefresh())
	})
}

func TestCredential_MarkAsUsed(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)
	cred.ClearEvents()

	assert.Nil(t, cred.LastUsedAt())

	cred.MarkAsUsed()

	assert.NotNil(t, cred.LastUsedAt())

	events := cred.DomainEvents()
	assert.Len(t, events, 1)
	usedEvent, ok := events[0].(CredentialUsedEvent)
	assert.True(t, ok)
	assert.Equal(t, cred.ID(), usedEvent.CredentialID)
}

func TestCredential_Metadata(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)

	t.Run("set and get metadata", func(t *testing.T) {
		cred.SetMetadata("key1", "value1")
		cred.SetMetadata("key2", 123)
		cred.SetMetadata("key3", true)

		val1, exists := cred.GetMetadata("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", val1)

		val2, exists := cred.GetMetadata("key2")
		assert.True(t, exists)
		assert.Equal(t, 123, val2)

		val3, exists := cred.GetMetadata("key3")
		assert.True(t, exists)
		assert.Equal(t, true, val3)
	})

	t.Run("get non-existent metadata", func(t *testing.T) {
		val, exists := cred.GetMetadata("non-existent")
		assert.False(t, exists)
		assert.Nil(t, val)
	})

	t.Run("metadata returns copy", func(t *testing.T) {
		metadata := cred.Metadata()
		metadata["new_key"] = "should not affect original"

		_, exists := cred.GetMetadata("new_key")
		assert.False(t, exists)
	})
}

func TestCredential_SetProjectID(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)

	assert.Nil(t, cred.ProjectID())

	projectID := uuid.New()
	cred.SetProjectID(projectID)

	assert.NotNil(t, cred.ProjectID())
	assert.Equal(t, projectID, *cred.ProjectID())
}

func TestCredential_UpdateDescription(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)

	assert.Empty(t, cred.Description())

	description := "This is a test credential for API access"
	cred.UpdateDescription(description)

	assert.Equal(t, description, cred.Description())
}

func TestCredential_Decrypt(t *testing.T) {
	encryptor := &mockEncryptor{}
	originalValue := "secret-api-key-123"
	cred := createTestCredentialWithValue(t, CredentialTypeAPIKey, originalValue, encryptor)

	decrypted, err := cred.Decrypt(encryptor)
	require.NoError(t, err)
	assert.Equal(t, originalValue, decrypted)
}

func TestCredential_EventManagement(t *testing.T) {
	encryptor := &mockEncryptor{}
	cred := createTestCredential(t, CredentialTypeAPIKey, encryptor)

	t.Run("domain events are collected", func(t *testing.T) {
		events := cred.DomainEvents()
		assert.Len(t, events, 1) // CredentialCreatedEvent
	})

	t.Run("clear events", func(t *testing.T) {
		cred.ClearEvents()
		events := cred.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestReconstructCredential(t *testing.T) {
	id := uuid.New()
	tenantID := "tenant-123"
	projectID := uuid.New()
	credType := CredentialTypeMetaWhatsApp
	name := "Test Credential"
	description := "Test Description"
	encryptedValue := NewEncryptedValue("cipher", "nonce")
	metadata := map[string]interface{}{"key": "value"}
	now := time.Now()
	expiresAt := time.Now().Add(1 * time.Hour)

	cred := ReconstructCredential(
		id,
		tenantID,
		&projectID,
		credType,
		name,
		description,
		encryptedValue,
		nil,
		metadata,
		true,
		&expiresAt,
		nil,
		now,
		now,
	)

	assert.NotNil(t, cred)
	assert.Equal(t, id, cred.ID())
	assert.Equal(t, tenantID, cred.TenantID())
	assert.Equal(t, projectID, *cred.ProjectID())
	assert.Equal(t, credType, cred.Type())
	assert.Equal(t, name, cred.Name())
	assert.Equal(t, description, cred.Description())
	assert.True(t, cred.IsActive())
	assert.Equal(t, expiresAt, *cred.ExpiresAt())
	assert.Equal(t, now, cred.CreatedAt())
	assert.Equal(t, now, cred.UpdatedAt())

	// Events should be empty for reconstructed entities
	events := cred.DomainEvents()
	assert.Len(t, events, 0)
}

// Helper functions
func createTestCredential(t *testing.T, credType CredentialType, encryptor Encryptor) *Credential {
	return createTestCredentialWithValue(t, credType, "test-secret-value", encryptor)
}

func createTestCredentialWithValue(t *testing.T, credType CredentialType, value string, encryptor Encryptor) *Credential {
	cred, err := NewCredential(
		"tenant-123",
		credType,
		"Test Credential",
		value,
		encryptor,
	)
	require.NoError(t, err)
	return cred
}
