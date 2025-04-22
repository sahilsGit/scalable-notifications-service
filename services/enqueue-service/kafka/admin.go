package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/config"
)

// Handles Kafka topic administration for this service
type TopicManager struct {
    admin  sarama.ClusterAdmin
    topics map[string]bool
}

// Creates a new TopicManager
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

// Checks if a topic exists and creates if needed
func (tm *TopicManager) EnsureTopicExists(cfg config.KafkaConfig) error {
    if _, exists := tm.topics[cfg.Topic]; exists {
        return nil
    }

    // Check if topic exists in Kafka
    topics, err := tm.admin.ListTopics()
    if err != nil {
        return fmt.Errorf("failed to list topics: %w", err)
    }

    // Log existing topics for debugging
    log.Println("Existing Kafka topics:", getTopicNames(topics))

    existingTopic, topicExists := topics[cfg.Topic]
    
    // Create new topic if it doesn't exist
    if !topicExists {
        return tm.createNewTopic(cfg)
    }
    
    // Otherwise, update existing topic if needed
    return tm.updateExistingTopic(cfg, existingTopic)
}

// Creates a new topic
func (tm *TopicManager) createNewTopic(cfg config.KafkaConfig) error {
    topicDetail := &sarama.TopicDetail{
        NumPartitions:     int32(cfg.Partitions),
        ReplicationFactor: int16(cfg.ReplicationFactor),
    }
    
    log.Printf("Creating new topic %s", cfg.Topic)
    err := tm.admin.CreateTopic(cfg.Topic, topicDetail, false)
    if err != nil {
        return fmt.Errorf("failed to create topic %s: %w", cfg.Topic, err)
    }
    
    log.Printf("Created topic %s with %d partitions and replication factor %d",
        cfg.Topic, cfg.Partitions, cfg.ReplicationFactor)
    
    // Mark this topic as checked
    tm.topics[cfg.Topic] = true
    return nil
}

// Updates an existing topic if configuration has changed
func (tm *TopicManager) updateExistingTopic(cfg config.KafkaConfig, existingTopic sarama.TopicDetail) error {
    log.Printf("Topic %s already exists with %d partitions and replication factor %d",
        cfg.Topic, existingTopic.NumPartitions, existingTopic.ReplicationFactor)
    
    // Check if partitions need to be increased
    if existingTopic.NumPartitions < int32(cfg.Partitions) {
        log.Printf("Attempting to update topic %s to have %d partitions",
            cfg.Topic, cfg.Partitions)
            
        err := tm.admin.CreatePartitions(cfg.Topic, int32(cfg.Partitions), nil, false)
        if err != nil {
            return fmt.Errorf("failed to update partitions for topic %s: %w", cfg.Topic, err)
        }
        log.Printf("Updated topic %s to %d partitions", cfg.Topic, cfg.Partitions)
    }
    
    // Warn if replication factor differs (can't be changed after creation)
    if existingTopic.ReplicationFactor != int16(cfg.ReplicationFactor) {
        log.Printf("Warning: Topic %s has replication factor %d but configuration specifies %d. "+
            "Replication factor cannot be changed after topic creation.",
            cfg.Topic, existingTopic.ReplicationFactor, cfg.ReplicationFactor)
    }
    
    // Mark this topic as checked
    tm.topics[cfg.Topic] = true
    return nil
}

// Helper function to get topic names for logging
func getTopicNames(topics map[string]sarama.TopicDetail) []string {
    names := make([]string, 0, len(topics))
    for name := range topics {
        names = append(names, name)
    }
    return names
}

// Close releases resources
func (tm *TopicManager) Close() error {
    
    if tm.admin != nil {
        return tm.admin.Close()
    }
    return nil
}