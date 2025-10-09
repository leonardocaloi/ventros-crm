package contact

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Contact é o Aggregate Root para contatos.
type Contact struct {
	id            uuid.UUID
	projectID     uuid.UUID
	tenantID      string
	name          string
	email         *Email
	phone         *Phone
	externalID    *string
	sourceChannel *string
	language      string
	timezone      *string
	tags          []string

	// WhatsApp Profile
	profilePictureURL       *string
	profilePictureFetchedAt *time.Time

	firstInteractionAt *time.Time
	lastInteractionAt  *time.Time
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          *time.Time

	events []DomainEvent
}

// NewContact cria um novo contato.
func NewContact(
	projectID uuid.UUID,
	tenantID string,
	name string,
) (*Contact, error) {
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	contact := &Contact{
		id:        uuid.New(),
		projectID: projectID,
		tenantID:  tenantID,
		name:      name,
		language:  "en",
		tags:      []string{},
		createdAt: now,
		updatedAt: now,
		events:    []DomainEvent{},
	}

	contact.addEvent(NewContactCreatedEvent(contact.id, projectID, tenantID, name))

	return contact, nil
}

// ReconstructContact reconstrói um Contact a partir de dados persistidos.
func ReconstructContact(
	id uuid.UUID,
	projectID uuid.UUID,
	tenantID string,
	name string,
	email *Email,
	phone *Phone,
	externalID *string,
	sourceChannel *string,
	language string,
	timezone *string,
	tags []string,
	profilePictureURL *string,
	profilePictureFetchedAt *time.Time,
	firstInteractionAt *time.Time,
	lastInteractionAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Contact {
	if tags == nil {
		tags = []string{}
	}

	return &Contact{
		id:                      id,
		projectID:               projectID,
		tenantID:                tenantID,
		name:                    name,
		email:                   email,
		phone:                   phone,
		externalID:              externalID,
		sourceChannel:           sourceChannel,
		language:                language,
		timezone:                timezone,
		tags:                    tags,
		profilePictureURL:       profilePictureURL,
		profilePictureFetchedAt: profilePictureFetchedAt,
		firstInteractionAt:      firstInteractionAt,
		lastInteractionAt:       lastInteractionAt,
		createdAt:               createdAt,
		updatedAt:               updatedAt,
		deletedAt:               deletedAt,
		events:                  []DomainEvent{},
	}
}

// SetEmail define o email do contato.
func (c *Contact) SetEmail(emailStr string) error {
	email, err := NewEmail(emailStr)
	if err != nil {
		return err
	}
	c.email = &email
	c.updatedAt = time.Now()
	return nil
}

// SetPhone define o telefone do contato.
func (c *Contact) SetPhone(phoneStr string) error {
	phone, err := NewPhone(phoneStr)
	if err != nil {
		return err
	}
	c.phone = &phone
	c.updatedAt = time.Now()
	return nil
}

// UpdateName atualiza o nome do contato.
func (c *Contact) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	c.name = name
	c.updatedAt = time.Now()

	c.addEvent(NewContactUpdatedEvent(c.id))

	return nil
}

// AddTag adiciona uma tag ao contato.
func (c *Contact) AddTag(tag string) {
	// Evita duplicatas
	for _, t := range c.tags {
		if t == tag {
			return
		}
	}
	c.tags = append(c.tags, tag)
	c.updatedAt = time.Now()
}

// RemoveTag remove uma tag do contato.
func (c *Contact) RemoveTag(tag string) {
	for i, t := range c.tags {
		if t == tag {
			c.tags = append(c.tags[:i], c.tags[i+1:]...)
			c.updatedAt = time.Now()
			return
		}
	}
}

// ClearTags remove todas as tags do contato.
func (c *Contact) ClearTags() {
	c.tags = []string{}
	c.updatedAt = time.Now()
}

// SetExternalID define o ID externo do contato.
func (c *Contact) SetExternalID(externalID string) {
	if externalID == "" {
		c.externalID = nil
	} else {
		c.externalID = &externalID
	}
	c.updatedAt = time.Now()
}

// SetSourceChannel define o canal de origem do contato.
func (c *Contact) SetSourceChannel(sourceChannel string) {
	if sourceChannel == "" {
		c.sourceChannel = nil
	} else {
		c.sourceChannel = &sourceChannel
	}
	c.updatedAt = time.Now()
}

// SetLanguage define o idioma do contato.
func (c *Contact) SetLanguage(language string) {
	if language == "" {
		c.language = "en" // default
	} else {
		c.language = language
	}
	c.updatedAt = time.Now()
}

// SetTimezone define o fuso horário do contato.
func (c *Contact) SetTimezone(timezone string) {
	if timezone == "" {
		c.timezone = nil
	} else {
		c.timezone = &timezone
	}
	c.updatedAt = time.Now()
}

// SetProfilePicture define a URL da foto de perfil do WhatsApp
func (c *Contact) SetProfilePicture(url string) {
	if url == "" {
		c.profilePictureURL = nil
		c.profilePictureFetchedAt = nil
	} else {
		c.profilePictureURL = &url
		now := time.Now()
		c.profilePictureFetchedAt = &now
	}
	c.updatedAt = time.Now()
}

// Delete é um alias para SoftDelete para compatibilidade.
func (c *Contact) Delete() error {
	return c.SoftDelete()
}

// RecordInteraction registra uma interação.
func (c *Contact) RecordInteraction() {
	now := time.Now()

	if c.firstInteractionAt == nil {
		c.firstInteractionAt = &now
	}
	c.lastInteractionAt = &now
	c.updatedAt = now
}

// SoftDelete marca o contato como deletado.
func (c *Contact) SoftDelete() error {
	if c.deletedAt != nil {
		return errors.New("contact already deleted")
	}

	now := time.Now()
	c.deletedAt = &now
	c.updatedAt = now

	c.addEvent(NewContactDeletedEvent(c.id))

	return nil
}

// IsDeleted verifica se o contato foi deletado.
func (c *Contact) IsDeleted() bool {
	return c.deletedAt != nil
}

// Getters

func (c *Contact) ID() uuid.UUID                       { return c.id }
func (c *Contact) ProjectID() uuid.UUID                { return c.projectID }
func (c *Contact) TenantID() string                    { return c.tenantID }
func (c *Contact) Name() string                        { return c.name }
func (c *Contact) Email() *Email                       { return c.email }
func (c *Contact) Phone() *Phone                       { return c.phone }
func (c *Contact) ExternalID() *string                 { return c.externalID }
func (c *Contact) SourceChannel() *string              { return c.sourceChannel }
func (c *Contact) Language() string                    { return c.language }
func (c *Contact) Timezone() *string                   { return c.timezone }
func (c *Contact) Tags() []string                      { return append([]string{}, c.tags...) } // Copy
func (c *Contact) ProfilePictureURL() *string          { return c.profilePictureURL }
func (c *Contact) ProfilePictureFetchedAt() *time.Time { return c.profilePictureFetchedAt }
func (c *Contact) FirstInteractionAt() *time.Time      { return c.firstInteractionAt }
func (c *Contact) LastInteractionAt() *time.Time       { return c.lastInteractionAt }
func (c *Contact) CreatedAt() time.Time                { return c.createdAt }
func (c *Contact) UpdatedAt() time.Time                { return c.updatedAt }
func (c *Contact) DeletedAt() *time.Time               { return c.deletedAt }

// DomainEvents retorna os eventos de domínio.
func (c *Contact) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, c.events...)
}

// ClearEvents limpa os eventos.
func (c *Contact) ClearEvents() {
	c.events = []DomainEvent{}
}

func (c *Contact) addEvent(event DomainEvent) {
	c.events = append(c.events, event)
}
