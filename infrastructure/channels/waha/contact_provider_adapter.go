package waha

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/crm/channel"
	"go.uber.org/zap"
)

// ContactProviderAdapter implements domain.ContactProvider for WAHA
//
// This adapter is used by both channel types:
// - TypeWAHA (manual mode)
// - TypeWhatsAppBusiness (auto mode)
//
// Both types share the same WAHA implementation underneath.
type ContactProviderAdapter struct {
	client      *Client
	sessionName string
	logger      *zap.Logger
}

// NewContactProviderAdapter creates a new contact provider adapter
func NewContactProviderAdapter(client *Client, sessionName string, logger *zap.Logger) *ContactProviderAdapter {
	return &ContactProviderAdapter{
		client:      client,
		sessionName: sessionName,
		logger:      logger,
	}
}

// GetAllContacts returns all contacts from the channel
//
// Implementation of domain.ContactProvider interface
func (a *ContactProviderAdapter) GetAllContacts(ctx context.Context, sortBy, sortOrder string, limit, offset int) ([]channel.Contact, error) {
	a.logger.Info("Getting all contacts",
		zap.String("session_name", a.sessionName),
		zap.String("sort_by", sortBy),
		zap.String("sort_order", sortOrder),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Build query parameters
	path := fmt.Sprintf("/api/contacts/all?session=%s", a.sessionName)

	if sortBy != "" {
		path += fmt.Sprintf("&sortBy=%s", sortBy)
	}
	if sortOrder != "" {
		path += fmt.Sprintf("&sortOrder=%s", sortOrder)
	}
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}
	if offset > 0 {
		path += fmt.Sprintf("&offset=%d", offset)
	}

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get all contacts",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get all contacts: %w", err)
	}

	var wahaContacts []WAHAContact
	if err := a.client.ParseResponse(resp, &wahaContacts); err != nil {
		return nil, fmt.Errorf("failed to parse contacts response: %w", err)
	}

	// Convert WAHA contacts to domain contacts
	contacts := make([]channel.Contact, len(wahaContacts))
	for i, wc := range wahaContacts {
		contacts[i] = mapWAHAContactToDomain(wc)
	}

	a.logger.Info("Successfully retrieved contacts",
		zap.String("session_name", a.sessionName),
		zap.Int("count", len(contacts)))

	return contacts, nil
}

// GetContact gets basic contact information
//
// Always returns result even if phone is not registered in WhatsApp.
// Use CheckExists to verify if number is registered.
func (a *ContactProviderAdapter) GetContact(ctx context.Context, contactID string) (*channel.Contact, error) {
	a.logger.Info("Getting contact",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	path := fmt.Sprintf("/api/contacts?session=%s&contactId=%s", a.sessionName, contactID)

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get contact",
			zap.String("session_name", a.sessionName),
			zap.String("contact_id", contactID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	var wahaContact WAHAContact
	if err := a.client.ParseResponse(resp, &wahaContact); err != nil {
		return nil, fmt.Errorf("failed to parse contact response: %w", err)
	}

	contact := mapWAHAContactToDomain(wahaContact)

	a.logger.Info("Successfully retrieved contact",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	return &contact, nil
}

// CheckExists checks if phone number is registered in WhatsApp
func (a *ContactProviderAdapter) CheckExists(ctx context.Context, phoneNumber string) (*channel.ContactExistence, error) {
	a.logger.Info("Checking contact existence",
		zap.String("session_name", a.sessionName),
		zap.String("phone_number", phoneNumber))

	path := fmt.Sprintf("/api/contacts/check-exists?session=%s&phone=%s", a.sessionName, phoneNumber)

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to check contact existence",
			zap.String("session_name", a.sessionName),
			zap.String("phone_number", phoneNumber),
			zap.Error(err))
		return nil, fmt.Errorf("failed to check contact existence: %w", err)
	}

	var wahaExistence WAHAContactExistence
	if err := a.client.ParseResponse(resp, &wahaExistence); err != nil {
		return nil, fmt.Errorf("failed to parse existence response: %w", err)
	}

	existence := &channel.ContactExistence{
		PhoneNumber:  wahaExistence.NumberExists.PhoneNumber,
		NumberExists: wahaExistence.NumberExists.NumberExists,
		ChatID:       wahaExistence.NumberExists.ChatID,
	}

	a.logger.Info("Contact existence checked",
		zap.String("session_name", a.sessionName),
		zap.String("phone_number", phoneNumber),
		zap.Bool("exists", existence.NumberExists))

	return existence, nil
}

// GetAbout gets contact's "about" status text
//
// Returns nil if you don't have permission to read their status.
func (a *ContactProviderAdapter) GetAbout(ctx context.Context, contactID string) (*string, error) {
	a.logger.Info("Getting contact about",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	path := fmt.Sprintf("/api/contacts/about?session=%s&contactId=%s", a.sessionName, contactID)

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get contact about",
			zap.String("session_name", a.sessionName),
			zap.String("contact_id", contactID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get contact about: %w", err)
	}

	var wahaAbout WAHAContactAbout
	if err := a.client.ParseResponse(resp, &wahaAbout); err != nil {
		return nil, fmt.Errorf("failed to parse about response: %w", err)
	}

	a.logger.Info("Successfully retrieved contact about",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	return wahaAbout.About, nil
}

// GetProfilePicture gets contact's profile picture URL
//
// If privacy settings don't allow, returns nil.
// Set refresh=true to bypass 24h cache (use carefully - rate limits!)
func (a *ContactProviderAdapter) GetProfilePicture(ctx context.Context, contactID string, refresh bool) (*string, error) {
	a.logger.Info("Getting contact profile picture",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID),
		zap.Bool("refresh", refresh))

	path := fmt.Sprintf("/api/contacts/profile-picture?session=%s&contactId=%s", a.sessionName, contactID)
	if refresh {
		path += "&refresh=true"
	}

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get contact profile picture",
			zap.String("session_name", a.sessionName),
			zap.String("contact_id", contactID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get contact profile picture: %w", err)
	}

	var wahaPicture WAHAContactProfilePicture
	if err := a.client.ParseResponse(resp, &wahaPicture); err != nil {
		return nil, fmt.Errorf("failed to parse profile picture response: %w", err)
	}

	a.logger.Info("Successfully retrieved contact profile picture",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	return wahaPicture.ProfilePictureURL, nil
}

// BlockContact blocks a contact
func (a *ContactProviderAdapter) BlockContact(ctx context.Context, contactID string) error {
	a.logger.Info("Blocking contact",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	path := fmt.Sprintf("/api/contacts/block?session=%s", a.sessionName)

	payload := map[string]string{
		"contactId": contactID,
	}

	resp, err := a.client.Post(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to block contact",
			zap.String("session_name", a.sessionName),
			zap.String("contact_id", contactID),
			zap.Error(err))
		return fmt.Errorf("failed to block contact: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Contact blocked successfully",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	return nil
}

// UnblockContact unblocks a contact
func (a *ContactProviderAdapter) UnblockContact(ctx context.Context, contactID string) error {
	a.logger.Info("Unblocking contact",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	path := fmt.Sprintf("/api/contacts/unblock?session=%s", a.sessionName)

	payload := map[string]string{
		"contactId": contactID,
	}

	resp, err := a.client.Post(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to unblock contact",
			zap.String("session_name", a.sessionName),
			zap.String("contact_id", contactID),
			zap.Error(err))
		return fmt.Errorf("failed to unblock contact: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Contact unblocked successfully",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	return nil
}

// UpdateContact creates or updates contact in phone address book
//
// May not work if multiple WhatsApp apps installed on same phone.
func (a *ContactProviderAdapter) UpdateContact(ctx context.Context, contactID string, firstName, lastName string) error {
	a.logger.Info("Updating contact",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID),
		zap.String("first_name", firstName),
		zap.String("last_name", lastName))

	path := fmt.Sprintf("/api/%s/contacts", a.sessionName)

	payload := map[string]string{
		"contactId": contactID,
		"firstName": firstName,
		"lastName":  lastName,
	}

	resp, err := a.client.Put(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to update contact",
			zap.String("session_name", a.sessionName),
			zap.String("contact_id", contactID),
			zap.Error(err))
		return fmt.Errorf("failed to update contact: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Contact updated successfully",
		zap.String("session_name", a.sessionName),
		zap.String("contact_id", contactID))

	return nil
}

// WAHAContact represents WAHA contact response
type WAHAContact struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"number"`
	PushName    *string `json:"pushname,omitempty"`
	ShortName   *string `json:"shortName,omitempty"`
	IsMe        bool    `json:"isMe"`
	IsWAContact bool    `json:"isWAContact"`
	IsBlocked   bool    `json:"isBlocked"`
}

// WAHAContactExistence represents WAHA existence check response
type WAHAContactExistence struct {
	NumberExists struct {
		PhoneNumber  string  `json:"phoneNumber"`
		NumberExists bool    `json:"numberExists"`
		ChatID       *string `json:"chatId,omitempty"`
	} `json:"numberExists"`
}

// WAHAContactAbout represents WAHA about response
type WAHAContactAbout struct {
	About *string `json:"about,omitempty"`
}

// WAHAContactProfilePicture represents WAHA profile picture response
type WAHAContactProfilePicture struct {
	ProfilePictureURL *string `json:"profilePictureURL,omitempty"`
}

// mapWAHAContactToDomain converts WAHA contact to domain contact
func mapWAHAContactToDomain(wc WAHAContact) channel.Contact {
	return channel.Contact{
		ID:          wc.ID,
		Name:        wc.Name,
		PhoneNumber: wc.PhoneNumber,
		PushName:    wc.PushName,
		ShortName:   wc.ShortName,
		IsMe:        wc.IsMe,
		IsWAContact: wc.IsWAContact,
		IsBlocked:   wc.IsBlocked,
	}
}
