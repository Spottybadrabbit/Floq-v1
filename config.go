package main

import (
    "encoding/json"
    "fmt"
    "os"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     string `json:"port"`
    Database string `json:"database"`
    User     string `json:"user"`
    Password string `json:"password"`
    SSLMode  string `json:"sslmode"`
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() DatabaseConfig {
    return DatabaseConfig{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     getEnv("DB_PORT", "5432"),
        Database: getEnv("DB_NAME", "postgres"),
        User:     getEnv("DB_USER", "postgres"),
        Password: getEnv("DB_PASSWORD", ""),
        SSLMode:  getEnv("DB_SSLMODE", "disable"),
    }
}

// LoadConfigFromFile loads database configuration from JSON file
func LoadConfigFromFile(filename string) (DatabaseConfig, error) {
    var config DatabaseConfig
    
    data, err := os.ReadFile(filename)
    if err != nil {
        return config, fmt.Errorf("failed to read config file: %w", err)
    }
    
    err = json.Unmarshal(data, &config)
    if err != nil {
        return config, fmt.Errorf("failed to parse config file: %w", err)
    }
    
    return config, nil
}

// SaveConfigToFile saves database configuration to JSON file
func SaveConfigToFile(config DatabaseConfig, filename string) error {
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }
    
    return nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// ValidateConfig validates database configuration
func ValidateConfig(config DatabaseConfig) error {
    if config.Host == "" {
        return fmt.Errorf("database host is required")
    }
    if config.Database == "" {
        return fmt.Errorf("database name is required")
    }
    if config.User == "" {
        return fmt.Errorf("database user is required")
    }
    if config.Port == "" {
        config.Port = "5432"
    }
    if config.SSLMode == "" {
        config.SSLMode = "disable"
    }
    return nil
}