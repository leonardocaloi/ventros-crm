package chat

import "errors"

var (
	// ErrChatNotFound is returned when a chat is not found
	ErrChatNotFound = errors.New("chat not found")

	// ErrInvalidChatType is returned when chat type is invalid
	ErrInvalidChatType = errors.New("invalid chat type")

	// ErrInvalidChatStatus is returned when chat status is invalid
	ErrInvalidChatStatus = errors.New("invalid chat status")

	// ErrProjectIDRequired is returned when projectID is nil
	ErrProjectIDRequired = errors.New("projectID cannot be nil")

	// ErrTenantIDRequired is returned when tenantID is empty
	ErrTenantIDRequired = errors.New("tenantID cannot be empty")

	// ErrContactIDRequired is returned when contactID is nil (for individual chats)
	ErrContactIDRequired = errors.New("contactID cannot be nil")

	// ErrSubjectRequired is returned when subject is empty (for group/channel)
	ErrSubjectRequired = errors.New("subject cannot be empty for group/channel")

	// ErrCreatorIDRequired is returned when creatorID is nil (for groups)
	ErrCreatorIDRequired = errors.New("creatorID cannot be nil")

	// ErrChatClosed is returned when trying to modify closed chat
	ErrChatClosed = errors.New("cannot modify closed chat")

	// ErrParticipantAlreadyExists is returned when participant already in chat
	ErrParticipantAlreadyExists = errors.New("participant already in chat")

	// ErrParticipantNotFound is returned when participant not in chat
	ErrParticipantNotFound = errors.New("participant not in chat")

	// ErrCannotRemoveFromIndividual is returned when trying to remove participant from individual chat
	ErrCannotRemoveFromIndividual = errors.New("cannot remove participant from individual chat")

	// ErrIndividualChatLimitReached is returned when trying to add multiple contacts to individual chat
	ErrIndividualChatLimitReached = errors.New("individual chat can only have one contact")

	// ErrIndividualChatNoSubject is returned when trying to set subject on individual chat
	ErrIndividualChatNoSubject = errors.New("individual chats don't have subjects")

	// Label-related errors

	// ErrLabelIDRequired is returned when label ID is empty
	ErrLabelIDRequired = errors.New("label ID cannot be empty")

	// ErrLabelNotFound is returned when label is not found on chat
	ErrLabelNotFound = errors.New("label not found on chat")
)
