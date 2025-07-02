package http

import (
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/support/validation"
)

type FriendDTO struct {
	ID    string `json:"id" validate:"omitempty,uuid"`
	Name  string `json:"name" validate:"required,friend_name"`
	Title string `json:"title,omitempty" validate:"omitempty,max=50"`
}

func (f FriendDTO) ToDomain() (friendship.FriendData, error) {
	if err := validation.Validate(f); err != nil {
		return friendship.FriendData{}, err
	}
	return friendship.NewFriendData(f.Name, f.Title)
}

func NewFriendDTOFromDomain(f friendship.Friend) FriendDTO {
	return FriendDTO{
		ID:    f.Id.String(),
		Name:  f.Name,
		Title: f.Title,
	}
}
