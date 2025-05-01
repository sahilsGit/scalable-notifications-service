package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/rate-limiter-service/kafka"
)

func main() {
	log.Println("Starting Rate Limiter Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize rate limiter
	rateLimiter, err := cfg.CreateRateLimiter()
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}
	defer rateLimiter.Close()
	log.Println("Rate limiter initialized")

	// Initialize preferences service
	preferencesService, err := cfg.CreatePreferencesService()
	if err != nil {
		log.Fatalf("Failed to create preferences service: %v", err)
	}
	defer preferencesService.Close()
	log.Println("Preferences service initialized")

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.KafkaProducer)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()
	log.Println("Kafka producer initialized")

	// Create the processor
	processor := kafka.NewProcessor(ctx, rateLimiter, preferencesService, producer)

	// Initialize Kafka consumer
	consumer, err := kafka.NewPriorityConsumer(cfg.KafkaConsumer)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()
	log.Println("Kafka priority consumer initialized")

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v, initiating shutdown", sig)
		cancel()
	}()

	// Start the consumer
	log.Println("Starting Kafka priority consumer...")
	go func() {
		if err := consumer.Start(ctx, processor.ProcessMessage); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Rate Limiter Service started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Context canceled, shutting down...")

	// Create a new context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	// Wait for shutdown timeout
	<-shutdownCtx.Done()
	
	log.Println("Rate Limiter Service shut down")
}