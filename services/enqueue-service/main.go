package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/api"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/kafka"
)

func main() {
	log.Println("Starting Enqueue Service...")

	// Load configuration
	cfg, err := config.Load()
	
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.Kafka)

	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	
	defer producer.Close()

	// Initialize and start HTTP server
	server := api.NewServer(cfg.Server, producer)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Notification Service started successfully")

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigCh

	log.Println("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	
	log.Println("Server gracefully stopped")
}