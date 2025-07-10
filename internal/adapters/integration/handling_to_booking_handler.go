package integration

import (
	"context"
	"log/slog"

	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/booking/ports/bookingprimary"
	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/support/basedomain"
)

// HandlingToBookingEventHandler handles events from Handling context
// and applies them to the Booking context via Anti-Corruption Layer
type HandlingToBookingEventHandler struct {
	bookingService bookingprimary.BookingService
	logger         *slog.Logger
}

// NewHandlingToBookingEventHandler creates a new event handler for Handling->Booking integration
func NewHandlingToBookingEventHandler(
	bookingService bookingprimary.BookingService,
	logger *slog.Logger,
) *HandlingToBookingEventHandler {
	return &HandlingToBookingEventHandler{
		bookingService: bookingService,
		logger:         logger,
	}
}

// HandleCargoWasHandled processes HandlingEventRegistered events from the Handling context
func (h *HandlingToBookingEventHandler) HandleCargoWasHandled(ctx context.Context, event basedomain.DomainEvent) error {
	h.logger.Info("Handling HandlingEventRegistered event", "event_id", event.EventName())

	// Cast to specific event type (Anti-Corruption Layer)
	handlingEvent, ok := event.(handlingdomain.HandlingEventRegisteredEvent)
	if !ok {
		h.logger.Error("Invalid event type for HandlingEventRegistered handler")
		return bookingdomain.NewDomainValidationError("invalid event type", nil)
	}

	// Convert tracking ID
	trackingId, err := bookingdomain.TrackingIdFromString(handlingEvent.TrackingId)
	if err != nil {
		h.logger.Error("Invalid tracking ID in handling event", "trackingId", handlingEvent.TrackingId, "error", err)
		return err
	}

	// Convert handling event to delivery update format (Anti-Corruption Layer)
	handlingEventSummary := bookingdomain.HandlingEventSummary{
		Type:         string(handlingEvent.EventType),
		Location:     handlingEvent.Location,
		VoyageNumber: handlingEvent.VoyageNumber,
		Timestamp:    handlingEvent.CompletionTime,
	}

	// Update cargo delivery status in Booking context
	if err := h.bookingService.UpdateCargoDelivery(ctx, trackingId, []bookingdomain.HandlingEventSummary{handlingEventSummary}); err != nil {
		h.logger.Error("Failed to update cargo delivery status",
			"trackingId", trackingId.String(),
			"error", err)
		return err
	}

	h.logger.Info("Successfully updated cargo delivery status from handling event",
		"trackingId", trackingId.String(),
		"eventType", handlingEvent.EventType)

	return nil
}

// HandleOtherHandlingEvents can be extended to handle other events from Handling context
func (h *HandlingToBookingEventHandler) HandleOtherHandlingEvents(ctx context.Context, event basedomain.DomainEvent) error {
	h.logger.Debug("Received handling event", "event_name", event.EventName())
	// For now, we only handle CargoWasHandled events
	// Additional handling events can be processed here as needed
	return nil
}
