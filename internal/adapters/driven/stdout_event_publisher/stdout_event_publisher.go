package stdout_event_publisher

import (
	"encoding/json"
	"fmt"
	"go_hex/internal/support/basedomain"
)

// StdoutEventPublisher prints events to the console for testing and debugging purposes.
type StdoutEventPublisher struct{}

func NewStdoutEventPublisher() *StdoutEventPublisher {
	return &StdoutEventPublisher{}
}

func (p *StdoutEventPublisher) Publish(event basedomain.DomainEvent) error {
	b, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("[StdoutEventPublisher] %s Event published:\n", event.EventName())
	fmt.Println(string(b))
	return nil
}

var _ basedomain.EventPublisher = (*StdoutEventPublisher)(nil)
