package in_memory_repo

import (
	"errors"
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/core/ports/secondary"
	"sync"
)

// InMemoryFriendRepository is an in-memory implementation of FriendRepository.
type InMemoryFriendRepository struct {
	mu      sync.RWMutex
	friends map[string]friendship.Friend // We can use domain models directly or increase isolation by using independent types.
}

func NewInMemoryFriendRepository() *InMemoryFriendRepository {
	return &InMemoryFriendRepository{friends: make(map[string]friendship.Friend)}
}

func (r *InMemoryFriendRepository) StoreFriend(friend friendship.Friend) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.friends[friend.Id.String()] = friend
	return nil
}

func (r *InMemoryFriendRepository) GetFriend(id friendship.FriendID) (friendship.Friend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	foundFriend, ok := r.friends[id.String()]
	if !ok {
		return friendship.Friend{}, errors.New("friend not found")
	}
	return foundFriend, nil
}

func (r *InMemoryFriendRepository) GetAllFriends() (map[string]friendship.Friend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	friendData := make(map[string]friendship.Friend, len(r.friends))
	for k, v := range r.friends {
		friendData[k] = v
	}
	return friendData, nil
}

func (r *InMemoryFriendRepository) UpdateFriend(friend friendship.Friend) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.friends[friend.Id.String()]; !exists {
		return errors.New("friend not found")
	}
	r.friends[friend.Id.String()] = friend
	return nil
}

func (r *InMemoryFriendRepository) DeleteFriend(id friendship.FriendID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.friends[id.String()]; !exists {
		return errors.New("friend not found")
	}
	delete(r.friends, id.String())
	return nil
}

func (r *InMemoryFriendRepository) FriendExists(id friendship.FriendID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.friends[id.String()]
	return exists, nil
}

var _ secondary.FriendRepository = (*InMemoryFriendRepository)(nil)
