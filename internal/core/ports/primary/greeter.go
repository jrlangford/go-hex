package primary

import (
	"context"
	"go_hex/internal/core/domain/friendship"
)

// Greeter defines the primary port for greeting use cases.
type Greeter interface {
	AddFriend(ctx context.Context, data friendship.FriendData) (friendship.Friend, error)
	GetFriend(ctx context.Context, id friendship.FriendID) (friendship.Friend, error)
	UpdateFriend(ctx context.Context, id friendship.FriendID, data friendship.FriendData) (friendship.Friend, error)
	DeleteFriend(ctx context.Context, id friendship.FriendID) error
	Greet(ctx context.Context, id friendship.FriendID) (string, error)
	GetAllFriends(ctx context.Context) (map[string]friendship.Friend, error)
}
