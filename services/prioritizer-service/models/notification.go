package models

// Represents the notification events consumed from Kafka
type NotificationEvent struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type"`
	Content   string                 `json:"content,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt int64                  `json:"created_at"`
}

// Extends NotificationEvent with priority information
type PrioritizedNotification struct {
	NotificationEvent
	Priority string `json:"priority"`
}

// Priority levels for notifications
const (
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)