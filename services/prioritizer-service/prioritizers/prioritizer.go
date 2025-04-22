package prioritizers

import (
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/models"
)

// Prioritizes notifications based on event type
type NotificationPrioritizer struct {
	// Map of event types to priorities
	eventPriorities map[string]string
}

// Creates a new notification prioritizer
func NewPrioritizer() *NotificationPrioritizer {
	eventPriorities := map[string]string{
		// High priority events
		"security_alert":       models.PriorityHigh,
		"account_compromise":   models.PriorityHigh,
		"payment_failed":       models.PriorityHigh,
		"system_outage":        models.PriorityHigh,
		
		// Medium priority events
		"message_received":     models.PriorityMedium,
		"friend_request":       models.PriorityMedium,
		"comment":              models.PriorityMedium,
		"subscription_expiring": models.PriorityMedium,
		
		// Low priority events
		"like":                 models.PriorityLow,
		"follow":               models.PriorityLow,
		"recommendation":       models.PriorityLow,
		"newsletter":           models.PriorityLow,
	}
	
	return &NotificationPrioritizer{
		eventPriorities: eventPriorities,
	}
}

// Determines the priority of a notification based on its event type
func (p *NotificationPrioritizer) Prioritize(notification *models.NotificationEvent) *models.PrioritizedNotification {
	prioritized := &models.PrioritizedNotification{
		NotificationEvent: *notification,
		Priority:          models.PriorityLow, // Default to low priority
	}
	
	// Check if event type has a defined priority
	if priority, exists := p.eventPriorities[notification.EventType]; exists {
		prioritized.Priority = priority
	}
	
	// Additional priority logic could be implemented here:
	return prioritized
}