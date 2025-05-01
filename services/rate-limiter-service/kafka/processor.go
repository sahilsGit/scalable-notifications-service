package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/models"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/preferences"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/ratelimiter"
)

// Processor handles business logic for processing notifications
type Processor struct {
	rateLimiter       ratelimiter.RateLimiter
	preferencesService preferences.PreferencesService
	producer          Producer
	ctx               context.Context
}

// NewProcessor creates a new notification processor
func NewProcessor(ctx context.Context, rateLimiter ratelimiter.RateLimiter, 
	preferencesService preferences.PreferencesService, producer Producer) *Processor {
	return &Processor{
		ctx:               ctx,
		rateLimiter:       rateLimiter,
		preferencesService: preferencesService,
		producer:          producer,
	}
}

// ProcessMessage processes a notification message
func (p *Processor) ProcessMessage(notification *models.PrioritizedNotification) error {
	start := time.Now()
	
	log.Printf("Processing notification %s for user %s with priority %s",
		notification.ID, notification.UserID, notification.Priority)
	
	// Step 1: Apply rate limiting
	isLimited, err := p.rateLimiter.IsRateLimited(p.ctx, notification)
	if err != nil {
		return fmt.Errorf("rate limiting error: %w", err)
	}
	
	if isLimited {
		log.Printf("Notification %s rate limited for user %s", notification.ID, notification.UserID)
		// Notification is rate limited, stop processing
		return nil
	}
	
	// Step 2: Get user preferences
	userPreferences, err := p.preferencesService.GetUserPreferences(notification.UserID)
	if err != nil {
		return fmt.Errorf("error getting user preferences: %w", err)
	}
	
	// Step 3: Check global opt-out
	if !userPreferences.GlobalOptIn {
		log.Printf("User %s has opted out of all notifications", notification.UserID)
		return nil
	}
	
	// Step 4: Determine delivery channels based on preferences
	channels := p.determineDeliveryChannels(notification, userPreferences)
	
	if len(channels) == 0 {
		log.Printf("No delivery channels enabled for notification %s", notification.ID)
		return nil
	}
	
	// Step 5: Create processed notification with channels
	processedNotification := &models.ProcessedNotification{
		PrioritizedNotification: *notification,
		Channels:               channels,
	}
	
	// Step 6: Send to delivery topic
	if err := p.producer.SendMessage(processedNotification); err != nil {
		return fmt.Errorf("failed to send processed notification: %w", err)
	}
	
	elapsed := time.Since(start)
	log.Printf("Processed notification %s in %v, sending to channels: %v", 
		notification.ID, elapsed, channels)
	
	return nil
}

// determineDeliveryChannels determines which channels to deliver the notification to
func (p *Processor) determineDeliveryChannels(
	notification *models.PrioritizedNotification, 
	userPreferences *preferences.UserPreferences) []string {
	
	var enabledChannels []string
	
	// Check if event-specific preferences exist
	if eventPrefs, exists := userPreferences.EventTypes[notification.EventType]; exists {
		// Use event-specific preferences
		for channel, enabled := range eventPrefs {
			if enabled {
				enabledChannels = append(enabledChannels, channel)
			}
		}
	} else {
		// Fall back to general channel preferences
		for channel, enabled := range userPreferences.Channels {
			if enabled {
				enabledChannels = append(enabledChannels, channel)
			}
		}
	}
	
	// If notification is high priority and no channels are enabled,
	// force delivery to in-app at minimum
	if notification.Priority == models.PriorityHigh && len(enabledChannels) == 0 {
		log.Printf("Forcing in-app channel for high priority notification %s", notification.ID)
		enabledChannels = append(enabledChannels, models.ChannelInApp)
	}
	
	return enabledChannels
}