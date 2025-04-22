package config

import (
	"time"
)

// Holds HTTP server configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Holds Kafka consumer configuration
type KafkaConsumerConfig struct {
	Brokers         []string
	Topic           string
	GroupID         string
	SessionTimeout  time.Duration
	HeartbeatInterval time.Duration
}

// Holds Kafka producer configuration
type KafkaProducerConfig struct {
	Brokers          []string
	TopicHigh        string
	TopicMedium      string
	TopicLow         string
	RetryMax         int
	RequiredAcks     int
	DeliveryReport   bool
	Partitions       int
	ReplicationFactor int
}

// Holds all configuration for the service
type Config struct {
	Server          ServerConfig
	KafkaConsumer   KafkaConsumerConfig
	KafkaProducer   KafkaProducerConfig
	ShutdownTimeout time.Duration
}

// Provides default configuration values
var DefaultConfig = Config{
	Server: ServerConfig{
		Port:         8081,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	},
	KafkaConsumer: KafkaConsumerConfig{
		Brokers:          []string{"localhost:9092"},
		Topic:            "notifications.raw",
		GroupID:          "prioritizer-group",
		SessionTimeout:   30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
	},
	KafkaProducer: KafkaProducerConfig{
		Brokers:          []string{"localhost:9092"},
		TopicHigh:        "notifications.priority.high",
		TopicMedium:      "notifications.priority.medium",
		TopicLow:         "notifications.priority.low",
		RetryMax:         3,
		RequiredAcks:     1,
		DeliveryReport:   true,
		Partitions:       3,
		ReplicationFactor: 2,
	},
	ShutdownTimeout: 10 * time.Second,
}

// Loads configuration from environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig

	// Load server config
	LoadIntEnv("SERVER_PORT", &cfg.Server.Port)
	LoadDurationEnv("SERVER_READ_TIMEOUT", &cfg.Server.ReadTimeout)
	LoadDurationEnv("SERVER_WRITE_TIMEOUT", &cfg.Server.WriteTimeout)
	LoadDurationEnv("SERVER_IDLE_TIMEOUT", &cfg.Server.IdleTimeout)
	
	// Load Kafka consumer config
	LoadJSONStringArrayEnv("KAFKA_CONSUMER_BROKERS", &cfg.KafkaConsumer.Brokers)
	LoadStringEnv("KAFKA_CONSUMER_TOPIC", &cfg.KafkaConsumer.Topic)
	LoadStringEnv("KAFKA_CONSUMER_GROUP_ID", &cfg.KafkaConsumer.GroupID)
	LoadDurationEnv("KAFKA_CONSUMER_SESSION_TIMEOUT", &cfg.KafkaConsumer.SessionTimeout)
	LoadDurationEnv("KAFKA_CONSUMER_HEARTBEAT_INTERVAL", &cfg.KafkaConsumer.HeartbeatInterval)
	
	// Load Kafka producer config
	LoadJSONStringArrayEnv("KAFKA_PRODUCER_BROKERS", &cfg.KafkaProducer.Brokers)
	LoadStringEnv("KAFKA_PRODUCER_TOPIC_HIGH", &cfg.KafkaProducer.TopicHigh)
	LoadStringEnv("KAFKA_PRODUCER_TOPIC_MEDIUM", &cfg.KafkaProducer.TopicMedium)
	LoadStringEnv("KAFKA_PRODUCER_TOPIC_LOW", &cfg.KafkaProducer.TopicLow)
	LoadIntEnv("KAFKA_PRODUCER_RETRY_MAX", &cfg.KafkaProducer.RetryMax)
	LoadIntEnv("KAFKA_PRODUCER_REQUIRED_ACKS", &cfg.KafkaProducer.RequiredAcks)
	LoadBoolEnv("KAFKA_PRODUCER_DELIVERY_REPORT", &cfg.KafkaProducer.DeliveryReport)
	
	// Load general config
	LoadDurationEnv("SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout)

	return &cfg, nil
}