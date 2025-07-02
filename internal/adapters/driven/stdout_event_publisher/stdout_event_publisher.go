package stdout_event_publisher

import (
	"encoding/json"
	"fmt"
	"go_hex/internal/core/domain/shared"
	"go_hex/internal/core/ports/secondary"
)

// StdoutEventPublisher prints events to the console for testing and debugging purposes.
type StdoutEventPublisher struct{}

func NewStdoutEventPublisher() *StdoutEventPublisher {
	return &StdoutEventPublisher{}
}

func (p *StdoutEventPublisher) Publish(event shared.DomainEvent) error {
	b, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("[StdoutEventPublisher] %s Event published:\n", event.EventName())
	fmt.Println(string(b))
	return nil
}

var _ secondary.EventPublisher = (*StdoutEventPublisher)(nil)
