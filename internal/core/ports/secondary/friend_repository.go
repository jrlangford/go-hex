package secondary

import (
	"go_hex/internal/core/domain/friendship"
)

// FriendRepository defines the secondary port for storing and retrieving friends by unique id.
type FriendRepository interface {
	StoreFriend(friendship.Friend) error
	GetFriend(friendship.FriendID) (friendship.Friend, error)
	GetAllFriends() (map[string]friendship.Friend, error)
	UpdateFriend(friendship.Friend) error
	DeleteFriend(friendship.FriendID) error
	FriendExists(friendship.FriendID) (bool, error)
}
