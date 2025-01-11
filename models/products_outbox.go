package models

import "time"

// Outbox represents an outbox record for events waiting to be published to Pulsar.
type Outbox struct {
	ID          uint64    `json:"id"`           // Unique identifier for the outbox entry
	EventType   string    `json:"event_type"`   // Type of the event (e.g., "product_created", "product_updated")
	Payload     string    `json:"payload"`      // JSON-encoded event data
	Status      string    `json:"status"`       // Status of the event ("pending", "processed", "failed")
	CreatedAt   time.Time `json:"created_at"`   // Timestamp when the event was created
	ProcessedAt time.Time `json:"processed_at"` // Timestamp when the event was successfully processed
}
