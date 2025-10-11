package campaign

import "github.com/google/uuid"

// Repository defines the interface for campaign persistence
type Repository interface {
	Save(campaign *Campaign) error
	FindByID(id uuid.UUID) (*Campaign, error)
	FindByTenantID(tenantID string) ([]*Campaign, error)
	FindActiveByStatus(status CampaignStatus) ([]*Campaign, error)
	FindScheduled() ([]*Campaign, error) // Find campaigns scheduled to start
	Delete(id uuid.UUID) error
}

// EnrollmentRepository defines the interface for campaign enrollment persistence
type EnrollmentRepository interface {
	Save(enrollment *CampaignEnrollment) error
	FindByID(id uuid.UUID) (*CampaignEnrollment, error)
	FindByCampaignID(campaignID uuid.UUID) ([]*CampaignEnrollment, error)
	FindByContactID(contactID uuid.UUID) ([]*CampaignEnrollment, error)
	FindReadyForNextStep() ([]*CampaignEnrollment, error)
	FindActiveByCampaignAndContact(campaignID, contactID uuid.UUID) (*CampaignEnrollment, error)
	Delete(id uuid.UUID) error
}
