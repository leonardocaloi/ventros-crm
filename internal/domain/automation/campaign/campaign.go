package campaign

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Campaign represents a complex multi-step marketing campaign
// that can orchestrate broadcasts, sequences, and conditional logic
type Campaign struct {
	id          uuid.UUID
	tenantID    string
	name        string
	description string
	status      CampaignStatus
	steps       []CampaignStep

	// Goals and metrics
	goalType         GoalType
	goalValue        int
	contactsReached  int
	conversionsCount int

	// Scheduling
	startDate *time.Time
	endDate   *time.Time

	// Timestamps
	createdAt time.Time
	updatedAt time.Time

	// Domain events
	events []interface{}
}

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusArchived  CampaignStatus = "archived"
)

type GoalType string

const (
	GoalTypeReachContacts GoalType = "reach_contacts"
	GoalTypeConversions   GoalType = "conversions"
	GoalTypeEngagement    GoalType = "engagement"
)

// NewCampaign creates a new campaign in draft status
func NewCampaign(tenantID, name, description string, goalType GoalType, goalValue int) (*Campaign, error) {
	if tenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if goalValue < 0 {
		return nil, errors.New("goal_value must be non-negative")
	}

	now := time.Now()
	campaign := &Campaign{
		id:               uuid.New(),
		tenantID:         tenantID,
		name:             name,
		description:      description,
		status:           CampaignStatusDraft,
		steps:            []CampaignStep{},
		goalType:         goalType,
		goalValue:        goalValue,
		contactsReached:  0,
		conversionsCount: 0,
		createdAt:        now,
		updatedAt:        now,
		events:           []interface{}{},
	}

	// Emit domain event
	campaign.addEvent(CampaignCreatedEvent{
		CampaignID:  campaign.id,
		TenantID:    campaign.tenantID,
		Name:        campaign.name,
		Description: campaign.description,
		GoalType:    campaign.goalType,
		GoalValue:   campaign.goalValue,
		OccurredAt:  now,
	})

	return campaign, nil
}

// ReconstructCampaign reconstructs a campaign from persistence
func ReconstructCampaign(
	id uuid.UUID,
	tenantID string,
	name string,
	description string,
	status CampaignStatus,
	steps []CampaignStep,
	goalType GoalType,
	goalValue int,
	contactsReached int,
	conversionsCount int,
	startDate *time.Time,
	endDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Campaign {
	return &Campaign{
		id:               id,
		tenantID:         tenantID,
		name:             name,
		description:      description,
		status:           status,
		steps:            steps,
		goalType:         goalType,
		goalValue:        goalValue,
		contactsReached:  contactsReached,
		conversionsCount: conversionsCount,
		startDate:        startDate,
		endDate:          endDate,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		events:           []interface{}{},
	}
}

// State transition methods

// Activate activates a draft or scheduled campaign
func (c *Campaign) Activate() error {
	if c.status != CampaignStatusDraft && c.status != CampaignStatusScheduled {
		return errors.New("can only activate draft or scheduled campaigns")
	}
	if len(c.steps) == 0 {
		return errors.New("cannot activate campaign with no steps")
	}

	c.status = CampaignStatusActive
	c.updatedAt = time.Now()

	c.addEvent(CampaignActivatedEvent{
		CampaignID: c.id,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// Schedule schedules a campaign to start at a specific time
func (c *Campaign) Schedule(startDate time.Time) error {
	if c.status != CampaignStatusDraft {
		return errors.New("can only schedule draft campaigns")
	}
	if len(c.steps) == 0 {
		return errors.New("cannot schedule campaign with no steps")
	}
	if startDate.Before(time.Now()) {
		return errors.New("start date must be in the future")
	}

	c.status = CampaignStatusScheduled
	c.startDate = &startDate
	c.updatedAt = time.Now()

	c.addEvent(CampaignScheduledEvent{
		CampaignID: c.id,
		StartDate:  startDate,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// Pause pauses an active campaign
func (c *Campaign) Pause() error {
	if c.status != CampaignStatusActive {
		return errors.New("can only pause active campaigns")
	}

	c.status = CampaignStatusPaused
	c.updatedAt = time.Now()

	c.addEvent(CampaignPausedEvent{
		CampaignID: c.id,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// Resume resumes a paused campaign
func (c *Campaign) Resume() error {
	if c.status != CampaignStatusPaused {
		return errors.New("can only resume paused campaigns")
	}

	c.status = CampaignStatusActive
	c.updatedAt = time.Now()

	c.addEvent(CampaignResumedEvent{
		CampaignID: c.id,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// Complete marks a campaign as completed
func (c *Campaign) Complete() error {
	if c.status != CampaignStatusActive {
		return errors.New("can only complete active campaigns")
	}

	c.status = CampaignStatusCompleted
	now := time.Now()
	c.endDate = &now
	c.updatedAt = now

	c.addEvent(CampaignCompletedEvent{
		CampaignID: c.id,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// Archive archives a campaign
func (c *Campaign) Archive() error {
	if c.status == CampaignStatusArchived {
		return errors.New("campaign is already archived")
	}

	c.status = CampaignStatusArchived
	c.updatedAt = time.Now()

	c.addEvent(CampaignArchivedEvent{
		CampaignID: c.id,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// Step management

// AddStep adds a step to the campaign
func (c *Campaign) AddStep(step CampaignStep) error {
	if c.status != CampaignStatusDraft {
		return errors.New("can only add steps to draft campaigns")
	}

	// Validate step order uniqueness
	for _, existingStep := range c.steps {
		if existingStep.Order == step.Order {
			return errors.New("step with this order already exists")
		}
	}

	c.steps = append(c.steps, step)
	c.updatedAt = time.Now()

	c.addEvent(CampaignStepAddedEvent{
		CampaignID: c.id,
		StepID:     step.ID,
		StepType:   step.Type,
		Order:      step.Order,
		OccurredAt: c.updatedAt,
	})

	return nil
}

// RemoveStep removes a step from the campaign
func (c *Campaign) RemoveStep(stepID uuid.UUID) error {
	if c.status != CampaignStatusDraft {
		return errors.New("can only remove steps from draft campaigns")
	}

	for i, step := range c.steps {
		if step.ID == stepID {
			c.steps = append(c.steps[:i], c.steps[i+1:]...)
			c.updatedAt = time.Now()

			c.addEvent(CampaignStepRemovedEvent{
				CampaignID: c.id,
				StepID:     stepID,
				OccurredAt: c.updatedAt,
			})

			return nil
		}
	}

	return errors.New("step not found")
}

// UpdateStep updates a step in the campaign
func (c *Campaign) UpdateStep(stepID uuid.UUID, updatedStep CampaignStep) error {
	if c.status != CampaignStatusDraft {
		return errors.New("can only update steps in draft campaigns")
	}

	for i, step := range c.steps {
		if step.ID == stepID {
			// Validate order uniqueness if changed
			if step.Order != updatedStep.Order {
				for j, s := range c.steps {
					if i != j && s.Order == updatedStep.Order {
						return errors.New("step with this order already exists")
					}
				}
			}

			c.steps[i] = updatedStep
			c.updatedAt = time.Now()
			return nil
		}
	}

	return errors.New("step not found")
}

// GetStepByOrder retrieves a step by its order
func (c *Campaign) GetStepByOrder(order int) (*CampaignStep, error) {
	for _, step := range c.steps {
		if step.Order == order {
			return &step, nil
		}
	}
	return nil, errors.New("step not found")
}

// Statistics methods

// IncrementContactsReached increments the contacts reached counter
func (c *Campaign) IncrementContactsReached() {
	c.contactsReached++
	c.updatedAt = time.Now()
}

// IncrementConversions increments the conversions counter
func (c *Campaign) IncrementConversions() {
	c.conversionsCount++
	c.updatedAt = time.Now()
}

// GetStats returns campaign statistics
func (c *Campaign) GetStats() CampaignStats {
	var conversionRate float64
	if c.contactsReached > 0 {
		conversionRate = float64(c.conversionsCount) / float64(c.contactsReached) * 100
	}

	var progressRate float64
	if c.goalValue > 0 {
		switch c.goalType {
		case GoalTypeReachContacts:
			progressRate = float64(c.contactsReached) / float64(c.goalValue) * 100
		case GoalTypeConversions:
			progressRate = float64(c.conversionsCount) / float64(c.goalValue) * 100
		}
	}

	return CampaignStats{
		ContactsReached:  c.contactsReached,
		ConversionsCount: c.conversionsCount,
		ConversionRate:   conversionRate,
		ProgressRate:     progressRate,
	}
}

// Updaters

// UpdateName updates the campaign name
func (c *Campaign) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	c.name = name
	c.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the campaign description
func (c *Campaign) UpdateDescription(description string) {
	c.description = description
	c.updatedAt = time.Now()
}

// UpdateGoal updates the campaign goal
func (c *Campaign) UpdateGoal(goalType GoalType, goalValue int) error {
	if goalValue < 0 {
		return errors.New("goal_value must be non-negative")
	}
	c.goalType = goalType
	c.goalValue = goalValue
	c.updatedAt = time.Now()
	return nil
}

// Getters

func (c *Campaign) ID() uuid.UUID              { return c.id }
func (c *Campaign) TenantID() string           { return c.tenantID }
func (c *Campaign) Name() string               { return c.name }
func (c *Campaign) Description() string        { return c.description }
func (c *Campaign) Status() CampaignStatus     { return c.status }
func (c *Campaign) Steps() []CampaignStep      { return c.steps }
func (c *Campaign) GoalType() GoalType         { return c.goalType }
func (c *Campaign) GoalValue() int             { return c.goalValue }
func (c *Campaign) ContactsReached() int       { return c.contactsReached }
func (c *Campaign) ConversionsCount() int      { return c.conversionsCount }
func (c *Campaign) StartDate() *time.Time      { return c.startDate }
func (c *Campaign) EndDate() *time.Time        { return c.endDate }
func (c *Campaign) CreatedAt() time.Time       { return c.createdAt }
func (c *Campaign) UpdatedAt() time.Time       { return c.updatedAt }
func (c *Campaign) DomainEvents() []interface{} { return c.events }

func (c *Campaign) ClearEvents() {
	c.events = []interface{}{}
}

func (c *Campaign) addEvent(event interface{}) {
	c.events = append(c.events, event)
}

// CampaignStats represents campaign statistics
type CampaignStats struct {
	ContactsReached  int     `json:"contacts_reached"`
	ConversionsCount int     `json:"conversions_count"`
	ConversionRate   float64 `json:"conversion_rate"`
	ProgressRate     float64 `json:"progress_rate"`
}
