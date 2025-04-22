package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/config"
)

// Handles Kafka topic administration for the prioritizer service
type TopicManager struct {
	admin  sarama.ClusterAdmin
	topics map[string]bool
}

// Creates a new topic manager for managing Kafka topics
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

// Ensures all required topics exist with proper configuration
func (tm *TopicManager) EnsureTopicsExist(cfg config.KafkaProducerConfig) error {
	// Ensure all priority topics exist
	if err := tm.ensureTopicExists(cfg.TopicHigh, cfg.Partitions, cfg.ReplicationFactor); err != nil {
		return err
	}

	if err := tm.ensureTopicExists(cfg.TopicMedium, cfg.Partitions, cfg.ReplicationFactor); err != nil {
		return err
	}

	if err := tm.ensureTopicExists(cfg.TopicLow, cfg.Partitions, cfg.ReplicationFactor); err != nil {
		return err
	}

	return nil
}

// Checks if a topic exists and creates it if needed
func (tm *TopicManager) ensureTopicExists(topic string, partitions, replicationFactor int) error {
	// If we've already checked this topic, skip
	if _, exists := tm.topics[topic]; exists {
		return nil
	}

	// Check if topic exists in Kafka
	topics, err := tm.admin.ListTopics()
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}

	// Log existing topics for debugging
	log.Println("Checking for topic:", topic)

	existingTopic, topicExists := topics[topic]

	// Create new topic if it doesn't exist
	if !topicExists {
		return tm.createNewTopic(topic, partitions, replicationFactor)
	}

	// Otherwise, update existing topic if needed
	return tm.updateExistingTopic(topic, partitions, replicationFactor, existingTopic)
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
