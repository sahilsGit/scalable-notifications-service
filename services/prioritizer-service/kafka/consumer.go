package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/models"
)

// Interface for consuming messages from Kafka
type Consumer interface {
	Start(ctx context.Context, messageHandler func(*models.NotificationEvent) error) error
	Close() error
}

// Implements the Consumer interface using Sarama
type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	ready         chan bool
	mu            sync.Mutex
}

// Implements sarama.ConsumerGroupHandler
type consumerHandler struct {
	ready          chan bool
	messageHandler func(*models.NotificationEvent) error
	mu             sync.Mutex
	isReady        bool
}

// Creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConsumerConfig) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	
	// Create the consumer group
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	kafkaConsumer := KafkaConsumer{
		consumerGroup: consumerGroup,
		topic:         cfg.Topic,
		ready:         make(chan bool),
	} 

	// Create and return the consumer
	return &kafkaConsumer, nil
}

// Starts consuming messages from Kafka
func (c *KafkaConsumer) Start(ctx context.Context, messageHandler func(*models.NotificationEvent) error) error {
	// Define the consumer handler
	handler := consumerHandler{
		ready:          c.ready,
		messageHandler: messageHandler,
	}

	// Start consuming in a separate goroutine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}

			// Consume messages
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, &handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}
			
			// Continue consuming
			log.Println("Consumer restarting...")
		}
	}()

	// Wait until consumer is ready
	<-c.ready
	log.Println("Consumer is ready")

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Consumer context cancelled, shutting down...")
	wg.Wait()
	
	return nil
}

// Closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.consumerGroup == nil {
		return nil
	}
	
	err := c.consumerGroup.Close()
	if err != nil {
		return err
	}
	
	return nil
}

// Setup is run at the beginning of a new session
func (h *consumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Mark the consumer as ready
	if !h.isReady {
		close(h.ready)
		h.isReady = true
	}
	
	log.Println("Consumer session setup complete")
	return nil
}

// Cleanup is run at the end of a session
func (h *consumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Consumer session cleanup complete")
	return nil
}

// Consumes messages from a partition claim
func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Process messages
	for message := range claim.Messages() {
		// Parse message payload
		var event models.NotificationEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			session.MarkMessage(message, "")
			continue
		}

		// Process the message with the handler
		if err := h.messageHandler(&event); err != nil {
			log.Printf("Error processing message: %v", err)
			// We still mark the message as processed to avoid reprocessing invalid messages
		}

		// Mark message as processed
		session.MarkMessage(message, "")
		
		log.Printf("Processed message from topic %s, partition %d, offset %d", 
			message.Topic, message.Partition, message.Offset)
	}
	
	return nil
}