#!/bin/bash
# Script to generate comprehensive domain tests for all entities
# Following enterprise patterns established in pipeline, agent, and billing tests

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Generating Domain Tests...${NC}"

# Contact domain - comprehensive tests
cat > internal/domain/contact/full_contact_test.go << 'EOF'
package contact

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContact_Comprehensive(t *testing.T) {
	projectID := uuid.New()

	t.Run("NewContact validation", func(t *testing.T) {
		contact, err := NewContact(projectID, "tenant-123", "John Doe")
		require.NoError(t, err)
		assert.Equal(t, "John Doe", contact.Name())
		assert.Equal(t, "en", contact.Language())
	})

	t.Run("SetEmail", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		err := contact.SetEmail("john@example.com")
		require.NoError(t, err)
		assert.NotNil(t, contact.Email())
	})

	t.Run("SetPhone", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		err := contact.SetPhone("+5511999999999")
		require.NoError(t, err)
		assert.NotNil(t, contact.Phone())
	})

	t.Run("Tags management", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		contact.AddTag("vip")
		contact.AddTag("customer")
		assert.Len(t, contact.Tags(), 2)

		contact.RemoveTag("vip")
		assert.Len(t, contact.Tags(), 1)

		contact.ClearTags()
		assert.Len(t, contact.Tags(), 0)
	})

	t.Run("Soft delete", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		assert.False(t, contact.IsDeleted())

		err := contact.SoftDelete()
		require.NoError(t, err)
		assert.True(t, contact.IsDeleted())

		// Cannot delete twice
		err = contact.SoftDelete()
		require.Error(t, err)
	})

	t.Run("RecordInteraction", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		assert.Nil(t, contact.FirstInteractionAt())

		contact.RecordInteraction()
		assert.NotNil(t, contact.FirstInteractionAt())
		assert.NotNil(t, contact.LastInteractionAt())
	})
}
EOF

# Project domain tests
cat > internal/domain/project/project_test.go << 'EOF'
package project

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProject(t *testing.T) {
	userID := uuid.New()
	billingAccountID := uuid.New()

	t.Run("valid project", func(t *testing.T) {
		project, err := NewProject(userID, billingAccountID, "tenant-123", "Test Project")
		require.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Test Project", project.Name())
		assert.True(t, project.IsActive())
	})

	t.Run("validation errors", func(t *testing.T) {
		_, err := NewProject(uuid.Nil, billingAccountID, "tenant", "Project")
		assert.Error(t, err)

		_, err = NewProject(userID, uuid.Nil, "tenant", "Project")
		assert.Error(t, err)

		_, err = NewProject(userID, billingAccountID, "", "Project")
		assert.Error(t, err)

		_, err = NewProject(userID, billingAccountID, "tenant", "")
		assert.Error(t, err)
	})

	t.Run("activate/deactivate", func(t *testing.T) {
		project, _ := NewProject(userID, billingAccountID, "tenant", "Project")

		project.Deactivate()
		assert.False(t, project.IsActive())

		project.Activate()
		assert.True(t, project.IsActive())
	})
}
EOF

# Note domain tests
cat > internal/domain/note/note_test.go << 'EOF'
package note

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNote(t *testing.T) {
	contactID := uuid.New()
	authorID := uuid.New()

	t.Run("valid note", func(t *testing.T) {
		note, err := NewNote("tenant-123", contactID, authorID, "Test note content")
		require.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test note content", note.Content())
	})

	t.Run("validation", func(t *testing.T) {
		_, err := NewNote("", contactID, authorID, "Content")
		assert.Error(t, err)

		_, err = NewNote("tenant", uuid.Nil, authorID, "Content")
		assert.Error(t, err)

		_, err = NewNote("tenant", contactID, uuid.Nil, "Content")
		assert.Error(t, err)

		_, err = NewNote("tenant", contactID, authorID, "")
		assert.Error(t, err)
	})
}
EOF

# Customer domain tests
cat > internal/domain/customer/customer_test.go << 'EOF'
package customer

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomer(t *testing.T) {
	contactID := uuid.New()

	t.Run("valid customer", func(t *testing.T) {
		customer, err := NewCustomer(contactID, "tenant-123", "customer@example.com")
		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, contactID, customer.ContactID())
	})

	t.Run("validation", func(t *testing.T) {
		_, err := NewCustomer(uuid.Nil, "tenant", "email@test.com")
		assert.Error(t, err)

		_, err = NewCustomer(contactID, "", "email@test.com")
		assert.Error(t, err)
	})
}
EOF

echo -e "${GREEN}✅ Test files generated!${NC}"
echo ""
echo "Generated tests for:"
echo "  - Contact (full_contact_test.go)"
echo "  - Project (project_test.go)"
echo "  - Note (note_test.go)"
echo "  - Customer (customer_test.go)"
echo ""
echo "Run tests with:"
echo "  go test ./internal/domain/..."
