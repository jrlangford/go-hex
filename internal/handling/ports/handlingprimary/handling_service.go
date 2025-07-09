package handlingprimary

import (
	"context"
	"go_hex/internal/handling/handlingdomain"
)

// HandlingReportService defines the primary port for receiving handling reports from external systems
type HandlingReportService interface {
	// SubmitHandlingReport receives a report from an external system about a handling event
	SubmitHandlingReport(ctx context.Context, report handlingdomain.HandlingReport) error
}

// HandlingEventQueryService defines the primary port for querying handling events
type HandlingEventQueryService interface {
	// GetHandlingHistory retrieves the complete handling history for a cargo
	GetHandlingHistory(ctx context.Context, trackingId string) (handlingdomain.HandlingHistory, error)

	// GetHandlingEvent retrieves a specific handling event by ID
	GetHandlingEvent(ctx context.Context, eventId handlingdomain.HandlingEventId) (handlingdomain.HandlingEvent, error)

	// ListAllHandlingEvents retrieves all handling events from the repository
	ListAllHandlingEvents(ctx context.Context) ([]handlingdomain.HandlingEvent, error)
}
