package application

import (
	"context"
	"go_hex/internal/adapters/driving/http/middleware"
	"go_hex/internal/core/domain/authorization"
	"go_hex/internal/core/domain/friendship"
	primports "go_hex/internal/core/ports/primary"
	secports "go_hex/internal/core/ports/secondary"
	"go_hex/internal/support/auth"
	"log/slog"
)

// HelloService provides greeting use cases with business logic.
type HelloService struct {
	repo      secports.FriendRepository
	publisher secports.EventPublisher
	logger    *slog.Logger
}

// Ensure HelloService implements the Greeter interface.
var _ primports.Greeter = (*HelloService)(nil)

func NewHelloService(repo secports.FriendRepository, publisher secports.EventPublisher, logger *slog.Logger) *HelloService {
	return &HelloService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// getAuthorizationContext extracts the authorization context from the request context.
func (s *HelloService) getAuthorizationContext(ctx context.Context) (*authorization.AuthorizationContext, error) {
	// Use the same context key as the middleware
	claims, ok := ctx.Value(middleware.TokenClaimsKey).(*auth.Claims)
	if !ok || claims == nil {
		s.logger.Warn("No authentication claims found in context")
		return nil, middleware.ErrUnauthorized
	}

	authCtx := claims.ToAuthorizationContext()
	return &authCtx, nil
}

// requirePermission checks if the user has the required permission.
func (s *HelloService) requirePermission(ctx context.Context, permission authorization.Permission) (*authorization.AuthorizationContext, error) {
	authCtx, err := s.getAuthorizationContext(ctx)
	if err != nil {
		return nil, err
	}

	if !authCtx.HasPermission(permission) {
		s.logger.Warn("User lacks required permission",
			"userID", authCtx.UserID,
			"username", authCtx.Username,
			"permission", permission,
			"userRoles", authCtx.Roles)
		return nil, middleware.ErrInsufficientRole
	}

	return authCtx, nil
}

func (s *HelloService) AddFriend(ctx context.Context, data friendship.FriendData) (friendship.Friend, error) {
	// Check authorization
	authCtx, err := s.requirePermission(ctx, authorization.PermissionAddFriend)
	if err != nil {
		return friendship.Friend{}, err
	}

	s.logger.Info("Adding new friend",
		"name", data.Name,
		"userID", authCtx.UserID,
		"username", authCtx.Username)

	newFriend, err := friendship.NewFriend(data)
	if err != nil {
		s.logger.Error("Failed to create friend",
			"error", err,
			"userID", authCtx.UserID)
		return friendship.Friend{}, err
	}

	err = s.repo.StoreFriend(newFriend)
	if err != nil {
		s.logger.Error("Failed to store friend",
			"error", err,
			"friendID", newFriend.ID,
			"userID", authCtx.UserID)
		return friendship.Friend{}, err
	}

	// Publish all queued events from the aggregate
	events := newFriend.GetEvents()
	for _, event := range events {
		err = s.publisher.Publish(event)
		if err != nil {
			s.logger.Error("Failed to publish domain event",
				"error", err,
				"eventName", event.EventName(),
				"friendID", newFriend.ID,
				"userID", authCtx.UserID)
			return friendship.Friend{}, err
		}
	}
	newFriend.ClearEvents()

	s.logger.Info("Friend added successfully",
		"friendID", newFriend.ID,
		"name", newFriend.Name,
		"userID", authCtx.UserID)
	return newFriend, nil
}

func (s *HelloService) Greet(ctx context.Context, id friendship.FriendID) (string, error) {
	// Check authorization
	authCtx, err := s.requirePermission(ctx, authorization.PermissionGreet)
	if err != nil {
		return "", err
	}

	s.logger.Debug("Greeting friend",
		"friendID", id,
		"userID", authCtx.UserID,
		"username", authCtx.Username)

	foundFriend, err := s.repo.GetFriend(id)
	if err != nil {
		s.logger.Error("Failed to get friend for greeting",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return "", err
	}

	greeting := foundFriend.GetGreeting()
	s.logger.Debug("Generated greeting",
		"friendID", id,
		"greeting", greeting,
		"userID", authCtx.UserID)
	return greeting, nil
}

func (s *HelloService) GetAllFriends(ctx context.Context) (map[string]friendship.Friend, error) {
	// Check authorization
	authCtx, err := s.requirePermission(ctx, authorization.PermissionViewFriend)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("Retrieving all friends",
		"userID", authCtx.UserID,
		"username", authCtx.Username)

	friends, err := s.repo.GetAllFriends()
	if err != nil {
		s.logger.Error("Failed to retrieve all friends",
			"error", err,
			"userID", authCtx.UserID)
		return nil, err
	}

	s.logger.Debug("Retrieved friends",
		"count", len(friends),
		"userID", authCtx.UserID)
	return friends, nil
}

func (s *HelloService) GetFriend(ctx context.Context, id friendship.FriendID) (friendship.Friend, error) {
	// Check authorization
	authCtx, err := s.requirePermission(ctx, authorization.PermissionViewFriend)
	if err != nil {
		return friendship.Friend{}, err
	}

	s.logger.Debug("Retrieving friend",
		"friendID", id,
		"userID", authCtx.UserID,
		"username", authCtx.Username)

	foundFriend, err := s.repo.GetFriend(id)
	if err != nil {
		s.logger.Error("Failed to retrieve friend",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return friendship.Friend{}, err
	}

	s.logger.Debug("Retrieved friend",
		"friendID", id,
		"name", foundFriend.Name,
		"userID", authCtx.UserID)
	return foundFriend, nil
}

func (s *HelloService) UpdateFriend(ctx context.Context, id friendship.FriendID, data friendship.FriendData) (friendship.Friend, error) {
	// Check authorization
	authCtx, err := s.requirePermission(ctx, authorization.PermissionUpdateFriend)
	if err != nil {
		return friendship.Friend{}, err
	}

	s.logger.Info("Updating friend",
		"friendID", id,
		"name", data.Name,
		"userID", authCtx.UserID,
		"username", authCtx.Username)

	foundFriend, err := s.repo.GetFriend(id)
	if err != nil {
		s.logger.Error("Failed to find friend for update",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return friendship.Friend{}, err
	}

	err = foundFriend.UpdateData(data)
	if err != nil {
		s.logger.Error("Failed to update friend data",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return friendship.Friend{}, err
	}

	err = s.repo.UpdateFriend(foundFriend)
	if err != nil {
		s.logger.Error("Failed to save updated friend",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return friendship.Friend{}, err
	}

	// Publish all queued events from the aggregate
	events := foundFriend.GetEvents()
	for _, event := range events {
		err = s.publisher.Publish(event)
		if err != nil {
			s.logger.Error("Failed to publish domain event",
				"error", err,
				"eventName", event.EventName(),
				"friendID", id,
				"userID", authCtx.UserID)
			// Don't return error here as the update was successful
		}
	}
	foundFriend.ClearEvents()

	s.logger.Info("Friend updated successfully",
		"friendID", id,
		"name", foundFriend.Name,
		"userID", authCtx.UserID)
	return foundFriend, nil
}

func (s *HelloService) DeleteFriend(ctx context.Context, id friendship.FriendID) error {
	// Check authorization
	authCtx, err := s.requirePermission(ctx, authorization.PermissionDeleteFriend)
	if err != nil {
		return err
	}

	s.logger.Info("Deleting friend",
		"friendID", id,
		"userID", authCtx.UserID,
		"username", authCtx.Username)

	foundFriend, err := s.repo.GetFriend(id)
	if err != nil {
		s.logger.Error("Failed to find friend for deletion",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return err
	}

	// Mark friend for deletion to queue the event
	err = foundFriend.MarkForDeletion()
	if err != nil {
		s.logger.Error("Failed to mark friend for deletion",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return err
	}

	err = s.repo.DeleteFriend(id)
	if err != nil {
		s.logger.Error("Failed to delete friend",
			"error", err,
			"friendID", id,
			"userID", authCtx.UserID)
		return err
	}

	// Publish all queued events from the aggregate
	events := foundFriend.GetEvents()
	for _, event := range events {
		err = s.publisher.Publish(event)
		if err != nil {
			s.logger.Error("Failed to publish domain event",
				"error", err,
				"eventName", event.EventName(),
				"friendID", id,
				"userID", authCtx.UserID)
			// Don't return error here as the deletion was successful
		}
	}
	foundFriend.ClearEvents()

	s.logger.Info("Friend deleted successfully",
		"friendID", id,
		"userID", authCtx.UserID)
	return nil
}
