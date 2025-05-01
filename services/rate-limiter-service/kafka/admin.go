package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/config"
)

// Handles Kafka topic administration
type TopicManager struct {
	admin  sarama.ClusterAdmin
	topics map[string]bool
}

// Creates a new topic manager
func NewTopicManager(brokers []string) (*TopicManager, error) {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster admin: %w", err)
	}

	topicManager := TopicManager{
		admin:  admin,
		topics: make(map[string]bool),
	}

	return &topicManager, nil
}

// Ensures topic exists with proper configuration
func (tm *TopicManager) EnsureTopicExists(cfg config.KafkaProducerConfig) error {
	if _, exists := tm.topics[cfg.Topic]; exists {
		return nil
	}

	// Check if topic exists in Kafka
	topics, err := tm.admin.ListTopics()
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}

	// Log existing topics for debugging
	log.Println("Checking for topic:", cfg.Topic)

	existingTopic, topicExists := topics[cfg.Topic]
	
	// Create new topic if it doesn't exist
	if !topicExists {
		return tm.createNewTopic(cfg.Topic, cfg.Partitions, cfg.ReplicationFactor)
	}
	
	// Otherwise, update existing topic if needed
	return tm.updateExistingTopic(cfg.Topic, cfg.Partitions, cfg.ReplicationFactor, existingTopic)
}

// Creates a new Kafka topic
func (tm *TopicManager) createNewTopic(topic string, partitions, replicationFactor int) error {
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     int32(partitions),
		ReplicationFactor: int16(replicationFactor),
	}
	
	log.Printf("Creating new topic %s", topic)
	err := tm.admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}
	
	log.Printf("Created topic %s with %d partitions and replication factor %d",
		topic, partitions, replicationFactor)
	
	// Mark this topic as checked
	tm.topics[topic] = true
	return nil
}

// Updates an existing topic if configuration has changed
func (tm *TopicManager) updateExistingTopic(topic string, partitions, replicationFactor int, existingTopic sarama.TopicDetail) error {
	log.Printf("Topic %s already exists with %d partitions and replication factor %d",
		topic, existingTopic.NumPartitions, existingTopic.ReplicationFactor)
	
	// Check if partitions need to be increased
	if existingTopic.NumPartitions < int32(partitions) {
		log.Printf("Attempting to update topic %s to have %d partitions",
			topic, partitions)
			
		err := tm.admin.CreatePartitions(topic, int32(partitions), nil, false)
		if err != nil {
			return fmt.Errorf("failed to update partitions for topic %s: %w", topic, err)
		}
		log.Printf("Updated topic %s to %d partitions", topic, partitions)
	}
	
	// Warn if replication factor differs (can't be changed after creation)
	if existingTopic.ReplicationFactor != int16(replicationFactor) {
		log.Printf("Warning: Topic %s has replication factor %d but configuration specifies %d. "+
			"Replication factor cannot be changed after topic creation.",
			topic, existingTopic.ReplicationFactor, replicationFactor)
	}
	
	// Mark this topic as checked
	tm.topics[topic] = true
	return nil
}

// Close releases resources
func (tm *TopicManager) Close() error {
	if tm.admin != nil {
		return tm.admin.Close()
	}
	return nil
}