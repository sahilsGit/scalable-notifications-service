package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/models"
)

// Interface for sending messages to Kafka
type Producer interface {
	SendMessage(notification *models.PrioritizedNotification) error
	Close() error
}

// Implements the Producer interface using Sarama
type KafkaProducer struct {
	producer sarama.SyncProducer
	topics   map[string]string
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
	
	// Ensure all required topics exist
	if err := topicManager.EnsureTopicsExist(cfg); err != nil {
		return nil, fmt.Errorf("failed to ensure topics exist: %w", err)
	}

	// Create the producer
	sarama_producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	// Map priority levels to topics
	topics := map[string]string{
		models.PriorityHigh:   cfg.TopicHigh,
		models.PriorityMedium: cfg.TopicMedium,
		models.PriorityLow:    cfg.TopicLow,
	}

	kafkaProducer := KafkaProducer{
		producer: sarama_producer,
		topics:   topics,
	}

	return &kafkaProducer, nil
}

// Sends a prioritized notification to the appropriate Kafka topic
func (p *KafkaProducer) SendMessage(notification *models.PrioritizedNotification) error {
	// Determine target topic based on priority
	topic, exists := p.topics[notification.Priority]
	if !exists {
		return fmt.Errorf("unknown priority level: %s", notification.Priority)
	}

	// Marshal notification to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(notification.UserID), // Use user ID as key for partitioning
		Value: sarama.ByteEncoder(payload),
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message with priority %s sent to topic %s, partition %d at offset %d", 
		notification.Priority, topic, partition, offset)
	return nil
}

// Closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}