package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
)

// RepositoryProcessor manages processing of multiple repositories
type RepositoryProcessor struct {
    config     DatabaseConfig
    extractor  *GitHubFunctionExtractor
    results    map[string]*ProcessingResult
    logger     *log.Logger
    startTime  time.Time
    totalStats ProcessingStats
}

// ProcessingStats holds aggregate statistics
type ProcessingStats struct {
    TotalRepositories    int `json:"total_repositories"`
    TotalFunctions      int `json:"total_functions"`
    TotalExecuted       int `json:"total_executed"`
    TotalTables         int `json:"total_tables"`
    TotalErrors         int `json:"total_errors"`
    ProcessingTimeMs    int64 `json:"processing_time_ms"`
}

// NewRepositoryProcessor creates a new repository processor
func NewRepositoryProcessor(config DatabaseConfig) *RepositoryProcessor {
    logger := log.New(os.Stdout, "[PROCESSOR] ", log.LstdFlags|log.Lshortfile)
    
    return &RepositoryProcessor{
        config:  config,
        results: make(map[string]*ProcessingResult),
        logger:  logger,
    }
}

// ProcessRepositories processes a list of repository URLs
func (p *RepositoryProcessor) ProcessRepositories(repositories []string) error {
    p.startTime = time.Now()
    p.logger.Printf("Starting processing of %d repositories", len(repositories))
    
    for i, repoURL := range repositories {
        p.logger.Printf("Processing repository %d/%d: %s", i+1, len(repositories), repoURL)
        
        // Create new extractor for each repository
        p.extractor = NewGitHubFunctionExtractor(p.config)
        
        result, err := p.extractor.ProcessRepository(repoURL)
        if err != nil {
            p.logger.Printf("Failed to process repository %s: %v", repoURL, err)
            // Store partial results even on failure
            if result != nil {
                p.results[repoURL] = result
            } else {
                p.results[repoURL] = &ProcessingResult{
                    Errors: []string{err.Error()},
                }
            }
            continue
        }
        
        p.results[repoURL] = result
        p.logger.Printf("Successfully processed repository: %s", repoURL)
        
        // Update aggregate stats
        p.updateStats(result)
    }
    
    p.totalStats.TotalRepositories = len(repositories)
    p.totalStats.ProcessingTimeMs = time.Since(p.startTime).Milliseconds()
    
    p.logger.Printf("Completed processing %d repositories in %dms", 
        len(repositories), p.totalStats.ProcessingTimeMs)
    
    return nil
}

// updateStats updates aggregate statistics
func (p *RepositoryProcessor) updateStats(result *ProcessingResult) {
    p.totalStats.TotalFunctions += len(result.ProcessedFunctions)
    p.totalStats.TotalExecuted += len(result.ExecutedFunctions)
    p.totalStats.TotalTables += len(result.CreatedTables)
    p.totalStats.TotalErrors += len(result.Errors)
}

// PrintSummary prints a detailed summary of processing results
func (p *RepositoryProcessor) PrintSummary() {
    fmt.Println("\n" + "="*60)
    fmt.Println("🎉 PROCESSING SUMMARY")
    fmt.Println("="*60)
    
    fmt.Printf("📊 Total Repositories: %d\n", p.totalStats.TotalRepositories)
    fmt.Printf("⚡ Total Functions Processed: %d\n", p.totalStats.TotalFunctions)
    fmt.Printf("✅ Total Functions Executed: %d\n", p.totalStats.TotalExecuted)
    fmt.Printf("🗄️  Total Tables Created: %d\n", p.totalStats.TotalTables)
    fmt.Printf("❌ Total Errors: %d\n", p.totalStats.TotalErrors)
    fmt.Printf("⏱️  Processing Time: %dms\n", p.totalStats.ProcessingTimeMs)
    
    if p.totalStats.TotalFunctions > 0 {
        successRate := float64(p.totalStats.TotalExecuted) / float64(p.totalStats.TotalFunctions) * 100
        fmt.Printf("📈 Success Rate: %.1f%%\n", successRate)
    }
    
    fmt.Println("\n📋 REPOSITORY DETAILS:")
    fmt.Println("-" * 60)
    
    for repoURL, result := range p.results {
        fmt.Printf("\n🔗 Repository: %s\n", repoURL)
        fmt.Printf("   📝 Functions: %d\n", len(result.ProcessedFunctions))
        fmt.Printf("   ⚡ Executed: %d\n", len(result.ExecutedFunctions))
        fmt.Printf("   🗄️  Tables: %d\n", len(result.CreatedTables))
        fmt.Printf("   ❌ Errors: %d\n", len(result.Errors))
        
        if len(result.CreatedTables) > 0 {
            fmt.Printf("   📋 Created Tables: %s\n", joinStrings(result.CreatedTables, ", "))
        }
        
        if len(result.Errors) > 0 {
            fmt.Printf("   ⚠️  Error Details:\n")
            for _, err := range result.Errors {
                fmt.Printf("      • %s\n", err)
            }
        }
    }
}

// SaveResultsToFile saves processing results to a JSON file
func (p *RepositoryProcessor) SaveResultsToFile(filename string) error {
    // Create comprehensive results structure
    output := struct {
        Summary ProcessingStats                  `json:"summary"`
        Results map[string]*ProcessingResult   `json:"results"`
        GeneratedAt string                     `json:"generated_at"`
    }{
        Summary:     p.totalStats,
        Results:     p.results,
        GeneratedAt: time.Now().Format(time.RFC3339),
    }
    
    data, err := json.MarshalIndent(output, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal results: %w", err)
    }
    
    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        return fmt.Errorf("failed to write results file: %w", err)
    }
    
    p.logger.Printf("Results saved to %s", filename)
    return nil
}

// GetResults returns the processing results
func (p *RepositoryProcessor) GetResults() map[string]*ProcessingResult {
    return p.results
}

// GetStats returns the aggregate statistics
func (p *RepositoryProcessor) GetStats() ProcessingStats {
    return p.totalStats
}

// helper function to join strings
func joinStrings(slice []string, separator string) string {
    if len(slice) == 0 {
        return ""
    }
    
    result := slice[0]
    for i := 1; i < len(slice); i++ {
        result += separator + slice[i]
    }
    return result
}