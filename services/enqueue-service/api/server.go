package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/config"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/kafka"
	"github.com/sahilsGit/scalable-notifications-service/services/enqueue-service/models"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	producer kafka.Producer
}

// NewServer creates a new HTTP server
func NewServer(cfg config.ServerConfig, producer kafka.Producer) *Server {
	mux := http.NewServeMux()
	
	server := Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		producer: producer,
	}

	// Register routes
	mux.HandleFunc("/api/v1/notifications", server.handleCreateNotification)
	mux.HandleFunc("/health", server.handleHealth)

	return &server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleCreateNotification handles notification creation requests
func (s *Server) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.EventType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create notification event
	event := &models.NotificationEvent{
		ID:        generateID(),
		UserID:    req.UserID,
		EventType: req.EventType,
		Content:   req.Content,
		Metadata:  req.Metadata,
		CreatedAt: time.Now().Unix(),
	}

	// Send to Kafka
	if err := s.producer.SendMessage(event); err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		http.Error(w, "Failed to process notification", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      event.ID,
		"status":  "accepted",
		"message": "Notification is being processed",
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// generateID generates a unique ID for notifications
func generateID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}