package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestHelpers provides common utilities for domain tests

// NewTestUUID generates a new UUID for testing
func NewTestUUID() uuid.UUID {
	return uuid.New()
}

// NewTestTenantID generates a test tenant ID
func NewTestTenantID() string {
	return "test-tenant-" + uuid.New().String()[:8]
}

// AssertTimeAlmostEqual checks if two times are within a small delta (useful for timestamp assertions)
func AssertTimeAlmostEqual(t *testing.T, expected, actual time.Time, delta time.Duration) {
	t.Helper()
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	require.True(t, diff <= delta, "Times are not almost equal: expected %v, got %v (delta: %v)", expected, actual, diff)
}

// AssertTimeNotZero checks that a time is not zero
func AssertTimeNotZero(t *testing.T, tm time.Time, fieldName string) {
	t.Helper()
	require.False(t, tm.IsZero(), "%s should not be zero", fieldName)
}

// AssertTimeIsZero checks that a time is zero
func AssertTimeIsZero(t *testing.T, tm time.Time, fieldName string) {
	t.Helper()
	require.True(t, tm.IsZero(), "%s should be zero", fieldName)
}

// AssertUUIDNotNil checks that a UUID pointer is not nil and not zero
func AssertUUIDNotNil(t *testing.T, id *uuid.UUID, fieldName string) {
	t.Helper()
	require.NotNil(t, id, "%s should not be nil", fieldName)
	require.NotEqual(t, uuid.Nil, *id, "%s should not be zero UUID", fieldName)
}

// AssertUUIDEqual checks that two UUID pointers are equal
func AssertUUIDEqual(t *testing.T, expected, actual *uuid.UUID, fieldName string) {
	t.Helper()
	require.NotNil(t, expected, "expected %s should not be nil", fieldName)
	require.NotNil(t, actual, "actual %s should not be nil", fieldName)
	require.Equal(t, *expected, *actual, "%s should match", fieldName)
}

// NowPtr returns a pointer to the current time
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// UUIDPtr converts a UUID to a pointer
func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

// StringPtr converts a string to a pointer
func StringPtr(s string) *string {
	return &s
}

// IntPtr converts an int to a pointer
func IntPtr(i int) *int {
	return &i
}
