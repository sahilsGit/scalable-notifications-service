package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/models"
)

// Interface for sending messages to Kafka
type Producer interface {
    SendMessage(event *models.NotificationEvent) error
    Close() error
}

// Main producer Implements the Producer interface using Sarama
type KafkaProducer struct {
    producer sarama.SyncProducer
    topic    string
}

// Creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig) (Producer, error) {

    // Configure Sarama
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.RequiredAcks(cfg.RequiredAcks)
    config.Producer.Retry.Max = cfg.RetryMax
    config.Producer.Return.Successes = true

    // Create the sarama producer
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

// Sends a notification event to Kafka
func (p *KafkaProducer) SendMessage(event *models.NotificationEvent) error {

    // Marshal event to JSON
    payload, err := json.Marshal(event)

    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    // Create message
    msg := &sarama.ProducerMessage{
        Topic: p.topic,
        Key:   sarama.StringEncoder(event.UserID), // Use user ID as key for partitioning
        Value: sarama.ByteEncoder(payload),
    }

    // Send message
    partition, offset, err := p.producer.SendMessage(msg)
    
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }

    log.Printf("Message sent to partition %d at offset %d", partition, offset)
    return nil
}

// Closes the Kafka producer
func (p *KafkaProducer) Close() error {
    return p.producer.Close()
}