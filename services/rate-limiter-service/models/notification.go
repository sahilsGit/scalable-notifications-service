package models

// PrioritizedNotification represents a notification with priority
type PrioritizedNotification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type"`
	Content   string                 `json:"content,omitempty"`
	Metadata  map[string]any 				 `json:"metadata,omitempty"`
	CreatedAt int64                  `json:"created_at"`
	Priority  string                 `json:"priority"`
}

// ProcessedNotification represents a notification after rate limiting and preference checks
type ProcessedNotification struct {
	PrioritizedNotification
	Channels []string `json:"channels"` // delivery channels (email, in-app, whatsapp, etc.)
}

// Priority levels for notifications
const (
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

// Delivery channels
const (
	ChannelEmail    = "email"
	ChannelInApp    = "in-app"
	ChannelPush     = "push"
	ChannelWhatsApp = "whatsapp"
	ChannelSMS      = "sms"
)