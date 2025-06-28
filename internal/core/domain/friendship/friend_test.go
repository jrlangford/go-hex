package friendship

import (
	"testing"
	"time"
)

// =============================================================================
// FRIEND VALUE OBJECT AND ENTITY TESTS
// Tests for Friend value objects (FriendData) and entity creation
// =============================================================================

func TestNewFriendData_ValidInput(t *testing.T) {
	name := "Alice"
	title := "Ms."

	friendData, err := NewFriendData(name, title)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if friendData.Name != name {
		t.Errorf("Expected name %s, got %s", name, friendData.Name)
	}

	if friendData.Title != title {
		t.Errorf("Expected title %s, got %s", title, friendData.Title)
	}
}

func TestNewFriendData_EmptyName(t *testing.T) {
	_, err := NewFriendData("", "Ms.")

	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
}

func TestNewFriendData_ShortName(t *testing.T) {
	_, err := NewFriendData("A", "Ms.")

	if err == nil {
		t.Error("Expected error for short name, got nil")
	}
}

func TestNewFriend(t *testing.T) {
	friendData, _ := NewFriendData("Alice", "Ms.")
	friend, err := NewFriend(friendData)
	if err != nil {
		t.Fatalf("Failed to create friend: %v", err)
	}

	if friend.Name != friendData.Name {
		t.Errorf("Expected name %s, got %s", friendData.Name, friend.Name)
	}

	if friend.Title != friendData.Title {
		t.Errorf("Expected title %s, got %s", friendData.Title, friend.Title)
	}

	if friend.Id.String() == "" {
		t.Error("Expected non-empty ID")
	}

	// Verify that creation event was queued
	events := friend.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event after friend creation, got %d", len(events))
	}

	if events[0].EventName() != "FriendCreated" {
		t.Errorf("Expected FriendCreated event, got %s", events[0].EventName())
	}
}

// =============================================================================
// DOMAIN EVENT CONSTRUCTOR TESTS
// Tests for individual event type constructors and their properties
// =============================================================================

func TestNewFriendCreatedEvent(t *testing.T) {
	// Create a test friend
	friendData, err := NewFriendData("Alice", "Ms.")
	if err != nil {
		t.Fatalf("Failed to create friend data: %v", err)
	}
	friend, err := NewFriend(friendData)
	if err != nil {
		t.Fatalf("Failed to create friend: %v", err)
	}

	// Create event using constructor
	event, err := NewFriendCreatedEvent(friend.Id, friend.FriendData)
	if err != nil {
		t.Fatalf("Failed to create friend created event: %v", err)
	}

	// Verify event properties
	if event.EventName() != "FriendCreated" {
		t.Errorf("Expected event name 'FriendCreated', got '%s'", event.EventName())
	}

	if event.FriendID != friend.Id {
		t.Errorf("Expected friend ID '%s', got '%s'", friend.Id, event.FriendID)
	}

	if event.FriendData.Name != friend.FriendData.Name {
		t.Errorf("Expected friend name '%s', got '%s'", friend.FriendData.Name, event.FriendData.Name)
	}

	// Verify timestamp is recent (within last 5 seconds)
	timeDiff := time.Since(event.OccurredAt())
	if timeDiff > 5*time.Second {
		t.Errorf("Event timestamp is too old: %v", timeDiff)
	}
}

// =============================================================================
// FRIEND AGGREGATE EVENT QUEUEING TESTS
// Tests that the Friend aggregate properly queues domain events
// =============================================================================

func TestFriendAggregateEventQueue(t *testing.T) {
	// Test that Friend aggregate properly queues events
	friendData, _ := NewFriendData("Alice", "Ms.")
	friend, err := NewFriend(friendData)
	if err != nil {
		t.Fatalf("Failed to create friend: %v", err)
	}

	// Check that creation event was queued
	events := friend.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event after creation, got %d", len(events))
	}

	if events[0].EventName() != "FriendCreated" {
		t.Errorf("Expected FriendCreated event, got %s", events[0].EventName())
	}

	// Update the friend
	newData, _ := NewFriendData("Alicia", "Dr.")
	err = friend.UpdateData(newData)
	if err != nil {
		t.Fatalf("Failed to update friend: %v", err)
	}

	// Check that update event was queued
	events = friend.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events after update, got %d", len(events))
	}

	if events[1].EventName() != "FriendDataUpdated" {
		t.Errorf("Expected FriendDataUpdated event, got %s", events[1].EventName())
	}

	// Mark for deletion
	err = friend.MarkForDeletion()
	if err != nil {
		t.Fatalf("Failed to mark friend for deletion: %v", err)
	}

	// Check that deletion event was queued
	events = friend.GetEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 events after deletion, got %d", len(events))
	}

	if events[2].EventName() != "FriendDeleted" {
		t.Errorf("Expected FriendDeleted event, got %s", events[2].EventName())
	}

	// Clear events
	friend.ClearEvents()
	events = friend.GetEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events after clearing, got %d", len(events))
	}
}
