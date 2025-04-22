package config

import (
	"time"
)

// HTTP server config
type ServerConfig struct {
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// Kafka Topic config
type KafkaConfig struct {
    Brokers          []string
    Topic            string
    RetryMax         int
    RequiredAcks     int
    DeliveryReport   bool
    Partitions       int  
    ReplicationFactor int
}

// Main config
type Config struct {
    Server          ServerConfig
    Kafka           KafkaConfig
    ShutdownTimeout time.Duration
}

// DefaultConfig
var DefaultConfig = Config{
    Server: ServerConfig{
        Port:         8080,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    },
    Kafka: KafkaConfig{
        Brokers:          []string{"localhost:9092"}, // one for now
        Topic:            "notifications.raw",
        RetryMax:         3,
        RequiredAcks:     1,
        DeliveryReport:   true,
        Partitions:       3,
        ReplicationFactor: 2,
    },
    ShutdownTimeout: 10 * time.Second,
}

// Loads config from environment variables
func Load() (*Config, error) {
    cfg := DefaultConfig

    // Server config
    LoadIntEnv("SERVER_PORT", &cfg.Server.Port)
    LoadDurationEnv("SERVER_READ_TIMEOUT", &cfg.Server.ReadTimeout)
    LoadDurationEnv("SERVER_WRITE_TIMEOUT", &cfg.Server.WriteTimeout)
    LoadDurationEnv("SERVER_IDLE_TIMEOUT", &cfg.Server.IdleTimeout)
    
    // Kafka config
    LoadJSONStringArrayEnv("KAFKA_BROKERS", &cfg.Kafka.Brokers)
    LoadStringEnv("KAFKA_TOPIC", &cfg.Kafka.Topic)
    LoadIntEnv("KAFKA_RETRY_MAX", &cfg.Kafka.RetryMax)
    LoadIntEnv("KAFKA_REQUIRED_ACKS", &cfg.Kafka.RequiredAcks)
    LoadBoolEnv("KAFKA_DELIVERY_REPORT", &cfg.Kafka.DeliveryReport)
    LoadIntEnv("KAFKA_PARTITIONS", &cfg.Kafka.Partitions)
    LoadIntEnv("KAFKA_REPLICATION_FACTOR", &cfg.Kafka.ReplicationFactor)
    
    // General config
    LoadDurationEnv("SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout)

    return &cfg, nil
}