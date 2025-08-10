package main

import (
    "log"
    "os"
)

func main() {
    // Load configuration from environment or file
    var config DatabaseConfig
    var err error

    if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
        config, err = LoadConfigFromFile(configFile)
        if err != nil {
            log.Printf("Failed to load config from file: %v", err)
            config = LoadConfigFromEnv()
        }
    } else {
        config = LoadConfigFromEnv()
    }

    // Validate configuration
    if err := ValidateConfig(config); err != nil {
        log.Fatalf("Invalid configuration: %v", err)
    }

    // Example repositories to process - modify as needed
    repositories := []string{
        "https://github.com/golang/example.git",
        // Add your repositories here
    }

    // Create processor and process repositories
    processor := NewRepositoryProcessor(config)
    
    err = processor.ProcessRepositories(repositories)
    if err != nil {
        log.Fatalf("Failed to process repositories: %v", err)
    }

    // Print summary
    processor.PrintSummary()

    // Save results to file
    if err := processor.SaveResultsToFile("processing_results.json"); err != nil {
        log.Printf("Failed to save results: %v", err)
    }
}