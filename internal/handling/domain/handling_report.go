package domain

// HandlingReport represents raw data from external systems (Data Transfer Object)
// This is the Anti-Corruption Layer input format
type HandlingReport struct {
	TrackingId     string `json:"tracking_id" validate:"required"`
	EventType      string `json:"event_type" validate:"required"`
	Location       string `json:"location" validate:"required"`
	VoyageNumber   string `json:"voyage_number,omitempty"`
	CompletionTime string `json:"completion_time" validate:"required"` // RFC3339 format
}
