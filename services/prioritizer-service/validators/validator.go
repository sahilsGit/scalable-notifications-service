package validators

import (
	"fmt"
	"time"

	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/models"
)

// NotificationValidator validates notification events
type NotificationValidator struct {
	// Could be extended with additional dependencies if needed, such as:
	// - Database client for checking user existence
	// - User preferences client
}

// Creates a new notification validator
func NewValidator() *NotificationValidator {
	return &NotificationValidator{}
}

// Validates a notification event
func (v *NotificationValidator) Validate(notification *models.NotificationEvent) error {
	// Check for required fields
	if notification.ID == "" {
		return fmt.Errorf("notification ID is required")
	}

	if notification.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if notification.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	// Validate created timestamp (not in the future)
	if notification.CreatedAt > time.Now().Unix() {
		return fmt.Errorf("notification timestamp is in the future")
	}

	// Additional validations could be added here:
	// - Check if user exists
	// - Check if event type is valid
	// - Validate content format based on event type
	// - Check metadata fields

	return nil
}