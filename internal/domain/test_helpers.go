package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestUUID() uuid.UUID {
	return uuid.New()
}

func NewTestTenantID() string {
	return "test-tenant-" + uuid.New().String()[:8]
}

func AssertTimeAlmostEqual(t *testing.T, expected, actual time.Time, delta time.Duration) {
	t.Helper()
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	require.True(t, diff <= delta, "Times are not almost equal: expected %v, got %v (delta: %v)", expected, actual, diff)
}

func AssertTimeNotZero(t *testing.T, tm time.Time, fieldName string) {
	t.Helper()
	require.False(t, tm.IsZero(), "%s should not be zero", fieldName)
}

func AssertTimeIsZero(t *testing.T, tm time.Time, fieldName string) {
	t.Helper()
	require.True(t, tm.IsZero(), "%s should be zero", fieldName)
}

func AssertUUIDNotNil(t *testing.T, id *uuid.UUID, fieldName string) {
	t.Helper()
	require.NotNil(t, id, "%s should not be nil", fieldName)
	require.NotEqual(t, uuid.Nil, *id, "%s should not be zero UUID", fieldName)
}

func AssertUUIDEqual(t *testing.T, expected, actual *uuid.UUID, fieldName string) {
	t.Helper()
	require.NotNil(t, expected, "expected %s should not be nil", fieldName)
	require.NotNil(t, actual, "actual %s should not be nil", fieldName)
	require.Equal(t, *expected, *actual, "%s should match", fieldName)
}

func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}
