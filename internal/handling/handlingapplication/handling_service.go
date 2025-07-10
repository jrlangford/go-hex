package handlingapplication

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/handling/ports/handlingprimary"
	"go_hex/internal/handling/ports/handlingsecondary"
	"go_hex/internal/support/auth"
)

// HandlingReportService implements the primary port for handling reports
type HandlingReportService struct {
	handlingEventRepo handlingsecondary.HandlingEventRepository
	eventPublisher    handlingsecondary.EventPublisher
	logger            *slog.Logger
}

// NewHandlingReportService creates a new handling report service instance
func NewHandlingReportService(
	handlingEventRepo handlingsecondary.HandlingEventRepository,
	eventPublisher handlingsecondary.EventPublisher,
	logger *slog.Logger,
) handlingprimary.HandlingReportService {
	return &HandlingReportService{
		handlingEventRepo: handlingEventRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
	}
}

// SubmitHandlingReport processes a handling report from external systems
func (h *HandlingReportService) SubmitHandlingReport(ctx context.Context, report handlingdomain.HandlingReport) error {
	h.logger.Info("Processing handling report",
		"trackingId", report.TrackingId,
		"eventType", report.EventType,
		"location", report.Location,
	)

	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		h.logger.Warn("Unauthorized handling submission attempt", "error", err)
		return fmt.Errorf("unauthorized handling submission: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionSubmitHandling); err != nil {
		h.logger.Warn("Unauthorized handling submission attempt", "error", err)
		return fmt.Errorf("unauthorized handling submission: %w", err)
	}

	// Parse completion time
	completionTime, err := time.Parse(time.RFC3339, report.CompletionTime)
	if err != nil {
		h.logger.Error("Invalid completion time format", "error", err, "completionTime", report.CompletionTime)
		return fmt.Errorf("invalid completion time format: %w", err)
	}

	// Convert string event type to domain event type
	eventType := handlingdomain.HandlingEventType(report.EventType)

	// Create handling event domain object
	handlingEvent, err := handlingdomain.NewHandlingEvent(
		report.TrackingId,
		eventType,
		report.Location,
		report.VoyageNumber,
		completionTime,
	)
	if err != nil {
		h.logger.Error("Failed to create handling event", "error", err, "trackingId", report.TrackingId)
		return fmt.Errorf("failed to create handling event: %w", err)
	}

	// Store the handling event
	if err := h.handlingEventRepo.Store(handlingEvent); err != nil {
		h.logger.Error("Failed to store handling event", "error", err, "eventId", handlingEvent.Id.String())
		return fmt.Errorf("failed to store handling event: %w", err)
	}

	h.logger.Info("Handling event stored successfully", "eventId", handlingEvent.Id.String())

	// Publish domain events
	for _, event := range handlingEvent.GetEvents() {
		if err := h.eventPublisher.Publish(event); err != nil {
			h.logger.Error("Failed to publish handling event", "error", err, "eventType", fmt.Sprintf("%T", event))
			return fmt.Errorf("failed to publish handling event: %w", err)
		}
		h.logger.Debug("Published domain event", "eventType", fmt.Sprintf("%T", event))
	}

	h.logger.Info("Handling report processed successfully", "trackingId", report.TrackingId)
	return nil
}

// HandlingEventQueryService implements the primary port for querying handling events
type HandlingEventQueryService struct {
	handlingEventRepo handlingsecondary.HandlingEventRepository
	logger            *slog.Logger
}

// NewHandlingEventQueryService creates a new handling event query service instance
func NewHandlingEventQueryService(
	handlingEventRepo handlingsecondary.HandlingEventRepository,
	logger *slog.Logger,
) handlingprimary.HandlingEventQueryService {
	return &HandlingEventQueryService{
		handlingEventRepo: handlingEventRepo,
		logger:            logger,
	}
}

// GetHandlingHistory retrieves the complete handling history for a cargo
func (h *HandlingEventQueryService) GetHandlingHistory(ctx context.Context, trackingId string) (handlingdomain.HandlingHistory, error) {
	h.logger.Info("Retrieving handling history", "trackingId", trackingId)

	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		h.logger.Warn("Unauthorized handling history access attempt", "error", err)
		return handlingdomain.HandlingHistory{}, fmt.Errorf("unauthorized handling history access: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionViewHandling); err != nil {
		h.logger.Warn("Unauthorized handling history access attempt", "error", err)
		return handlingdomain.HandlingHistory{}, fmt.Errorf("unauthorized handling history access: %w", err)
	}

	events, err := h.handlingEventRepo.FindByTrackingId(trackingId)
	if err != nil {
		h.logger.Error("Failed to find handling events", "error", err, "trackingId", trackingId)
		return handlingdomain.HandlingHistory{}, fmt.Errorf("failed to find handling events for tracking ID %s: %w", trackingId, err)
	}

	history, err := handlingdomain.NewHandlingHistory(trackingId, events)
	if err != nil {
		h.logger.Error("Failed to create handling history", "error", err, "trackingId", trackingId)
		return handlingdomain.HandlingHistory{}, fmt.Errorf("failed to create handling history: %w", err)
	}

	h.logger.Info("Handling history retrieved successfully", "trackingId", trackingId, "eventCount", len(events))
	return history, nil
}

// ListAllHandlingEvents retrieves all handling events from the repository
func (h *HandlingEventQueryService) ListAllHandlingEvents(ctx context.Context) ([]handlingdomain.HandlingEvent, error) {
	h.logger.Info("Retrieving all handling events")

	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		h.logger.Warn("Unauthorized handling events access attempt", "error", err)
		return nil, fmt.Errorf("unauthorized handling events access: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionViewHandling); err != nil {
		h.logger.Warn("Unauthorized handling events access attempt", "error", err)
		return nil, fmt.Errorf("unauthorized handling events access: %w", err)
	}

	events, err := h.handlingEventRepo.FindAll()
	if err != nil {
		h.logger.Error("Failed to retrieve all handling events", "error", err)
		return nil, fmt.Errorf("failed to retrieve all handling events: %w", err)
	}

	h.logger.Info("All handling events retrieved successfully", "eventCount", len(events))
	return events, nil
}

// GetHandlingEvent retrieves a specific handling event by ID
func (h *HandlingEventQueryService) GetHandlingEvent(ctx context.Context, eventId handlingdomain.HandlingEventId) (handlingdomain.HandlingEvent, error) {
	h.logger.Info("Retrieving handling event by ID", "eventId", eventId.String())

	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		h.logger.Warn("Unauthorized handling event access attempt", "error", err)
		return handlingdomain.HandlingEvent{}, fmt.Errorf("unauthorized handling event access: %w", err)
	}
	if err := RequireHandlingPermission(claims, auth.PermissionViewHandling); err != nil {
		h.logger.Warn("Unauthorized handling event access attempt", "error", err)
		return handlingdomain.HandlingEvent{}, fmt.Errorf("unauthorized handling event access: %w", err)
	}

	event, err := h.handlingEventRepo.FindById(eventId)
	if err != nil {
		h.logger.Error("Failed to find handling event", "error", err, "eventId", eventId.String())
		return handlingdomain.HandlingEvent{}, fmt.Errorf("failed to find handling event with ID %s: %w", eventId.String(), err)
	}

	h.logger.Info("Handling event retrieved successfully", "eventId", eventId.String())
	return event, nil
}
