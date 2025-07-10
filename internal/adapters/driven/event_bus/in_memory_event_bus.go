package event_bus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go_hex/internal/support/basedomain"
)

// EventHandler defines a function that handles events
type EventHandler func(ctx context.Context, event basedomain.DomainEvent) error

// InMemoryEventBus is a simple in-memory event bus for inter-module communication
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	logger   *slog.Logger
}

func NewInMemoryEventBus(logger *slog.Logger) *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
	b.logger.Info("Event handler subscribed",
		"event_name", eventName,
		"handler_count", len(b.handlers[eventName]))
}

func (b *InMemoryEventBus) Publish(event basedomain.DomainEvent) error {
	b.mu.RLock()
	handlers, exists := b.handlers[event.EventName()]
	b.mu.RUnlock()

	if !exists {
		b.logger.Debug("No handlers registered for event", "event_name", event.EventName())
		return nil
	}

	b.logger.Info("Publishing event",
		"event_name", event.EventName(),
		"handler_count", len(handlers))

	ctx := context.Background()
	var errors []error

	for i, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Error("Event handler failed",
				"event_name", event.EventName(),
				"handler_index", i,
				"error", err)
			errors = append(errors, fmt.Errorf("handler %d failed for event %s: %w", i, event.EventName(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some event handlers failed: %v", errors)
	}

	return nil
}

func (b *InMemoryEventBus) GetSubscriberCount(eventName string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[eventName])
}

var _ basedomain.EventPublisher = (*InMemoryEventBus)(nil)
