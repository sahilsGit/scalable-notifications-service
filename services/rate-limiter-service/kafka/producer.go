package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/models"
)

// Interface for sending messages to Kafka
type Producer interface {
	SendMessage(notification *models.ProcessedNotification) error
	Close() error
}

// Implements the Producer interface using Sarama
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// Creates a new Kafka producer
func NewProducer(cfg config.KafkaProducerConfig) (Producer, error) {
	// Configure Sarama
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.RequiredAcks(cfg.RequiredAcks)
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Return.Successes = true

	// Create topic manager and ensure topics exist
	topicManager, err := NewTopicManager(cfg.Brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create topic manager: %w", err)
	}
	defer topicManager.Close()
	
	// Ensure topic exists
	if err := topicManager.EnsureTopicExists(cfg); err != nil {
		return nil, fmt.Errorf("failed to ensure topic exists: %w", err)
	}

	// Create the producer
	sarama_producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	kafkaProducer := KafkaProducer{
		producer: sarama_producer,
		topic:    cfg.Topic,
	}

	return &kafkaProducer, nil
}

// Sends a processed notification to Kafka
func (p *KafkaProducer) SendMessage(notification *models.ProcessedNotification) error {
	// Marshal notification to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create message
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(notification.UserID), // Use user ID as key for partitioning
		Value: sarama.ByteEncoder(payload),
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Processed notification sent to topic %s, partition %d at offset %d", 
		p.topic, partition, offset)
	return nil
}

// Closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}