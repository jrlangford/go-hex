package application_test

import (
	"bytes"
	"context"
	"errors"
	"go_hex/internal/adapters/driving/http/middleware"
	"go_hex/internal/core/application"
	"go_hex/internal/core/domain/authorization"
	"go_hex/internal/core/domain/shared"
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/core/ports/secondary"
	"go_hex/internal/support/auth"
	"log/slog"
	"testing"
)

// MockFriendRepository implements the FriendRepository interface for testing.
type MockFriendRepository struct {
	friends map[string]friendship.Friend
}

func NewMockFriendRepository() *MockFriendRepository {
	return &MockFriendRepository{
		friends: make(map[string]friendship.Friend),
	}
}

func (m *MockFriendRepository) StoreFriend(friend friendship.Friend) error {
	m.friends[friend.Id.String()] = friend
	return nil
}

func (m *MockFriendRepository) GetFriend(id friendship.FriendID) (friendship.Friend, error) {
	friend, exists := m.friends[id.String()]
	if !exists {
		return friendship.Friend{}, errors.New("friend not found")
	}
	return friend, nil
}

func (m *MockFriendRepository) GetAllFriends() (map[string]friendship.Friend, error) {
	return m.friends, nil
}

func (m *MockFriendRepository) UpdateFriend(friend friendship.Friend) error {
	if _, exists := m.friends[friend.Id.String()]; !exists {
		return errors.New("friend not found")
	}
	m.friends[friend.Id.String()] = friend
	return nil
}

func (m *MockFriendRepository) DeleteFriend(id friendship.FriendID) error {
	if _, exists := m.friends[id.String()]; !exists {
		return errors.New("friend not found")
	}
	delete(m.friends, id.String())
	return nil
}

func (m *MockFriendRepository) FriendExists(id friendship.FriendID) (bool, error) {
	_, exists := m.friends[id.String()]
	return exists, nil
}

// MockEventPublisher implements the EventPublisher interface for testing.
type MockEventPublisher struct {
	events []shared.DomainEvent
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		events: make([]shared.DomainEvent, 0),
	}
}

func (m *MockEventPublisher) Publish(event shared.DomainEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventPublisher) GetEvents() []shared.DomainEvent {
	return m.events
}

var _ secondary.FriendRepository = (*MockFriendRepository)(nil)
var _ secondary.EventPublisher = (*MockEventPublisher)(nil)

// createTestLogger creates a slog.Logger that writes to a buffer for testing
func createTestLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	return logger, &buf
}

// createTestContext creates a context with token claims for testing
func createTestContext(userID, username string, roles []string) context.Context {
	claims, err := auth.NewClaims(
		userID,
		username,
		"",
		roles,
		nil,
	)
	if err != nil {
		panic("Failed to create test claims: " + err.Error())
	}
	// Use the same context key as the middleware
	return context.WithValue(context.Background(), middleware.TokenClaimsKey, claims)
}

// createAdminContext creates a context with admin privileges
func createAdminContext() context.Context {
	return createTestContext("admin-123", "admin", []string{"admin", "user"})
}

// createUserContext creates a context with standard user privileges
func createUserContext() context.Context {
	return createTestContext("user-456", "john.doe", []string{"user"})
}

// createReadOnlyContext creates a context with read-only privileges
func createReadOnlyContext() context.Context {
	return createTestContext("readonly-789", "viewer", []string{"readonly"})
}

// createUnauthenticatedContext creates a context without authentication
func createUnauthenticatedContext() context.Context {
	return context.Background()
}

func TestHelloService_AddFriend(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	friendData, _ := friendship.NewFriendData("Alice", "Ms.")
	ctx := createUserContext() // Use authenticated context
	friend, err := service.AddFriend(ctx, friendData)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if friend.Name != friendData.Name {
		t.Errorf("Expected name %s, got %s", friendData.Name, friend.Name)
	}

	// Check if event was published
	events := publisher.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].EventName() != "FriendCreated" {
		t.Errorf("Expected FriendCreated event, got %s", events[0].EventName())
	}
}

func TestHelloService_Greet(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	friendData, _ := friendship.NewFriendData("Alice", "Ms.")
	ctx := createUserContext() // Use authenticated context
	friend, _ := service.AddFriend(ctx, friendData)

	greeting, err := service.Greet(ctx, friend.Id)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "Hello Ms. Alice!"
	if greeting != expected {
		t.Errorf("Expected greeting %s, got %s", expected, greeting)
	}
}

func TestHelloService_LogsMessages(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, logBuf := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	friendData, _ := friendship.NewFriendData("Alice", "Ms.")
	ctx := createUserContext() // Use authenticated context

	// Add a friend to generate log messages
	friend, err := service.AddFriend(ctx, friendData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that log messages were generated
	logOutput := logBuf.String()
	if logOutput == "" {
		t.Error("Expected log messages to be generated")
	}

	// Check for specific log content
	if !bytes.Contains(logBuf.Bytes(), []byte("Adding new friend")) {
		t.Error("Expected 'Adding new friend' log message")
	}
	if !bytes.Contains(logBuf.Bytes(), []byte("Friend added successfully")) {
		t.Error("Expected 'Friend added successfully' log message")
	}

	// Reset log buffer for greeting test
	logBuf.Reset()

	// Test greeting logs
	_, err = service.Greet(ctx, friend.Id)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check greeting log messages
	if !bytes.Contains(logBuf.Bytes(), []byte("Greeting friend")) {
		t.Error("Expected 'Greeting friend' debug log message")
	}
	if !bytes.Contains(logBuf.Bytes(), []byte("Generated greeting")) {
		t.Error("Expected 'Generated greeting' debug log message")
	}
}

// Authorization Tests

func TestHelloService_Authorization_AddFriend(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	friendData, _ := friendship.NewFriendData("Alice", "Ms.")

	// Test admin can add friends
	t.Run("Admin can add friends", func(t *testing.T) {
		ctx := createAdminContext()
		_, err := service.AddFriend(ctx, friendData)
		if err != nil {
			t.Errorf("Admin should be able to add friends, got error: %v", err)
		}
	})

	// Test user can add friends
	t.Run("User can add friends", func(t *testing.T) {
		ctx := createUserContext()
		_, err := service.AddFriend(ctx, friendData)
		if err != nil {
			t.Errorf("User should be able to add friends, got error: %v", err)
		}
	})

	// Test readonly cannot add friends
	t.Run("ReadOnly cannot add friends", func(t *testing.T) {
		ctx := createReadOnlyContext()
		_, err := service.AddFriend(ctx, friendData)
		if err != middleware.ErrInsufficientRole {
			t.Errorf("ReadOnly should not be able to add friends, expected %v, got %v", middleware.ErrInsufficientRole, err)
		}
	})

	// Test unauthenticated cannot add friends
	t.Run("Unauthenticated cannot add friends", func(t *testing.T) {
		ctx := createUnauthenticatedContext()
		_, err := service.AddFriend(ctx, friendData)
		if err != middleware.ErrUnauthorized {
			t.Errorf("Unauthenticated should not be able to add friends, expected %v, got %v", middleware.ErrUnauthorized, err)
		}
	})
}

func TestHelloService_Authorization_ViewFriend(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	// Setup: Add a friend as admin
	friendData, _ := friendship.NewFriendData("Alice", "Ms.")
	adminCtx := createAdminContext()
	friend, _ := service.AddFriend(adminCtx, friendData)

	// Test admin can view friends
	t.Run("Admin can view friends", func(t *testing.T) {
		ctx := createAdminContext()
		_, err := service.GetFriend(ctx, friend.Id)
		if err != nil {
			t.Errorf("Admin should be able to view friends, got error: %v", err)
		}

		_, err = service.GetAllFriends(ctx)
		if err != nil {
			t.Errorf("Admin should be able to view all friends, got error: %v", err)
		}
	})

	// Test user can view friends
	t.Run("User can view friends", func(t *testing.T) {
		ctx := createUserContext()
		_, err := service.GetFriend(ctx, friend.Id)
		if err != nil {
			t.Errorf("User should be able to view friends, got error: %v", err)
		}

		_, err = service.GetAllFriends(ctx)
		if err != nil {
			t.Errorf("User should be able to view all friends, got error: %v", err)
		}
	})

	// Test readonly can view friends
	t.Run("ReadOnly can view friends", func(t *testing.T) {
		ctx := createReadOnlyContext()
		_, err := service.GetFriend(ctx, friend.Id)
		if err != nil {
			t.Errorf("ReadOnly should be able to view friends, got error: %v", err)
		}

		_, err = service.GetAllFriends(ctx)
		if err != nil {
			t.Errorf("ReadOnly should be able to view all friends, got error: %v", err)
		}
	})

	// Test unauthenticated cannot view friends
	t.Run("Unauthenticated cannot view friends", func(t *testing.T) {
		ctx := createUnauthenticatedContext()
		_, err := service.GetFriend(ctx, friend.Id)
		if err != middleware.ErrUnauthorized {
			t.Errorf("Unauthenticated should not be able to view friends, expected %v, got %v", middleware.ErrUnauthorized, err)
		}

		_, err = service.GetAllFriends(ctx)
		if err != middleware.ErrUnauthorized {
			t.Errorf("Unauthenticated should not be able to view all friends, expected %v, got %v", middleware.ErrUnauthorized, err)
		}
	})
}

func TestHelloService_Authorization_UpdateFriend(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	// Setup: Add a friend as admin
	friendData, _ := friendship.NewFriendData("Alice", "Ms.")
	adminCtx := createAdminContext()
	friend, _ := service.AddFriend(adminCtx, friendData)

	updateData, _ := friendship.NewFriendData("Alice Updated", "Mrs.")

	// Test admin can update friends
	t.Run("Admin can update friends", func(t *testing.T) {
		ctx := createAdminContext()
		_, err := service.UpdateFriend(ctx, friend.Id, updateData)
		if err != nil {
			t.Errorf("Admin should be able to update friends, got error: %v", err)
		}
	})

	// Test user can update friends
	t.Run("User can update friends", func(t *testing.T) {
		ctx := createUserContext()
		_, err := service.UpdateFriend(ctx, friend.Id, updateData)
		if err != nil {
			t.Errorf("User should be able to update friends, got error: %v", err)
		}
	})

	// Test readonly cannot update friends
	t.Run("ReadOnly cannot update friends", func(t *testing.T) {
		ctx := createReadOnlyContext()
		_, err := service.UpdateFriend(ctx, friend.Id, updateData)
		if err != middleware.ErrInsufficientRole {
			t.Errorf("ReadOnly should not be able to update friends, expected %v, got %v", middleware.ErrInsufficientRole, err)
		}
	})

	// Test unauthenticated cannot update friends
	t.Run("Unauthenticated cannot update friends", func(t *testing.T) {
		ctx := createUnauthenticatedContext()
		_, err := service.UpdateFriend(ctx, friend.Id, updateData)
		if err != middleware.ErrUnauthorized {
			t.Errorf("Unauthenticated should not be able to update friends, expected %v, got %v", middleware.ErrUnauthorized, err)
		}
	})
}

func TestHelloService_Authorization_DeleteFriend(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	// Test admin can delete friends
	t.Run("Admin can delete friends", func(t *testing.T) {
		friendData, _ := friendship.NewFriendData("Alice", "Ms.")
		adminCtx := createAdminContext()
		friend, _ := service.AddFriend(adminCtx, friendData)

		err := service.DeleteFriend(adminCtx, friend.Id)
		if err != nil {
			t.Errorf("Admin should be able to delete friends, got error: %v", err)
		}
	})

	// Test user cannot delete friends (only admin can delete)
	t.Run("User cannot delete friends", func(t *testing.T) {
		friendData, _ := friendship.NewFriendData("Bob", "Mr.")
		adminCtx := createAdminContext()
		friend, _ := service.AddFriend(adminCtx, friendData)

		userCtx := createUserContext()
		err := service.DeleteFriend(userCtx, friend.Id)
		if err != middleware.ErrInsufficientRole {
			t.Errorf("User should not be able to delete friends, expected %v, got %v", middleware.ErrInsufficientRole, err)
		}
	})

	// Test readonly cannot delete friends
	t.Run("ReadOnly cannot delete friends", func(t *testing.T) {
		friendData, _ := friendship.NewFriendData("Charlie", "Mr.")
		adminCtx := createAdminContext()
		friend, _ := service.AddFriend(adminCtx, friendData)

		ctx := createReadOnlyContext()
		err := service.DeleteFriend(ctx, friend.Id)
		if err != middleware.ErrInsufficientRole {
			t.Errorf("ReadOnly should not be able to delete friends, expected %v, got %v", middleware.ErrInsufficientRole, err)
		}
	})

	// Test unauthenticated cannot delete friends
	t.Run("Unauthenticated cannot delete friends", func(t *testing.T) {
		friendData, _ := friendship.NewFriendData("Dave", "Mr.")
		adminCtx := createAdminContext()
		friend, _ := service.AddFriend(adminCtx, friendData)

		ctx := createUnauthenticatedContext()
		err := service.DeleteFriend(ctx, friend.Id)
		if err != middleware.ErrUnauthorized {
			t.Errorf("Unauthenticated should not be able to delete friends, expected %v, got %v", middleware.ErrUnauthorized, err)
		}
	})
}

func TestHelloService_Authorization_Greet(t *testing.T) {
	repo := NewMockFriendRepository()
	publisher := NewMockEventPublisher()
	logger, _ := createTestLogger()
	service := application.NewHelloService(repo, publisher, logger)

	// Setup: Add a friend as admin
	friendData, _ := friendship.NewFriendData("Alice", "Ms.")
	adminCtx := createAdminContext()
	friend, _ := service.AddFriend(adminCtx, friendData)

	// Test admin can greet
	t.Run("Admin can greet", func(t *testing.T) {
		ctx := createAdminContext()
		_, err := service.Greet(ctx, friend.Id)
		if err != nil {
			t.Errorf("Admin should be able to greet friends, got error: %v", err)
		}
	})

	// Test user can greet
	t.Run("User can greet", func(t *testing.T) {
		ctx := createUserContext()
		_, err := service.Greet(ctx, friend.Id)
		if err != nil {
			t.Errorf("User should be able to greet friends, got error: %v", err)
		}
	})

	// Test readonly can greet
	t.Run("ReadOnly can greet", func(t *testing.T) {
		ctx := createReadOnlyContext()
		_, err := service.Greet(ctx, friend.Id)
		if err != nil {
			t.Errorf("ReadOnly should be able to greet friends, got error: %v", err)
		}
	})

	// Test unauthenticated cannot greet
	t.Run("Unauthenticated cannot greet", func(t *testing.T) {
		ctx := createUnauthenticatedContext()
		_, err := service.Greet(ctx, friend.Id)
		if err != middleware.ErrUnauthorized {
			t.Errorf("Unauthenticated should not be able to greet friends, expected %v, got %v", middleware.ErrUnauthorized, err)
		}
	})
}

func TestDomain_AuthorizationContext(t *testing.T) {
	// Test role permissions
	t.Run("Admin has all permissions", func(t *testing.T) {
		authCtx := authorization.NewAuthorizationContext("admin-123", "admin", []string{"admin"})

		permissions := []authorization.Permission{
			authorization.PermissionAddFriend,
			authorization.PermissionViewFriend,
			authorization.PermissionUpdateFriend,
			authorization.PermissionDeleteFriend,
			authorization.PermissionGreet,
		}

		for _, permission := range permissions {
			if !authCtx.HasPermission(permission) {
				t.Errorf("Admin should have permission %s", permission)
			}
		}
	})

	t.Run("User has limited permissions", func(t *testing.T) {
		authCtx := authorization.NewAuthorizationContext("user-456", "user", []string{"user"})

		// Should have these permissions
		allowedPermissions := []authorization.Permission{
			authorization.PermissionAddFriend,
			authorization.PermissionViewFriend,
			authorization.PermissionUpdateFriend,
			authorization.PermissionGreet,
		}

		for _, permission := range allowedPermissions {
			if !authCtx.HasPermission(permission) {
				t.Errorf("User should have permission %s", permission)
			}
		}

		// Should NOT have delete permission
		if authCtx.HasPermission(authorization.PermissionDeleteFriend) {
			t.Error("User should not have delete permission")
		}
	})

	t.Run("ReadOnly has minimal permissions", func(t *testing.T) {
		authCtx := authorization.NewAuthorizationContext("readonly-789", "readonly", []string{"readonly"})

		// Should have these permissions
		allowedPermissions := []authorization.Permission{
			authorization.PermissionViewFriend,
			authorization.PermissionGreet,
		}

		for _, permission := range allowedPermissions {
			if !authCtx.HasPermission(permission) {
				t.Errorf("ReadOnly should have permission %s", permission)
			}
		}

		// Should NOT have these permissions
		deniedPermissions := []authorization.Permission{
			authorization.PermissionAddFriend,
			authorization.PermissionUpdateFriend,
			authorization.PermissionDeleteFriend,
		}

		for _, permission := range deniedPermissions {
			if authCtx.HasPermission(permission) {
				t.Errorf("ReadOnly should not have permission %s", permission)
			}
		}
	})
}
