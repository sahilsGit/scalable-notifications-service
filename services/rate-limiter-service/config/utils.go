package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Loads an integer value from environment variable
func LoadIntEnv(key string, target *int) {
    if value := os.Getenv(key); value != "" {
        fmt.Sscanf(value, "%d", target)
    }
}

// Loads a string value from environment variable
func LoadStringEnv(key string, target *string) {
    if value := os.Getenv(key); value != "" {
        *target = value
    }
}

// Loads a duration value from environment variable
func LoadDurationEnv(key string, target *time.Duration) {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            *target = duration
        }
    }
}

// Loads a boolean value from environment variable
func LoadBoolEnv(key string, target *bool) {
    if value := os.Getenv(key); value != "" {
        *target = value == "true"
    }
}

// Loads a JSON string array from environment variable
func LoadJSONStringArrayEnv(key string, target *[]string) {
    if value := os.Getenv(key); value != "" {
        var result []string
        if err := json.Unmarshal([]byte(value), &result); err == nil {
            *target = result
        }
    }
}