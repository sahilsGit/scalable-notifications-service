package kafka

import (
	"fmt"
	"log"

	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/models"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/prioritizers"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/validators"
)

// Handles the business logic of validating and prioritizing notifications
type Processor struct {
	validator  *validators.NotificationValidator
	prioritizer *prioritizers.NotificationPrioritizer
	producer   Producer
}

// Creates a new notification processor
func NewProcessor(validator *validators.NotificationValidator, prioritizer *prioritizers.NotificationPrioritizer, producer Producer) *Processor {
	processor := Processor{
		validator:  validator,
		prioritizer: prioritizer,
		producer:   producer,
	}

	return &processor
}

// Processes a notification message
func (p *Processor) ProcessMessage(notification *models.NotificationEvent) error {
	// Validate the notification
	if err := p.validator.Validate(notification); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Prioritize the notification
	prioritizedNotification := p.prioritizer.Prioritize(notification)
	
	// Log the prioritization result
	log.Printf("Notification %s prioritized as %s", notification.ID, prioritizedNotification.Priority)
	
	// Send to the appropriate Kafka topic based on priority
	if err := p.producer.SendMessage(prioritizedNotification); err != nil {
		return fmt.Errorf("failed to send prioritized notification: %w", err)
	}
	
	return nil
}