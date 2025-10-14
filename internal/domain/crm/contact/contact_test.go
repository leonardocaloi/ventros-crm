package contact

import (
	"testing"
	"time"

	domain "github.com/ventros/crm/internal/domain/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================
// 1.2.1 - Testes de Factory Method
// ===========================

func TestNewContact_Success(t *testing.T) {
	// Arrange
	projectID := domain.NewTestUUID()
	tenantID := domain.NewTestTenantID()
	name := "João Silva"

	// Act
	contact, err := NewContact(projectID, tenantID, name)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, contact)
	assert.NotEqual(t, uuid.Nil, contact.ID())
	assert.Equal(t, projectID, contact.ProjectID())
	assert.Equal(t, tenantID, contact.TenantID())
	assert.Equal(t, name, contact.Name())
	assert.Equal(t, "en", contact.Language()) // default language
	assert.Empty(t, contact.Tags())
	domain.AssertTimeNotZero(t, contact.CreatedAt(), "CreatedAt")
	domain.AssertTimeNotZero(t, contact.UpdatedAt(), "UpdatedAt")
	assert.Nil(t, contact.DeletedAt())
}

func TestNewContact_EmptyProjectID(t *testing.T) {
	// Arrange
	projectID := uuid.Nil
	tenantID := domain.NewTestTenantID()
	name := "João Silva"

	// Act
	contact, err := NewContact(projectID, tenantID, name)

	// Assert
	require.Error(t, err)
	assert.Nil(t, contact)
	assert.Contains(t, err.Error(), "projectID cannot be nil")
}

func TestNewContact_EmptyTenantID(t *testing.T) {
	// Arrange
	projectID := domain.NewTestUUID()
	tenantID := ""
	name := "João Silva"

	// Act
	contact, err := NewContact(projectID, tenantID, name)

	// Assert
	require.Error(t, err)
	assert.Nil(t, contact)
	assert.Contains(t, err.Error(), "tenantID cannot be empty")
}

func TestNewContact_EmptyName(t *testing.T) {
	// Arrange
	projectID := domain.NewTestUUID()
	tenantID := domain.NewTestTenantID()
	name := ""

	// Act
	contact, err := NewContact(projectID, tenantID, name)

	// Assert
	require.Error(t, err)
	assert.Nil(t, contact)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestNewContact_GeneratesEvent(t *testing.T) {
	// Arrange
	projectID := domain.NewTestUUID()
	tenantID := domain.NewTestTenantID()
	name := "João Silva"

	// Act
	contact, err := NewContact(projectID, tenantID, name)

	// Assert
	require.NoError(t, err)
	events := contact.DomainEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(ContactCreatedEvent)
	require.True(t, ok, "Event should be ContactCreatedEvent")
	assert.Equal(t, contact.ID(), event.ContactID)
	assert.Equal(t, projectID, event.ProjectID)
	assert.Equal(t, tenantID, event.TenantID)
	assert.Equal(t, name, event.Name)
	domain.AssertTimeNotZero(t, event.CreatedAt, "Event CreatedAt")
}

// ===========================
// 1.2.2 - Testes de Email Value Object
// ===========================

func TestSetEmail_ValidEmail(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	validEmail := "test@example.com"
	beforeUpdate := time.Now()

	// Act
	err := contact.SetEmail(validEmail)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, contact.Email())
	assert.Equal(t, validEmail, contact.Email().String())
	assert.True(t, contact.UpdatedAt().After(beforeUpdate) || contact.UpdatedAt().Equal(beforeUpdate))
}

func TestSetEmail_InvalidFormat(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	invalidEmail := "not-an-email"

	// Act
	err := contact.SetEmail(invalidEmail)

	// Assert
	require.Error(t, err)
	assert.Nil(t, contact.Email())
}

func TestSetEmail_UpdatesTimestamp(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	originalUpdatedAt := contact.UpdatedAt()
	time.Sleep(10 * time.Millisecond) // Garante diferença no timestamp

	// Act
	err := contact.SetEmail("test@example.com")

	// Assert
	require.NoError(t, err)
	assert.True(t, contact.UpdatedAt().After(originalUpdatedAt))
}

// ===========================
// 1.2.3 - Testes de Phone Value Object
// ===========================

func TestSetPhone_ValidPhone(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	validPhone := "5511999999999"
	beforeUpdate := time.Now()

	// Act
	err := contact.SetPhone(validPhone)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, contact.Phone())
	assert.Equal(t, validPhone, contact.Phone().String())
	assert.True(t, contact.UpdatedAt().After(beforeUpdate) || contact.UpdatedAt().Equal(beforeUpdate))
}

func TestSetPhone_InvalidFormat(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	invalidPhone := "123" // muito curto

	// Act
	err := contact.SetPhone(invalidPhone)

	// Assert
	require.Error(t, err)
	assert.Nil(t, contact.Phone())
}

func TestSetPhone_UpdatesTimestamp(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	originalUpdatedAt := contact.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	// Act
	err := contact.SetPhone("5511999999999")

	// Assert
	require.NoError(t, err)
	assert.True(t, contact.UpdatedAt().After(originalUpdatedAt))
}

// ===========================
// 1.2.4 - Testes de Métodos de Negócio
// ===========================

func TestUpdateName_Success(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Old Name")
	newName := "New Name"
	originalUpdatedAt := contact.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	// Act
	err := contact.UpdateName(newName)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, newName, contact.Name())
	assert.True(t, contact.UpdatedAt().After(originalUpdatedAt))
}

func TestUpdateName_EmptyName(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Old Name")
	originalName := contact.Name()

	// Act
	err := contact.UpdateName("")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
	assert.Equal(t, originalName, contact.Name()) // Nome não deve mudar
}

func TestUpdateName_GeneratesEvent(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Old Name")
	contact.ClearEvents() // Limpa evento de criação
	newName := "New Name"

	// Act
	err := contact.UpdateName(newName)

	// Assert
	require.NoError(t, err)
	events := contact.DomainEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(ContactUpdatedEvent)
	require.True(t, ok, "Event should be ContactUpdatedEvent")
	assert.Equal(t, contact.ID(), event.ContactID)
	domain.AssertTimeNotZero(t, event.UpdatedAt, "Event UpdatedAt")
}

func TestAddTag_NewTag(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	tag := "vip"

	// Act
	contact.AddTag(tag)

	// Assert
	tags := contact.Tags()
	require.Len(t, tags, 1)
	assert.Contains(t, tags, tag)
}

func TestAddTag_DuplicateTag(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	tag := "vip"
	contact.AddTag(tag)

	// Act
	contact.AddTag(tag) // Adiciona novamente

	// Assert
	tags := contact.Tags()
	require.Len(t, tags, 1, "Should not add duplicate tag")
	assert.Contains(t, tags, tag)
}

func TestRemoveTag_ExistingTag(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	tag := "vip"
	contact.AddTag(tag)

	// Act
	contact.RemoveTag(tag)

	// Assert
	tags := contact.Tags()
	assert.Empty(t, tags)
}

func TestRemoveTag_NonExistingTag(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	contact.AddTag("tag1")

	// Act
	contact.RemoveTag("nonexistent") // Não deve causar erro

	// Assert
	tags := contact.Tags()
	require.Len(t, tags, 1)
	assert.Contains(t, tags, "tag1")
}

func TestRecordInteraction_FirstTime(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	assert.Nil(t, contact.FirstInteractionAt())

	// Act
	contact.RecordInteraction()

	// Assert
	require.NotNil(t, contact.FirstInteractionAt())
	require.NotNil(t, contact.LastInteractionAt())
	domain.AssertTimeNotZero(t, *contact.FirstInteractionAt(), "FirstInteractionAt")
	domain.AssertTimeNotZero(t, *contact.LastInteractionAt(), "LastInteractionAt")
}

func TestRecordInteraction_UpdatesLastInteraction(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	contact.RecordInteraction()
	firstInteraction := contact.FirstInteractionAt()
	time.Sleep(10 * time.Millisecond)

	// Act
	contact.RecordInteraction()

	// Assert
	assert.Equal(t, firstInteraction, contact.FirstInteractionAt(), "FirstInteractionAt should not change")
	assert.True(t, contact.LastInteractionAt().After(*firstInteraction), "LastInteractionAt should be updated")
}

// ===========================
// 1.2.5 - Testes de Soft Delete
// ===========================

func TestSoftDelete_Success(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	assert.Nil(t, contact.DeletedAt())
	assert.False(t, contact.IsDeleted())

	// Act
	err := contact.SoftDelete()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, contact.DeletedAt())
	assert.True(t, contact.IsDeleted())
	domain.AssertTimeNotZero(t, *contact.DeletedAt(), "DeletedAt")
}

func TestSoftDelete_AlreadyDeleted(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	_ = contact.SoftDelete()

	// Act
	err := contact.SoftDelete()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already deleted")
}

func TestSoftDelete_GeneratesEvent(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	contact.ClearEvents()

	// Act
	err := contact.SoftDelete()

	// Assert
	require.NoError(t, err)
	events := contact.DomainEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(ContactDeletedEvent)
	require.True(t, ok, "Event should be ContactDeletedEvent")
	assert.Equal(t, contact.ID(), event.ContactID)
	domain.AssertTimeNotZero(t, event.DeletedAt, "Event DeletedAt")
}

func TestIsDeleted_ReturnsTrueAfterDelete(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")
	assert.False(t, contact.IsDeleted())

	// Act
	_ = contact.SoftDelete()

	// Assert
	assert.True(t, contact.IsDeleted())
}

func TestDelete_AliasForSoftDelete(t *testing.T) {
	// Arrange
	contact, _ := NewContact(domain.NewTestUUID(), domain.NewTestTenantID(), "Test")

	// Act
	err := contact.Delete()

	// Assert
	require.NoError(t, err)
	assert.True(t, contact.IsDeleted())
}
