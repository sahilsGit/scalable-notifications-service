package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadIntEnv loads an integer value from environment variable
func LoadIntEnv(key string, target *int) {
    if value := os.Getenv(key); value != "" {
        fmt.Sscanf(value, "%d", target)
    }
}

// LoadStringEnv loads a string value from environment variable
func LoadStringEnv(key string, target *string) {
    if value := os.Getenv(key); value != "" {
        *target = value
    }
}

// LoadDurationEnv loads a duration value from environment variable
func LoadDurationEnv(key string, target *time.Duration) {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            *target = duration
        }
    }
}

// LoadBoolEnv loads a boolean value from environment variable
func LoadBoolEnv(key string, target *bool) {
    if value := os.Getenv(key); value != "" {
        *target = value == "true"
    }
}

// LoadJSONStringArrayEnv loads a JSON string array from environment variable
func LoadJSONStringArrayEnv(key string, target *[]string) {
    if value := os.Getenv(key); value != "" {
        var result []string
        if err := json.Unmarshal([]byte(value), &result); err == nil {
            *target = result
        }
    }
}