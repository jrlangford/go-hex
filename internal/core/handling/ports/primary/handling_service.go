package primary

import (
	"context"
	"go_hex/internal/core/handling/domain"
)

// HandlingReportService defines the primary port for receiving handling reports from external systems
type HandlingReportService interface {
	// SubmitHandlingReport receives a report from an external system about a handling event
	SubmitHandlingReport(ctx context.Context, report HandlingReport) error
}

// HandlingEventQueryService defines the primary port for querying handling events
type HandlingEventQueryService interface {
	// GetHandlingHistory retrieves the complete handling history for a cargo
	GetHandlingHistory(ctx context.Context, trackingId string) (domain.HandlingHistory, error)

	// GetHandlingEvent retrieves a specific handling event by ID
	GetHandlingEvent(ctx context.Context, eventId domain.HandlingEventId) (domain.HandlingEvent, error)

	// ListAllHandlingEvents retrieves all handling events from the repository
	ListAllHandlingEvents(ctx context.Context) ([]domain.HandlingEvent, error)
}

// HandlingReport represents raw data from external systems (Data Transfer Object)
// This is the Anti-Corruption Layer input format
type HandlingReport struct {
	TrackingId     string `json:"tracking_id" validate:"required"`
	EventType      string `json:"event_type" validate:"required"`
	Location       string `json:"location" validate:"required"`
	VoyageNumber   string `json:"voyage_number,omitempty"`
	CompletionTime string `json:"completion_time" validate:"required"` // RFC3339 format
}
