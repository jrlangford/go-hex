package application

import (
	"context"
	"fmt"
	"time"

	"go_hex/internal/core/handling/domain"
	"go_hex/internal/core/handling/ports/primary"
	"go_hex/internal/core/handling/ports/secondary"
	"go_hex/internal/support/auth"
)

// HandlingReportService implements the primary port for handling reports
type HandlingReportService struct {
	handlingEventRepo secondary.HandlingEventRepository
	eventPublisher    secondary.EventPublisher
}

// NewHandlingReportService creates a new handling report service instance
func NewHandlingReportService(
	handlingEventRepo secondary.HandlingEventRepository,
	eventPublisher secondary.EventPublisher,
) primary.HandlingReportService {
	return &HandlingReportService{
		handlingEventRepo: handlingEventRepo,
		eventPublisher:    eventPublisher,
	}
}

// SubmitHandlingReport processes a handling report from external systems
func (h *HandlingReportService) SubmitHandlingReport(ctx context.Context, report primary.HandlingReport) error {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		return fmt.Errorf("unauthorized handling submission: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionSubmitHandling); err != nil {
		return fmt.Errorf("unauthorized handling submission: %w", err)
	}

	// Parse completion time
	completionTime, err := time.Parse(time.RFC3339, report.CompletionTime)
	if err != nil {
		return fmt.Errorf("invalid completion time format: %w", err)
	}

	// Convert string event type to domain event type
	eventType := domain.HandlingEventType(report.EventType)

	// Create handling event domain object
	handlingEvent, err := domain.NewHandlingEvent(
		report.TrackingId,
		eventType,
		report.Location,
		report.VoyageNumber,
		completionTime,
	)
	if err != nil {
		return fmt.Errorf("failed to create handling event: %w", err)
	}

	// Store the handling event
	if err := h.handlingEventRepo.Store(handlingEvent); err != nil {
		return fmt.Errorf("failed to store handling event: %w", err)
	}

	// Publish domain events
	for _, event := range handlingEvent.GetEvents() {
		if err := h.eventPublisher.Publish(event); err != nil {
			return fmt.Errorf("failed to publish handling event: %w", err)
		}
	}

	return nil
}

// HandlingEventQueryService implements the primary port for querying handling events
type HandlingEventQueryService struct {
	handlingEventRepo secondary.HandlingEventRepository
}

// NewHandlingEventQueryService creates a new handling event query service instance
func NewHandlingEventQueryService(
	handlingEventRepo secondary.HandlingEventRepository,
) primary.HandlingEventQueryService {
	return &HandlingEventQueryService{
		handlingEventRepo: handlingEventRepo,
	}
}

// GetHandlingHistory retrieves the complete handling history for a cargo
func (h *HandlingEventQueryService) GetHandlingHistory(ctx context.Context, trackingId string) (domain.HandlingHistory, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		return domain.HandlingHistory{}, fmt.Errorf("unauthorized handling history access: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionViewHandling); err != nil {
		return domain.HandlingHistory{}, fmt.Errorf("unauthorized handling history access: %w", err)
	}

	events, err := h.handlingEventRepo.FindByTrackingId(trackingId)
	if err != nil {
		return domain.HandlingHistory{}, fmt.Errorf("failed to find handling events for tracking ID %s: %w", trackingId, err)
	}

	history, err := domain.NewHandlingHistory(trackingId, events)
	if err != nil {
		return domain.HandlingHistory{}, fmt.Errorf("failed to create handling history: %w", err)
	}

	return history, nil
}

// ListAllHandlingEvents retrieves all handling events from the repository
func (h *HandlingEventQueryService) ListAllHandlingEvents(ctx context.Context) ([]domain.HandlingEvent, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized handling events access: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionViewHandling); err != nil {
		return nil, fmt.Errorf("unauthorized handling events access: %w", err)
	}

	events, err := h.handlingEventRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all handling events: %w", err)
	}

	return events, nil
}

// GetHandlingEvent retrieves a specific handling event by ID
func (h *HandlingEventQueryService) GetHandlingEvent(ctx context.Context, eventId domain.HandlingEventId) (domain.HandlingEvent, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		return domain.HandlingEvent{}, fmt.Errorf("unauthorized handling event access: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionViewHandling); err != nil {
		return domain.HandlingEvent{}, fmt.Errorf("unauthorized handling event access: %w", err)
	}

	event, err := h.handlingEventRepo.FindById(eventId)
	if err != nil {
		return domain.HandlingEvent{}, fmt.Errorf("failed to find handling event with ID %s: %w", eventId.String(), err)
	}

	return event, nil
}
