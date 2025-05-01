package config

import (
	"time"

	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/preferences"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/ratelimiter"
)

// Holds Kafka consumer configuration
type KafkaConsumerConfig struct {
	Brokers          []string
	GroupID          string
	TopicHigh        string
	TopicMedium      string
	TopicLow         string
	SessionTimeout   time.Duration
	HeartbeatInterval time.Duration
}

// Holds Kafka producer configuration
type KafkaProducerConfig struct {
	Brokers          []string
	Topic            string
	RetryMax         int
	RequiredAcks     int
	DeliveryReport   bool
	Partitions       int
	ReplicationFactor int
}

// Holds Redis configuration
type RedisConfig struct {
	Addr          string
	Password      string
	DB            int
	WindowSeconds int
	LimitHigh     int
	LimitMedium   int
	LimitLow      int
}

// Holds database configuration
type DatabaseConfig struct {
	Driver   string
	DSN      string
	MaxConns int
	MaxIdle  int
}

// Holds all configuration for the service
type Config struct {
	KafkaConsumer   KafkaConsumerConfig
	KafkaProducer   KafkaProducerConfig
	Redis           RedisConfig
	Database        DatabaseConfig
	ShutdownTimeout time.Duration
	MockMode        bool
}

// Provides default configuration values
var DefaultConfig = Config{
	KafkaConsumer: KafkaConsumerConfig{
		Brokers:          []string{"localhost:9092"},
		GroupID:          "rate-limiter-group",
		TopicHigh:        "notifications.priority.high",
		TopicMedium:      "notifications.priority.medium",
		TopicLow:         "notifications.priority.low",
		SessionTimeout:   30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
	},
	KafkaProducer: KafkaProducerConfig{
		Brokers:          []string{"localhost:9092"},
		Topic:            "notifications.delivery",
		RetryMax:         3,
		RequiredAcks:     1,
		DeliveryReport:   true,
		Partitions:       3,
		ReplicationFactor: 3,
	},
	Redis: RedisConfig{
		Addr:          "localhost:6379",
		Password:      "",
		DB:            0,
		WindowSeconds: 3600, // 1 hour window for rate limiting
		LimitHigh:     100,  // Higher limits for high priority
		LimitMedium:   50,   // Medium limits for medium priority
		LimitLow:      20,   // Lower limits for low priority
	},
	Database: DatabaseConfig{
		Driver:   "mysql",
		DSN:      "",
		MaxConns: 10,
		MaxIdle:  5,
	},
	ShutdownTimeout: 10 * time.Second,
	MockMode:        false, // Set to true for testing without external dependencies
}

// Loads configuration from environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig

	// Load Kafka consumer config
	LoadJSONStringArrayEnv("KAFKA_CONSUMER_BROKERS", &cfg.KafkaConsumer.Brokers)
	LoadStringEnv("KAFKA_CONSUMER_GROUP_ID", &cfg.KafkaConsumer.GroupID)
	LoadStringEnv("KAFKA_CONSUMER_TOPIC_HIGH", &cfg.KafkaConsumer.TopicHigh)
	LoadStringEnv("KAFKA_CONSUMER_TOPIC_MEDIUM", &cfg.KafkaConsumer.TopicMedium)
	LoadStringEnv("KAFKA_CONSUMER_TOPIC_LOW", &cfg.KafkaConsumer.TopicLow)
	LoadDurationEnv("KAFKA_CONSUMER_SESSION_TIMEOUT", &cfg.KafkaConsumer.SessionTimeout)
	LoadDurationEnv("KAFKA_CONSUMER_HEARTBEAT_INTERVAL", &cfg.KafkaConsumer.HeartbeatInterval)
	
	// Load Kafka producer config
	LoadJSONStringArrayEnv("KAFKA_PRODUCER_BROKERS", &cfg.KafkaProducer.Brokers)
	LoadStringEnv("KAFKA_PRODUCER_TOPIC", &cfg.KafkaProducer.Topic)
	LoadIntEnv("KAFKA_PRODUCER_RETRY_MAX", &cfg.KafkaProducer.RetryMax)
	LoadIntEnv("KAFKA_PRODUCER_REQUIRED_ACKS", &cfg.KafkaProducer.RequiredAcks)
	LoadBoolEnv("KAFKA_PRODUCER_DELIVERY_REPORT", &cfg.KafkaProducer.DeliveryReport)
	LoadIntEnv("KAFKA_PRODUCER_PARTITIONS", &cfg.KafkaProducer.Partitions)
	LoadIntEnv("KAFKA_PRODUCER_REPLICATION_FACTOR", &cfg.KafkaProducer.ReplicationFactor)
	
	// Load Redis config
	LoadStringEnv("REDIS_ADDR", &cfg.Redis.Addr)
	LoadStringEnv("REDIS_PASSWORD", &cfg.Redis.Password)
	LoadIntEnv("REDIS_DB", &cfg.Redis.DB)
	LoadIntEnv("REDIS_WINDOW_SECONDS", &cfg.Redis.WindowSeconds)
	LoadIntEnv("REDIS_LIMIT_HIGH", &cfg.Redis.LimitHigh)
	LoadIntEnv("REDIS_LIMIT_MEDIUM", &cfg.Redis.LimitMedium)
	LoadIntEnv("REDIS_LIMIT_LOW", &cfg.Redis.LimitLow)
	
	// Load Database config
	LoadStringEnv("DB_DRIVER", &cfg.Database.Driver)
	LoadStringEnv("DB_DSN", &cfg.Database.DSN)
	LoadIntEnv("DB_MAX_CONNS", &cfg.Database.MaxConns)
	LoadIntEnv("DB_MAX_IDLE", &cfg.Database.MaxIdle)
	
	// Load general config
	LoadDurationEnv("SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout)
	LoadBoolEnv("MOCK_MODE", &cfg.MockMode)

	return &cfg, nil
}

// Creates rate limiter based on configuration
func (c *Config) CreateRateLimiter() (ratelimiter.RateLimiter, error) {
	if c.MockMode {
		return ratelimiter.NewMockRateLimiter(false), nil
	}
	
	return ratelimiter.NewRedisRateLimiter(ratelimiter.Config{
		Addr:          c.Redis.Addr,
		Password:      c.Redis.Password,
		DB:            c.Redis.DB,
		WindowSeconds: c.Redis.WindowSeconds,
		LimitHigh:     c.Redis.LimitHigh,
		LimitMedium:   c.Redis.LimitMedium,
		LimitLow:      c.Redis.LimitLow,
	})
}

// Creates preferences service based on configuration
func (c *Config) CreatePreferencesService() (preferences.PreferencesService, error) {
	if c.MockMode {
		return preferences.NewMockPreferencesService(), nil
	}
	
	return preferences.NewSQLPreferencesService(preferences.Config{
		Driver:   c.Database.Driver,
		DSN:      c.Database.DSN,
		MaxConns: c.Database.MaxConns,
		MaxIdle:  c.Database.MaxIdle,
	})
}