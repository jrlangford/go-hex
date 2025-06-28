package friendship

import (
	"go_hex/internal/core/domain/common"
	"go_hex/internal/support/validation"
	"time"
)

// FriendCreatedEvent is a domain event generated when a friend is created.
type FriendCreatedEvent struct {
	FriendID   FriendID   `json:"friend_id" validate:"required"`
	FriendData FriendData `json:"friend_data" validate:"required"`
	Timestamp  time.Time  `json:"occurred_at" validate:"required"`
}

func (e FriendCreatedEvent) EventName() string {
	return "FriendCreated"
}

func (e FriendCreatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// NewFriendCreatedEvent creates a new FriendCreatedEvent using value objects
func NewFriendCreatedEvent(friendID FriendID, friendData FriendData) (FriendCreatedEvent, error) {
	event := FriendCreatedEvent{
		FriendID:   friendID,
		FriendData: friendData, // Already validated when created
		Timestamp:  time.Now(),
	}

	if err := validation.Validate(event); err != nil {
		return FriendCreatedEvent{}, common.NewDomainValidationError("invalid friend created event", err)
	}

	return event, nil
}

// FriendDataUpdatedEvent is a domain event generated when a friend's data is updated.
type FriendDataUpdatedEvent struct {
	FriendID  FriendID   `json:"friend_id" validate:"required"`
	OldData   FriendData `json:"old_data" validate:"required"`
	NewData   FriendData `json:"new_data" validate:"required"`
	Timestamp time.Time  `json:"occurred_at" validate:"required"`
}

func (e FriendDataUpdatedEvent) EventName() string {
	return "FriendDataUpdated"
}

func (e FriendDataUpdatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// NewFriendDataUpdatedEvent creates a new FriendDataUpdatedEvent using value objects
func NewFriendDataUpdatedEvent(friendID FriendID, oldData, newData FriendData) (FriendDataUpdatedEvent, error) {
	event := FriendDataUpdatedEvent{
		FriendID:  friendID,
		OldData:   oldData, // Already validated when created
		NewData:   newData, // Already validated when created
		Timestamp: time.Now(),
	}

	if err := validation.Validate(event); err != nil {
		return FriendDataUpdatedEvent{}, common.NewDomainValidationError("invalid friend data updated event", err)
	}

	return event, nil
}

// FriendDeletedEvent is a domain event generated when a friend is deleted.
type FriendDeletedEvent struct {
	FriendID    FriendID   `json:"friend_id" validate:"required"`
	DeletedData FriendData `json:"deleted_data" validate:"required"`
	Timestamp   time.Time  `json:"occurred_at" validate:"required"`
}

func (e FriendDeletedEvent) EventName() string {
	return "FriendDeleted"
}

func (e FriendDeletedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// NewFriendDeletedEvent creates a new FriendDeletedEvent with complete data for audit purposes
func NewFriendDeletedEvent(friendID FriendID, deletedData FriendData) (FriendDeletedEvent, error) {
	event := FriendDeletedEvent{
		FriendID:    friendID,
		DeletedData: deletedData, // Store full data for audit/recovery purposes
		Timestamp:   time.Now(),
	}

	if err := validation.Validate(event); err != nil {
		return FriendDeletedEvent{}, common.NewDomainValidationError("invalid friend deleted event", err)
	}

	return event, nil
}

// Ensure events implement DomainEvent interface
var _ common.DomainEvent = (*FriendCreatedEvent)(nil)
var _ common.DomainEvent = (*FriendDataUpdatedEvent)(nil)
var _ common.DomainEvent = (*FriendDeletedEvent)(nil)
