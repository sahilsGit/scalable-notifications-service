package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/models"
)

// RateLimiter for controlling notification rate
type RateLimiter interface {
	IsRateLimited(ctx context.Context, notification *models.PrioritizedNotification) (bool, error)
	Close() error
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client        *redis.Client
	windowSeconds int           // Time window for rate limiting in seconds
	limits        map[string]int // Limits per priority level
}

// Config for Redis rate limiter
type Config struct {
	Addr          string
	Password      string
	DB            int
	WindowSeconds int
	LimitHigh     int
	LimitMedium   int
	LimitLow      int
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(config Config) (RateLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRateLimiter{
		client:        client,
		windowSeconds: config.WindowSeconds,
		limits: map[string]int{
			models.PriorityHigh:   config.LimitHigh,
			models.PriorityMedium: config.LimitMedium,
			models.PriorityLow:    config.LimitLow,
		},
	}, nil
}

// IsRateLimited checks if the notification exceeds rate limits
func (r *RedisRateLimiter) IsRateLimited(ctx context.Context, notification *models.PrioritizedNotification) (bool, error) {
	// Define keys for different granularities
	userKey := fmt.Sprintf("rate:user:%s", notification.UserID)
	eventTypeKey := fmt.Sprintf("rate:user:%s:event:%s", notification.UserID, notification.EventType)
	
	// Current time for window calculation
	now := time.Now().Unix()
	windowStart := now - int64(r.windowSeconds)
	
	// Remove counts outside the window (using ZREMRANGEBYSCORE)
	if err := r.cleanupOldEntries(ctx, userKey, windowStart); err != nil {
		return false, fmt.Errorf("failed to clean up old entries: %w", err)
	}
	
	if err := r.cleanupOldEntries(ctx, eventTypeKey, windowStart); err != nil {
		return false, fmt.Errorf("failed to clean up old entries: %w", err)
	}
	
	// Get current count for user
	userCount, err := r.getCurrentCount(ctx, userKey)
	if err != nil {
		return false, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get current count for event type
	eventTypeCount, err := r.getCurrentCount(ctx, eventTypeKey)
	if err != nil {
		return false, fmt.Errorf("failed to get event type count: %w", err)
	}
	
	// Check if user has exceeded their limit
	limit := r.getLimitForPriority(notification.Priority)
	
	if userCount >= limit {
		log.Printf("User %s rate limited (count: %d, limit: %d)", 
			notification.UserID, userCount, limit)
		return true, nil
	}
	
	// Additional check for specific event types (e.g., limit "like" notifications)
	if eventTypeCount >= 20 && notification.EventType == "like" {
		log.Printf("User %s rate limited for event type %s (count: %d, limit: 20)", 
			notification.UserID, notification.EventType, eventTypeCount)
		return true, nil
	}
	
	// Increment counters
	if err := r.incrementCounter(ctx, userKey, now); err != nil {
		return false, fmt.Errorf("failed to increment user counter: %w", err)
	}
	
	if err := r.incrementCounter(ctx, eventTypeKey, now); err != nil {
		return false, fmt.Errorf("failed to increment event type counter: %w", err)
	}
	
	return false, nil
}

// cleanupOldEntries removes entries outside the current time window
func (r *RedisRateLimiter) cleanupOldEntries(ctx context.Context, key string, windowStart int64) error {
	_, err := r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10)).Result()
	return err
}

// getCurrentCount gets the current count of items in the sorted set
func (r *RedisRateLimiter) getCurrentCount(ctx context.Context, key string) (int, error) {
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// incrementCounter adds a new entry to the sorted set with the timestamp as score
func (r *RedisRateLimiter) incrementCounter(ctx context.Context, key string, timestamp int64) error {
	// Use timestamp as both score and member to ensure uniqueness
	member := fmt.Sprintf("%d", timestamp)
	_, err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(timestamp),
		Member: member,
	}).Result()
	
	// Set expiration on the key to auto-cleanup
	_, err = r.client.Expire(ctx, key, time.Duration(r.windowSeconds*2)*time.Second).Result()
	return err
}

// getLimitForPriority returns the rate limit based on notification priority
func (r *RedisRateLimiter) getLimitForPriority(priority string) int {
	if limit, exists := r.limits[priority]; exists {
		return limit
	}
	// Default to lowest limit if priority not recognized
	return r.limits[models.PriorityLow]
}

// Close closes the Redis connection
func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}

// MockRateLimiter implements a mock rate limiter for testing
type MockRateLimiter struct {
	ShouldLimit bool // For testing different scenarios
}

// IsRateLimited checks if notification is rate limited (mock)
func (m *MockRateLimiter) IsRateLimited(ctx context.Context, notification *models.PrioritizedNotification) (bool, error) {
	return m.ShouldLimit, nil
}

// Close for mock implementation
func (m *MockRateLimiter) Close() error {
	return nil
}

// NewMockRateLimiter creates a new mock rate limiter
func NewMockRateLimiter(shouldLimit bool) RateLimiter {
	return &MockRateLimiter{
		ShouldLimit: shouldLimit,
	}
}