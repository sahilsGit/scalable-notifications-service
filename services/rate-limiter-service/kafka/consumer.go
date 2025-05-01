package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/models"
)

// PriorityConsumer consumes messages from multiple Kafka topics with priority ordering
type PriorityConsumer interface {
	Start(ctx context.Context, messageHandler func(*models.PrioritizedNotification) error) error
	Close() error
}

// KafkaPriorityConsumer implements the PriorityConsumer interface using Sarama
type KafkaPriorityConsumer struct {
	// Separate consumer group for each priority level
	highConsumerGroup  sarama.ConsumerGroup
	mediumConsumerGroup sarama.ConsumerGroup
	lowConsumerGroup    sarama.ConsumerGroup

	topicHigh     string
	topicMedium   string
	topicLow      string
	readyHigh     chan bool
	readyMedium   chan bool
	readyLow      chan bool
	mu            sync.Mutex

	// Channels for controlling consumption rate between different priority levels
	highPriorityMessages   chan *models.PrioritizedNotification
	mediumPriorityMessages chan *models.PrioritizedNotification
	lowPriorityMessages    chan *models.PrioritizedNotification
}

// Sarama ConsumerGroupHandler implementation for high priority messages
type highPriorityHandler struct {
	ready          chan bool
	messages       chan <- *models.PrioritizedNotification
	mu             sync.Mutex
	isReady        bool
}

// Sarama ConsumerGroupHandler implementation for medium priority messages
type mediumPriorityHandler struct {
	ready          chan bool
	messages       chan<- *models.PrioritizedNotification
	mu             sync.Mutex
	isReady        bool
}

// Sarama ConsumerGroupHandler implementation for low priority messages
type lowPriorityHandler struct {
	ready          chan bool
	messages       chan<- *models.PrioritizedNotification
	mu             sync.Mutex
	isReady        bool
}

// NewPriorityConsumer creates a new Kafka consumer with priority handling
func NewPriorityConsumer(cfg config.KafkaConsumerConfig) (PriorityConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	
	// Create separate consumer groups for each priority level
	highConsumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID+"-high", config)
	if err != nil {
		return nil, err
	}
	
	mediumConsumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID+"-medium", config)
	if err != nil {
		// Close the high consumer group if medium fails
		highConsumerGroup.Close()
		return nil, err
	}
	
	lowConsumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID+"-low", config)
	if err != nil {
		// Close the other consumer groups if low fails
		highConsumerGroup.Close()
		mediumConsumerGroup.Close()
		return nil, err
	}

	consumer := &KafkaPriorityConsumer{
		highConsumerGroup:   highConsumerGroup,
		mediumConsumerGroup: mediumConsumerGroup,
		lowConsumerGroup:    lowConsumerGroup,
		topicHigh:     cfg.TopicHigh,
		topicMedium:   cfg.TopicMedium,
		topicLow:      cfg.TopicLow,
		readyHigh:     make(chan bool),
		readyMedium:   make(chan bool),
		readyLow:      make(chan bool),
		
		// Buffered channels for each priority level
		// Higher priority has larger buffer to ensure it's processed first
		highPriorityMessages:   make(chan *models.PrioritizedNotification, 1000),
		mediumPriorityMessages: make(chan *models.PrioritizedNotification, 500),
		lowPriorityMessages:    make(chan *models.PrioritizedNotification, 100),
	}

	return consumer, nil
}

// Start consuming messages from Kafka
func (c *KafkaPriorityConsumer) Start(ctx context.Context, messageHandler func(*models.PrioritizedNotification) error) error {
	// Create context for consumers
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Create wait group for all goroutines
	wg := &sync.WaitGroup{}
	wg.Add(4) // 3 consumer handlers + 1 processor
	
	// Start high priority consumer
	go func() {
		defer wg.Done()
		handler := &highPriorityHandler{
			ready:    c.readyHigh,
			messages: c.highPriorityMessages,
		}
		
		for {
			if consumerCtx.Err() != nil {
				return
			}
			
			if err := c.highConsumerGroup.Consume(consumerCtx, []string{c.topicHigh}, handler); err != nil {
				log.Printf("Error consuming from high priority topic: %v", err)
			}
			
			if consumerCtx.Err() != nil {
				return
			}
		}
	}()
	
	// Start medium priority consumer
	go func() {
		defer wg.Done()
		handler := &mediumPriorityHandler{
			ready:    c.readyMedium,
			messages: c.mediumPriorityMessages,
		}
		
		for {
			if consumerCtx.Err() != nil {
				return
			}
			
			if err := c.mediumConsumerGroup.Consume(consumerCtx, []string{c.topicMedium}, handler); err != nil {
				log.Printf("Error consuming from medium priority topic: %v", err)
			}
			
			if consumerCtx.Err() != nil {
				return
			}
		}
	}()
	
	// Start low priority consumer
	go func() {
		defer wg.Done()
		handler := &lowPriorityHandler{
			ready:    c.readyLow,
			messages: c.lowPriorityMessages,
		}
		
		for {
			if consumerCtx.Err() != nil {
				return
			}
			
			if err := c.lowConsumerGroup.Consume(consumerCtx, []string{c.topicLow}, handler); err != nil {
				log.Printf("Error consuming from low priority topic: %v", err)
			}
			
			if consumerCtx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Waiting for all priority consumers to start")
	
	// Wait for all consumers to be ready
	<-c.readyHigh
	<-c.readyMedium
	<-c.readyLow
	
	log.Println("All priority consumers are ready")
	
	// Start priority message processor
	go func() {
		defer wg.Done()
		for {
			select {
			case <-consumerCtx.Done():
				log.Println("Priority processor shutting down...")
				return
				
			// First check high priority messages
			case msg := <-c.highPriorityMessages:
				if err := messageHandler(msg); err != nil {
					log.Printf("Error processing high priority message: %v", err)
				}
				
			// If no high priority messages, check medium priority
			case msg := <-c.mediumPriorityMessages:
				// Before processing medium priority, check if high priority is available
				select {
				case highMsg := <-c.highPriorityMessages:
					// Process high priority first
					if err := messageHandler(highMsg); err != nil {
						log.Printf("Error processing high priority message: %v", err)
					}
					// Then process medium priority
					if err := messageHandler(msg); err != nil {
						log.Printf("Error processing medium priority message: %v", err)
					}
				default:
					// No high priority, process medium
					if err := messageHandler(msg); err != nil {
						log.Printf("Error processing medium priority message: %v", err)
					}
				}
				
			// If no high or medium priority, check low priority
			case msg := <-c.lowPriorityMessages:
				// Before processing low priority, check if higher priorities are available
				select {
				case highMsg := <-c.highPriorityMessages:
					// Process high priority first
					if err := messageHandler(highMsg); err != nil {
						log.Printf("Error processing high priority message: %v", err)
					}
					// Then process low priority
					if err := messageHandler(msg); err != nil {
						log.Printf("Error processing low priority message: %v", err)
					}
				case medMsg := <-c.mediumPriorityMessages:
					// Process medium priority first
					if err := messageHandler(medMsg); err != nil {
						log.Printf("Error processing medium priority message: %v", err)
					}
					// Then process low priority
					if err := messageHandler(msg); err != nil {
						log.Printf("Error processing low priority message: %v", err)
					}
				default:
					// No high or medium priority, process low
					if err := messageHandler(msg); err != nil {
						log.Printf("Error processing low priority message: %v", err)
					}
				}
			}
			
			// Add a small sleep to prevent CPU thrashing
			time.Sleep(5 * time.Millisecond)
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Context cancelled, shutting down consumers...")
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	return nil
}

// Close the consumer and release resources
func (c *KafkaPriorityConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var errs []error
	
	if c.highConsumerGroup != nil {
		if err := c.highConsumerGroup.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if c.mediumConsumerGroup != nil {
		if err := c.mediumConsumerGroup.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if c.lowConsumerGroup != nil {
		if err := c.lowConsumerGroup.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if len(errs) > 0 {
		log.Printf("Errors closing consumer groups: %v", errs)
		return errs[0] // Return the first error
	}
	
	return nil
}

// Implementation of ConsumerGroupHandler for high priority messages
// Setup is run at the beginning of a new session
func (h *highPriorityHandler) Setup(session sarama.ConsumerGroupSession) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Mark the consumer as ready
	if !h.isReady {
		close(h.ready)
		h.isReady = true
	}
	
	log.Println("High priority consumer session setup complete")
	return nil
}

// Cleanup is run at the end of a session
func (h *highPriorityHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("High priority consumer session cleanup complete")
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *highPriorityHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Process messages
	for message := range claim.Messages() {
		// Parse message
		var notification models.PrioritizedNotification
		if err := json.Unmarshal(message.Value, &notification); err != nil {
			log.Printf("Error unmarshalling high priority message: %v", err)
			session.MarkMessage(message, "")
			continue
		}
		
		// Set priority explicitly (in case it wasn't set in the message)
		notification.Priority = models.PriorityHigh
		
		// Send to channel for processing
		h.messages <- &notification
		
		// Mark message as processed
		session.MarkMessage(message, "")
		
		log.Printf("Received high priority message from topic %s, partition %d, offset %d", 
			message.Topic, message.Partition, message.Offset)
	}
	
	return nil
}

// Similar implementations for medium and low priority handlers

// Setup is run at the beginning of a new session
func (m *mediumPriorityHandler) Setup(session sarama.ConsumerGroupSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Mark the consumer as ready
	if !m.isReady {
		close(m.ready)
		m.isReady = true
	}
	
	log.Println("Medium priority consumer session setup complete")
	return nil
}

// Cleanup is run at the end of a session
func (m *mediumPriorityHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Medium priority consumer session cleanup complete")
	return nil
}

// ConsumeClaim processes messages from a partition
func (m *mediumPriorityHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Process messages
	for message := range claim.Messages() {
		// Parse message
		var notification models.PrioritizedNotification
		if err := json.Unmarshal(message.Value, &notification); err != nil {
			log.Printf("Error unmarshalling medium priority message: %v", err)
			session.MarkMessage(message, "")
			continue
		}
		
		// Set priority explicitly
		notification.Priority = models.PriorityMedium
		
		// Send to channel for processing
		m.messages <- &notification
		
		// Mark message as processed
		session.MarkMessage(message, "")
		
		log.Printf("Received medium priority message from topic %s, partition %d, offset %d", 
			message.Topic, message.Partition, message.Offset)
	}
	
	return nil
}

// Setup is run at the beginning of a new session
func (l *lowPriorityHandler) Setup(session sarama.ConsumerGroupSession) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Mark the consumer as ready
	if !l.isReady {
		close(l.ready)
		l.isReady = true
	}
	
	log.Println("Low priority consumer session setup complete")
	return nil
}

// Cleanup is run at the end of a session
func (l *lowPriorityHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Low priority consumer session cleanup complete")
	return nil
}

// ConsumeClaim processes messages from a partition
func (l *lowPriorityHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Process messages
	for message := range claim.Messages() {
		// Parse message
		var notification models.PrioritizedNotification
		if err := json.Unmarshal(message.Value, &notification); err != nil {
			log.Printf("Error unmarshalling low priority message: %v", err)
			session.MarkMessage(message, "")
			continue
		}
		
		// Set priority explicitly
		notification.Priority = models.PriorityLow
		
		// Send to channel for processing
		l.messages <- &notification
		
		// Mark message as processed
		session.MarkMessage(message, "")
		
		log.Printf("Received low priority message from topic %s, partition %d, offset %d", 
			message.Topic, message.Partition, message.Offset)
	}
	
	return nil
}