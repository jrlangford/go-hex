package friendship

import (
	"go_hex/internal/core/domain/common"
	"go_hex/internal/support/validation"
	"strings"

	"github.com/google/uuid"
)

// FriendID is a UUID type for friend identifiers.
type FriendID struct {
	uuid.UUID `json:"id" validate:"required,uuid4"`
}

func NewFriendID(id uuid.UUID) FriendID {
	return FriendID{UUID: id}
}

// FriendData is a value object that describes the friend.
type FriendData struct {
	Name  string `json:"name" validate:"required,friend_name"`
	Title string `json:"title,omitempty" validate:"omitempty,max=50"`
}

func NewFriendData(name, title string) (FriendData, error) {
	data := FriendData{
		Name:  strings.TrimSpace(name),
		Title: strings.TrimSpace(title),
	}

	// Use annotation-based validation
	if err := validation.Validate(data); err != nil {
		return FriendData{}, common.NewDomainValidationError(err.Error(), err)
	}

	return data, nil
}

// Friend is an entity representing a person in the system.
type Friend struct {
	common.BaseModel[FriendID]
	FriendData
}

func NewFriend(data FriendData) (Friend, error) {
	friendID := NewFriendID(uuid.New())
	friend := Friend{
		BaseModel:  common.NewBaseModel(friendID),
		FriendData: data,
	}

	// Queue the friend created event
	event, err := NewFriendCreatedEvent(friendID, data)
	if err != nil {
		return Friend{}, err
	}
	friend.AddEvent(event)

	return friend, nil
}

func (f *Friend) UpdateData(data FriendData) error {
	if err := validation.Validate(data); err != nil {
		return common.NewDomainValidationError(err.Error(), err)
	}

	// Store old data for event
	oldData := f.FriendData

	// Update the data
	f.FriendData = data
	f.Touch()

	// Queue the friend data updated event
	event, err := NewFriendDataUpdatedEvent(f.Id, oldData, data)
	if err != nil {
		return err
	}
	f.AddEvent(event)

	return nil
}

func (f *Friend) GetDisplayName() string {
	if f.Title != "" {
		return f.Title + " " + f.Name
	}
	return f.Name
}

func (f *Friend) GetGreeting() string {
	return "Hello " + f.GetDisplayName() + "!"
}

// MarkForDeletion marks the friend for deletion and queues a deletion event
func (f *Friend) MarkForDeletion() error {
	event, err := NewFriendDeletedEvent(f.Id, f.FriendData)
	if err != nil {
		return err
	}
	f.AddEvent(event)
	return nil
}
