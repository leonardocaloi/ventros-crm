package shared

import "github.com/google/uuid"

// AggregateRoot is a marker interface that all aggregate roots must implement.
// Aggregates are the unit of consistency in DDD - all invariants must be
// maintained within a single aggregate boundary.
//
// An Aggregate Root is responsible for:
// 1. Maintaining invariants across all entities within the aggregate
// 2. Coordinating state changes
// 3. Emitting domain events for significant state changes
// 4. Ensuring transactional boundaries
//
// Key Principles:
// - Aggregates should be small (2-3 entities max)
// - All changes to the aggregate must go through the root
// - References between aggregates should be by ID only
// - Optimistic locking (version) prevents lost updates
//
// Based on: Eric Evans' DDD (2003), Vaughn Vernon's IDDD (2013)
type AggregateRoot interface {
	// ID returns the unique identifier of the aggregate root
	ID() uuid.UUID

	// Version returns the current version for optimistic locking
	// Version is incremented on every successful update
	Version() int

	// DomainEvents returns all uncommitted domain events
	// Events are emitted when significant state changes occur
	DomainEvents() []DomainEvent

	// ClearEvents clears all uncommitted domain events
	// Should be called after events are persisted/published
	ClearEvents()
}

// Compile-time checks that our aggregates implement AggregateRoot
// Usage in aggregate files:
// var _ shared.AggregateRoot = (*Contact)(nil)
