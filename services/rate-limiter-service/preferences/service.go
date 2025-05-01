package preferences

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// PreferencesService is responsible for retrieving user preferences
type PreferencesService interface {
	GetUserPreferences(userID string) (*UserPreferences, error)
	Close() error
}

// SQLPreferencesService implements PreferencesService using SQL database
type SQLPreferencesService struct {
	db *sql.DB
}

// Config for preferences service
type Config struct {
	Driver   string
	DSN      string
	MaxConns int
	MaxIdle  int
}

// NewSQLPreferencesService creates a new preferences service
func NewSQLPreferencesService(config Config) (PreferencesService, error) {
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLPreferencesService{
		db: db,
	}, nil
}

// GetUserPreferences retrieves a user's notification preferences
func (s *SQLPreferencesService) GetUserPreferences(userID string) (*UserPreferences, error) {
	// Start with default preferences
	prefs := &UserPreferences{
		UserID:      userID,
		GlobalOptIn: true,
		Channels: map[string]bool{
			"email":    true,
			"in-app":   true,
			"push":     false,
			"whatsapp": false,
			"sms":      false,
		},
		EventTypes: make(map[string]map[string]bool),
	}

	// Query for basic preferences from users table directly
	var globalOptIn bool
	err := s.db.QueryRow("SELECT global_opt_in FROM users WHERE id = ?", userID).Scan(&globalOptIn)
	if err != nil {
		if err == sql.ErrNoRows {
			// No preferences found, use defaults
			log.Print("No user found", userID)
			return prefs, nil
		}
		return nil, fmt.Errorf("error querying user preferences: %w", err)
	}
	prefs.GlobalOptIn = globalOptIn

	// Query for channel preferences
	rows, err := s.db.Query(
		"SELECT channel_name, enabled FROM user_channel_preferences WHERE user_id = ?", 
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying channel preferences: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var channelName string
		var enabled bool
		if err := rows.Scan(&channelName, &enabled); err != nil {
			return nil, fmt.Errorf("error scanning channel preferences: %w", err)
		}
		prefs.Channels[channelName] = enabled
	}

	// Query for event type preferences
	rows, err = s.db.Query(
		"SELECT event_type, channel_name, enabled FROM user_event_preferences WHERE user_id = ?", 
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying event preferences: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventType, channelName string
		var enabled bool
		if err := rows.Scan(&eventType, &channelName, &enabled); err != nil {
			return nil, fmt.Errorf("error scanning event preferences: %w", err)
		}
		
		// Initialize the event type map if it doesn't exist
		if _, ok := prefs.EventTypes[eventType]; !ok {
			prefs.EventTypes[eventType] = make(map[string]bool)
		}
		
		prefs.EventTypes[eventType][channelName] = enabled
	}

	return prefs, nil
}

// Close closes the database connection
func (s *SQLPreferencesService) Close() error {
	return s.db.Close()
}

// MockPreferencesService is a mock implementation for testing
type MockPreferencesService struct{}

// GetUserPreferences retrieves mock user preferences
func (m *MockPreferencesService) GetUserPreferences(userID string) (*UserPreferences, error) {
	// Return mock preferences that are the same for all users
	return &UserPreferences{
		UserID:      userID,
		GlobalOptIn: true,
		Channels: map[string]bool{
			"email":    true,
			"in-app":   true,
			"push":     true,
			"whatsapp": true,
			"sms":      false,
		},
		EventTypes: map[string]map[string]bool{
			"security_alert": {
				"email":    true,
				"in-app":   true,
				"push":     true,
				"whatsapp": false,
				"sms":      false,
			},
			"message_received": {
				"email":    false,
				"in-app":   true,
				"push":     true,
				"whatsapp": true,
				"sms":      false,
			},
			"like": {
				"email":    false,
				"in-app":   true,
				"push":     false,
				"whatsapp": false,
				"sms":      false,
			},
		},
	}, nil
}

// Close for mock implementation
func (m *MockPreferencesService) Close() error {
	return nil
}

// NewMockPreferencesService creates a mock preferences service
func NewMockPreferencesService() PreferencesService {
	return &MockPreferencesService{}
}