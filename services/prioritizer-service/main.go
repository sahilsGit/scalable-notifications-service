package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/kafka"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/prioritizers"
	"github.com/sahilsGit/scalable-notifications-service/services/prioritizer-service/validators"
)

func main() {
	log.Println("Starting Prioritizer Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create validator and prioritizer
	validator := validators.NewValidator()
	prioritizer := prioritizers.NewPrioritizer()

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.KafkaProducer)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Create the processor
	processor := kafka.NewProcessor(validator, prioritizer, producer)

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.KafkaConsumer)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Create a context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v, initiating shutdown", sig)
		cancel()
	}()

	// Start the consumer
	log.Println("Starting Kafka consumer...")
	go func() {
		if err := consumer.Start(ctx, processor.ProcessMessage); err != nil {
			log.Fatalf("Failed to start consumer: %v", err)
		}
	}()

	log.Println("Prioritizer Service started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Context canceled, shutting down...")

	// Create a new context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	// Wait for shutdown timeout
	<-shutdownCtx.Done()
	
	log.Println("Prioritizer Service shut down")
}