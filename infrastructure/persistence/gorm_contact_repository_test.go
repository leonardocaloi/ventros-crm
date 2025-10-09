package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================
// 2.2 - Testes de GormContactRepository
// ===========================

func TestGormContactRepository_Save_NewContact(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create domain contact
	domainContact, err := contact.NewContact(
		testData.ProjectID,
		testData.TenantID,
		"João Silva",
	)
	require.NoError(t, err)

	domainContact.SetEmail("joao@example.com")
	domainContact.SetPhone("5511999999999")

	originalID := domainContact.ID()

	// Act
	err = repo.Save(ctx, domainContact)

	// Assert
	require.NoError(t, err)

	// Verify in database
	found, err := repo.FindByID(ctx, originalID)
	require.NoError(t, err)
	assert.Equal(t, originalID, found.ID())
	assert.Equal(t, "João Silva", found.Name())
	assert.Equal(t, "joao@example.com", found.Email().String())
	assert.Equal(t, "5511999999999", found.Phone().String())
}

func TestGormContactRepository_Save_UpdateExisting(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create and save contact
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Maria")
	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act - Update contact
	domainContact.UpdateName("Maria Silva")
	domainContact.SetEmail("maria@example.com")

	err = repo.Save(ctx, domainContact)

	// Assert
	require.NoError(t, err)

	// Verify in database
	found, err := repo.FindByID(ctx, domainContact.ID())
	require.NoError(t, err)
	assert.Equal(t, "Maria Silva", found.Name())
	require.NotNil(t, found.Email())
	assert.Equal(t, "maria@example.com", found.Email().String())
}

func TestGormContactRepository_Save_PreservesID(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Pedro")
	originalID := domainContact.ID()

	// Act
	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Assert - ID should not change
	found, _ := repo.FindByID(ctx, originalID)
	assert.Equal(t, originalID, found.ID(), "ID should be preserved after save")
}

func TestGormContactRepository_FindByID_Success(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Ana")
	domainContact.SetEmail("ana@example.com")
	domainContact.SetPhone("5511988888888")
	domainContact.SetLanguage("pt-BR")
	domainContact.AddTag("vip")

	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act
	found, err := repo.FindByID(ctx, domainContact.ID())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domainContact.ID(), found.ID())
	assert.Equal(t, "Ana", found.Name())
	assert.Equal(t, "ana@example.com", found.Email().String())
	assert.Equal(t, "5511988888888", found.Phone().String())
	assert.Equal(t, "pt-BR", found.Language())
	assert.Contains(t, found.Tags(), "vip")
}

func TestGormContactRepository_FindByID_NotFound(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	nonExistentID := uuid.New()

	// Act
	found, err := repo.FindByID(ctx, nonExistentID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, contact.ErrContactNotFound)
	assert.Nil(t, found)
}

func TestGormContactRepository_FindByID_ReconstructsDomain(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact with all fields
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Carlos")
	domainContact.SetEmail("carlos@example.com")
	domainContact.SetPhone("5511977777777")
	domainContact.SetLanguage("es")
	domainContact.AddTag("premium")
	domainContact.AddTag("enterprise")

	domainContact.SetExternalID("ext-123")
	domainContact.SetSourceChannel("whatsapp")
	domainContact.SetTimezone("America/Sao_Paulo")

	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act
	found, err := repo.FindByID(ctx, domainContact.ID())

	// Assert
	require.NoError(t, err)

	// Verify all fields are reconstructed correctly
	assert.Equal(t, domainContact.ID(), found.ID())
	assert.Equal(t, testData.ProjectID, found.ProjectID())
	assert.Equal(t, testData.TenantID, found.TenantID())
	assert.Equal(t, "Carlos", found.Name())

	require.NotNil(t, found.Email())
	assert.Equal(t, "carlos@example.com", found.Email().String())

	require.NotNil(t, found.Phone())
	assert.Equal(t, "5511977777777", found.Phone().String())

	assert.Equal(t, "es", found.Language())

	require.NotNil(t, found.ExternalID())
	assert.Equal(t, "ext-123", *found.ExternalID())

	require.NotNil(t, found.SourceChannel())
	assert.Equal(t, "whatsapp", *found.SourceChannel())

	require.NotNil(t, found.Timezone())
	assert.Equal(t, "America/Sao_Paulo", *found.Timezone())

	assert.Len(t, found.Tags(), 2)
	assert.Contains(t, found.Tags(), "premium")
	assert.Contains(t, found.Tags(), "enterprise")

	// Timestamps should be preserved
	assert.NotZero(t, found.CreatedAt())
	assert.NotZero(t, found.UpdatedAt())

	// Should not be deleted
	assert.Nil(t, found.DeletedAt())
}

func TestGormContactRepository_FindByPhone_Success(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact with phone
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Fernanda")
	domainContact.SetPhone("5511966666666")

	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act
	found, err := repo.FindByPhone(ctx, testData.ProjectID, "5511966666666")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domainContact.ID(), found.ID())
	assert.Equal(t, "Fernanda", found.Name())
	assert.Equal(t, "5511966666666", found.Phone().String())
}

func TestGormContactRepository_FindByPhone_NotFound(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Act
	found, err := repo.FindByPhone(ctx, testData.ProjectID, "5511999999999")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, contact.ErrContactNotFound)
	assert.Nil(t, found)
}

func TestGormContactRepository_FindByPhone_IgnoresDeleted(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create and soft delete contact
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Roberto")
	domainContact.SetPhone("5511955555555")
	repo.Save(ctx, domainContact)

	domainContact.SoftDelete()
	repo.Save(ctx, domainContact)

	// Act
	found, err := repo.FindByPhone(ctx, testData.ProjectID, "5511955555555")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, contact.ErrContactNotFound, "Deleted contacts should not be found")
	assert.Nil(t, found)
}

func TestGormContactRepository_FindByEmail_Success(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact with email
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Lucia")
	domainContact.SetEmail("lucia@example.com")

	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act
	found, err := repo.FindByEmail(ctx, testData.ProjectID, "lucia@example.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domainContact.ID(), found.ID())
	assert.Equal(t, "Lucia", found.Name())
	assert.Equal(t, "lucia@example.com", found.Email().String())
}

func TestGormContactRepository_FindByEmail_NotFound(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Act
	found, err := repo.FindByEmail(ctx, testData.ProjectID, "notfound@example.com")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, contact.ErrContactNotFound)
	assert.Nil(t, found)
}

func TestGormContactRepository_FindByEmail_IgnoresDeleted(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create and soft delete contact
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Paula")
	domainContact.SetEmail("paula@example.com")
	repo.Save(ctx, domainContact)

	domainContact.SoftDelete()
	repo.Save(ctx, domainContact)

	// Act
	found, err := repo.FindByEmail(ctx, testData.ProjectID, "paula@example.com")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, contact.ErrContactNotFound, "Deleted contacts should not be found")
	assert.Nil(t, found)
}

func TestGormContactRepository_FindByExternalID_Success(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact with external ID
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Ricardo")
	domainContact.SetExternalID("ext-456")

	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act
	found, err := repo.FindByExternalID(ctx, testData.ProjectID, "ext-456")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domainContact.ID(), found.ID())
	assert.Equal(t, "Ricardo", found.Name())
	require.NotNil(t, found.ExternalID())
	assert.Equal(t, "ext-456", *found.ExternalID())
}

func TestGormContactRepository_FindByProject_Pagination(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create 5 contacts
	for i := 1; i <= 5; i++ {
		c, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Contact"+string(rune(i)))
		repo.Save(ctx, c)
	}

	// Act - Get first page (limit 2)
	page1, err := repo.FindByProject(ctx, testData.ProjectID, 2, 0)
	require.NoError(t, err)

	// Act - Get second page (limit 2, offset 2)
	page2, err := repo.FindByProject(ctx, testData.ProjectID, 2, 2)
	require.NoError(t, err)

	// Assert
	assert.Len(t, page1, 2, "First page should have 2 contacts")
	assert.Len(t, page2, 2, "Second page should have 2 contacts")

	// Verify different contacts
	assert.NotEqual(t, page1[0].ID(), page2[0].ID(), "Pages should return different contacts")
}

func TestGormContactRepository_FindByProject_ExcludesDeleted(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create 3 contacts, delete 1
	c1, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Active1")
	c2, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Active2")
	c3, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Deleted")

	repo.Save(ctx, c1)
	repo.Save(ctx, c2)
	repo.Save(ctx, c3)

	c3.SoftDelete()
	repo.Save(ctx, c3)

	// Act
	contacts, err := repo.FindByProject(ctx, testData.ProjectID, 10, 0)

	// Assert
	require.NoError(t, err)
	assert.Len(t, contacts, 2, "Should only return active contacts")
}

func TestGormContactRepository_CountByProject(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create 3 contacts
	for i := 1; i <= 3; i++ {
		c, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Contact"+string(rune(i)))
		repo.Save(ctx, c)
	}

	// Act
	count, err := repo.CountByProject(ctx, testData.ProjectID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestGormContactRepository_CountByProject_ExcludesDeleted(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create 2 contacts, delete 1
	c1, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Active")
	c2, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Deleted")
	repo.Save(ctx, c1)
	repo.Save(ctx, c2)

	c2.SoftDelete()
	repo.Save(ctx, c2)

	// Act
	count, err := repo.CountByProject(ctx, testData.ProjectID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should only count active contacts")
}

func TestGormContactRepository_Save_UpdatesTimestamps(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Timestamp Test")
	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	originalUpdatedAt := domainContact.UpdatedAt()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Act - Update contact
	domainContact.UpdateName("Updated Name")
	err = repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Reload from database
	found, err := repo.FindByID(ctx, domainContact.ID())
	require.NoError(t, err)

	// Assert
	assert.True(t, found.UpdatedAt().After(originalUpdatedAt), "UpdatedAt should be more recent after update")
}

func TestGormContactRepository_ProfilePicture(t *testing.T) {
	// Arrange
	testDB := SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)
	testData := testDB.SeedTestData(t)

	repo := NewGormContactRepository(testDB.DB)
	ctx := context.Background()

	// Create contact
	domainContact, _ := contact.NewContact(testData.ProjectID, testData.TenantID, "Profile Test")
	domainContact.SetProfilePicture("https://example.com/profile.jpg")

	err := repo.Save(ctx, domainContact)
	require.NoError(t, err)

	// Act
	found, err := repo.FindByID(ctx, domainContact.ID())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, found.ProfilePictureURL())
	assert.Equal(t, "https://example.com/profile.jpg", *found.ProfilePictureURL())
	assert.NotNil(t, found.ProfilePictureFetchedAt())
}
