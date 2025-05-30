package models

// Incoming request structure
type NotificationRequest struct {
	UserID		string      `json:"user_id"`
	EventType string      `json:"event_type"`
	Content   string      `json:"content,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Event sent to Kafka
type NotificationEvent struct {
	ID        string      `json:"id"`
	UserID		string      `json:"user_id"`
	EventType string      `json:"event_type"`
	Content   string      `json:"content,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt int64       `json:"created_at"`
}