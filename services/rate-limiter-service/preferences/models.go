package preferences

// UserPreferences represents a user's notification preferences
type UserPreferences struct {
	UserID      string                       `json:"user_id"`
	GlobalOptIn bool                         `json:"global_opt_in"` // Whether user has opted in to any notifications
	Channels    map[string]bool              `json:"channels"`      // Which channels are enabled (email, in-app, etc)
	EventTypes  map[string]map[string]bool   `json:"event_types"`   // Preferences by event type -> channel
}

// ChannelInfo contains information needed to deliver to a channel
type ChannelInfo struct {
	Enabled     bool   `json:"enabled"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	DeviceToken string `json:"device_token,omitempty"`
}