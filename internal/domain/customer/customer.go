package customer

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Customer é o Aggregate Root para organizações/empresas.
type Customer struct {
	id        uuid.UUID
	name      string
	email     string
	status    Status
	settings  map[string]interface{}
	createdAt time.Time
	updatedAt time.Time

	events []DomainEvent
}

// NewCustomer cria um novo cliente.
func NewCustomer(name, email string) (*Customer, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	now := time.Now()
	customer := &Customer{
		id:        uuid.New(),
		name:      name,
		email:     email,
		status:    StatusActive,
		settings:  make(map[string]interface{}),
		createdAt: now,
		updatedAt: now,
		events:    []DomainEvent{},
	}

	customer.addEvent(CustomerCreatedEvent{
		CustomerID: customer.id,
		Name:       name,
		Email:      email,
		CreatedAt:  now,
	})

	return customer, nil
}

// ReconstructCustomer reconstrói um Customer a partir de dados persistidos.
func ReconstructCustomer(
	id uuid.UUID,
	name string,
	email string,
	status Status,
	settings map[string]interface{},
	createdAt time.Time,
	updatedAt time.Time,
) *Customer {
	if settings == nil {
		settings = make(map[string]interface{})
	}

	return &Customer{
		id:        id,
		name:      name,
		email:     email,
		status:    status,
		settings:  settings,
		createdAt: createdAt,
		updatedAt: updatedAt,
		events:    []DomainEvent{},
	}
}

// Activate ativa o cliente.
func (c *Customer) Activate() error {
	if c.status == StatusActive {
		return nil
	}

	c.status = StatusActive
	c.updatedAt = time.Now()

	c.addEvent(CustomerActivatedEvent{
		CustomerID:  c.id,
		ActivatedAt: time.Now(),
	})

	return nil
}

// Suspend suspende o cliente.
func (c *Customer) Suspend() error {
	if c.status == StatusSuspended {
		return nil
	}

	c.status = StatusSuspended
	c.updatedAt = time.Now()

	c.addEvent(CustomerSuspendedEvent{
		CustomerID:  c.id,
		SuspendedAt: time.Now(),
	})

	return nil
}

// UpdateSettings atualiza as configurações do cliente.
func (c *Customer) UpdateSettings(settings map[string]interface{}) {
	c.settings = settings
	c.updatedAt = time.Now()
}

// GetSetting retorna uma configuração específica.
func (c *Customer) GetSetting(key string) (interface{}, bool) {
	val, ok := c.settings[key]
	return val, ok
}

// IsActive verifica se o cliente está ativo.
func (c *Customer) IsActive() bool {
	return c.status == StatusActive
}

// Getters
func (c *Customer) ID() uuid.UUID  { return c.id }
func (c *Customer) Name() string   { return c.name }
func (c *Customer) Email() string  { return c.email }
func (c *Customer) Status() Status { return c.status }
func (c *Customer) Settings() map[string]interface{} {
	// Return copy
	copy := make(map[string]interface{})
	for k, v := range c.settings {
		copy[k] = v
	}
	return copy
}
func (c *Customer) CreatedAt() time.Time { return c.createdAt }
func (c *Customer) UpdatedAt() time.Time { return c.updatedAt }

// Domain Events
func (c *Customer) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, c.events...)
}

func (c *Customer) ClearEvents() {
	c.events = []DomainEvent{}
}

func (c *Customer) addEvent(event DomainEvent) {
	c.events = append(c.events, event)
}
