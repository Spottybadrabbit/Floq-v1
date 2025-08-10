package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/go-git/go-git/v5"
    _ "github.com/lib/pq"
)

// FunctionInfo represents extracted function information
type FunctionInfo struct {
    Name        string   `json:"name"`
    FilePath    string   `json:"file_path"`
    PackageName string   `json:"package_name"`
    LineNumber  int      `json:"line_number"`
    Parameters  []string `json:"parameters"`
    ReturnTypes []string `json:"return_types"`
    Comment     string   `json:"comment"`
    IsExported  bool     `json:"is_exported"`
}

// ProcessingResult holds the results of repository processing
type ProcessingResult struct {
    ProcessedFunctions []FunctionInfo `json:"processed_functions"`
    CreatedTables      []string       `json:"created_tables"`
    Errors             []string       `json:"errors"`
    ExecutedFunctions  []string       `json:"executed_functions"`
}

// GitHubFunctionExtractor handles the extraction and execution of functions
type GitHubFunctionExtractor struct {
    dbConfig   DatabaseConfig
    db         *sql.DB
    tempDir    string
    repoPath   string
    logger     *log.Logger
}

// NewGitHubFunctionExtractor creates a new extractor instance
func NewGitHubFunctionExtractor(config DatabaseConfig) *GitHubFunctionExtractor {
    logger := log.New(os.Stdout, "[EXTRACTOR] ", log.LstdFlags|log.Lshortfile)
    
    return &GitHubFunctionExtractor{
        dbConfig: config,
        logger:   logger,
    }
}

// ConnectToDB establishes database connection
func (g *GitHubFunctionExtractor) ConnectToDB() error {
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        g.dbConfig.Host, g.dbConfig.Port, g.dbConfig.User, 
        g.dbConfig.Password, g.dbConfig.Database, g.dbConfig.SSLMode)

    var err error
    g.db, err = sql.Open("postgres", connStr)
    if err != nil {
        return fmt.Errorf("failed to open database connection: %w", err)
    }

    if err = g.db.Ping(); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }

    g.logger.Println("Connected to PostgreSQL database")
    return nil
}

// CloseDB closes the database connection
func (g *GitHubFunctionExtractor) CloseDB() error {
    if g.db != nil {
        return g.db.Close()
    }
    return nil
}

// CloneRepository clones a GitHub repository to a temporary directory
func (g *GitHubFunctionExtractor) CloneRepository(repoURL string) error {
    tempDir, err := ioutil.TempDir("", "repo_*")
    if err != nil {
        return fmt.Errorf("failed to create temp directory: %w", err)
    }

    g.tempDir = tempDir
    g.repoPath = filepath.Join(tempDir, "repo")

    g.logger.Printf("Cloning repository %s to %s", repoURL, g.repoPath)
    
    _, err = git.PlainClone(g.repoPath, false, &git.CloneOptions{
        URL:      repoURL,
        Progress: os.Stdout,
    })

    if err != nil {
        return fmt.Errorf("failed to clone repository: %w", err)
    }

    g.logger.Printf("Repository cloned successfully to %s", g.repoPath)
    return nil
}

// Cleanup removes temporary directories
func (g *GitHubFunctionExtractor) Cleanup() error {
    if g.tempDir != "" {
        return os.RemoveAll(g.tempDir)
    }
    return nil
}

// FindGoFiles recursively finds all Go files in the repository
func (g *GitHubFunctionExtractor) FindGoFiles() ([]string, error) {
    var goFiles []string

    err := filepath.Walk(g.repoPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip vendor, .git, and test files
        if strings.Contains(path, "vendor/") || 
           strings.Contains(path, ".git/") || 
           strings.HasSuffix(info.Name(), "_test.go") {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        if strings.HasSuffix(info.Name(), ".go") && !info.IsDir() {
            goFiles = append(goFiles, path)
        }

        return nil
    })

    return goFiles, err
}

// ExtractFunctionsFromFile parses a Go file and extracts function information
func (g *GitHubFunctionExtractor) ExtractFunctionsFromFile(filePath string) ([]FunctionInfo, error) {
    var functions []FunctionInfo

    // Create a file set for position information
    fset := token.NewFileSet()

    // Parse the file
    node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
    if err != nil {
        return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
    }

    packageName := node.Name.Name

    // Extract functions
    for _, decl := range node.Decls {
        if funcDecl, ok := decl.(*ast.FuncDecl); ok {
            // Skip methods (functions with receivers) and private functions
            if funcDecl.Recv != nil || !ast.IsExported(funcDecl.Name.Name) {
                continue
            }

            function := FunctionInfo{
                Name:        funcDecl.Name.Name,
                FilePath:    filePath,
                PackageName: packageName,
                LineNumber:  fset.Position(funcDecl.Pos()).Line,
                IsExported:  ast.IsExported(funcDecl.Name.Name),
            }

            // Extract parameters
            if funcDecl.Type.Params != nil {
                for _, param := range funcDecl.Type.Params.List {
                    paramType := g.formatType(param.Type)
                    if len(param.Names) > 0 {
                        for _, name := range param.Names {
                            function.Parameters = append(function.Parameters, 
                                fmt.Sprintf("%s %s", name.Name, paramType))
                        }
                    } else {
                        function.Parameters = append(function.Parameters, paramType)
                    }
                }
            }

            // Extract return types
            if funcDecl.Type.Results != nil {
                for _, result := range funcDecl.Type.Results.List {
                    returnType := g.formatType(result.Type)
                    function.ReturnTypes = append(function.ReturnTypes, returnType)
                }
            }

            // Extract comment/documentation
            if funcDecl.Doc != nil {
                function.Comment = funcDecl.Doc.Text()
            }

            functions = append(functions, function)
        }
    }

    return functions, nil
}

// formatType converts an AST type to a string representation
func (g *GitHubFunctionExtractor) formatType(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.StarExpr:
        return "*" + g.formatType(t.X)
    case *ast.ArrayType:
        return "[]" + g.formatType(t.Elt)
    case *ast.MapType:
        return fmt.Sprintf("map[%s]%s", g.formatType(t.Key), g.formatType(t.Value))
    case *ast.SelectorExpr:
        return fmt.Sprintf("%s.%s", g.formatType(t.X), t.Sel.Name)
    case *ast.InterfaceType:
        return "interface{}"
    default:
        return "unknown"
    }
}

// ExecuteFunction attempts to execute a Go function and capture its output
func (g *GitHubFunctionExtractor) ExecuteFunction(function FunctionInfo) (interface{}, error) {
    // Only execute functions with no parameters that return data
    if len(function.Parameters) > 0 {
        return nil, fmt.Errorf("function %s requires parameters, skipping", function.Name)
    }

    // Create a temporary main.go file to execute the function
    mainContent := g.generateMainFile(function)
    
    tempMainPath := filepath.Join(g.tempDir, "temp_main.go")
    err := ioutil.WriteFile(tempMainPath, []byte(mainContent), 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to create temp main file: %w", err)
    }
    defer os.Remove(tempMainPath)

    // Execute the temporary program
    cmd := exec.Command("go", "run", tempMainPath)
    cmd.Dir = g.repoPath // Set working directory to repo path for imports
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to execute function %s: %w", function.Name, err)
    }

    // Try to parse output as JSON
    var result interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        // If not valid JSON, return as string
        return strings.TrimSpace(string(output)), nil
    }

    return result, nil
}

// generateMainFile creates a temporary main.go file to execute a function
func (g *GitHubFunctionExtractor) generateMainFile(function FunctionInfo) string {
    // Extract package import path relative to repo
    relPath, _ := filepath.Rel(g.repoPath, filepath.Dir(function.FilePath))
    
    var importPath string
    if relPath == "." {
        importPath = "."
    } else {
        importPath = "./" + strings.ReplaceAll(relPath, "\\", "/")
    }

    return fmt.Sprintf(`package main

import (
    "encoding/json"
    "fmt"
    "log"
    
    pkg "%s"
)

func main() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Function panicked: %%v", r)
        }
    }()
    
    result := pkg.%s()
    
    // Try to marshal result as JSON
    jsonResult, err := json.Marshal(result)
    if err != nil {
        // If marshaling fails, print as string
        fmt.Print(result)
    } else {
        fmt.Print(string(jsonResult))
    }
}
`, importPath, function.Name)
}

// CreateTableFromData creates a PostgreSQL table based on data structure
func (g *GitHubFunctionExtractor) CreateTableFromData(tableName string, data interface{}) error {
    // Drop table if exists
    dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
    _, err := g.db.Exec(dropQuery)
    if err != nil {
        return fmt.Errorf("failed to drop existing table: %w", err)
    }

    // Determine table structure based on data type
    var createQuery string
    
    switch v := data.(type) {
    case map[string]interface{}:
        columns := []string{"id SERIAL PRIMARY KEY"}
        for key, value := range v {
            columnType := g.getPostgreSQLType(value)
            columns = append(columns, fmt.Sprintf("%s %s", key, columnType))
        }
        createQuery = fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(columns, ", "))
        
    case []interface{}:
        if len(v) > 0 {
            if firstItem, ok := v[0].(map[string]interface{}); ok {
                // Array of objects
                columns := []string{"id SERIAL PRIMARY KEY"}
                for key, value := range firstItem {
                    columnType := g.getPostgreSQLType(value)
                    columns = append(columns, fmt.Sprintf("%s %s", key, columnType))
                }
                createQuery = fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(columns, ", "))
            } else {
                // Array of primitives
                createQuery = fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, value TEXT)", tableName)
            }
        } else {
            createQuery = fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, data JSONB)", tableName)
        }
        
    default:
        // Single value or unknown structure
        createQuery = fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, data JSONB)", tableName)
    }

    _, err = g.db.Exec(createQuery)
    if err != nil {
        return fmt.Errorf("failed to create table %s: %w", tableName, err)
    }

    g.logger.Printf("Created table %s", tableName)
    return nil
}

// getPostgreSQLType maps Go types to PostgreSQL types
func (g *GitHubFunctionExtractor) getPostgreSQLType(value interface{}) string {
    switch value.(type) {
    case int, int32, int64:
        return "INTEGER"
    case float32, float64:
        return "NUMERIC"
    case bool:
        return "BOOLEAN"
    case []interface{}, map[string]interface{}:
        return "JSONB"
    case string:
        return "TEXT"
    default:
        return "TEXT"
    }
}

// InsertDataToTable inserts data into PostgreSQL table
func (g *GitHubFunctionExtractor) InsertDataToTable(tableName string, data interface{}) error {
    switch v := data.(type) {
    case map[string]interface{}:
        return g.insertSingleRecord(tableName, v)
        
    case []interface{}:
        if len(v) > 0 {
            if _, ok := v[0].(map[string]interface{}); ok {
                // Array of objects
                for _, item := range v {
                    if record, ok := item.(map[string]interface{}); ok {
                        if err := g.insertSingleRecord(tableName, record); err != nil {
                            return err
                        }
                    }
                }
            } else {
                // Array of primitives
                for _, item := range v {
                    query := fmt.Sprintf("INSERT INTO %s (value) VALUES ($1)", tableName)
                    _, err := g.db.Exec(query, fmt.Sprintf("%v", item))
                    if err != nil {
                        return fmt.Errorf("failed to insert primitive value: %w", err)
                    }
                }
            }
        }
        
    default:
        // Single value as JSON
        jsonData, err := json.Marshal(data)
        if err != nil {
            return fmt.Errorf("failed to marshal data to JSON: %w", err)
        }
        
        query := fmt.Sprintf("INSERT INTO %s (data) VALUES ($1)", tableName)
        _, err = g.db.Exec(query, string(jsonData))
        if err != nil {
            return fmt.Errorf("failed to insert JSON data: %w", err)
        }
    }

    g.logger.Printf("Data inserted into table %s", tableName)
    return nil
}

// insertSingleRecord inserts a single record (map) into a table
func (g *GitHubFunctionExtractor) insertSingleRecord(tableName string, record map[string]interface{}) error {
    if len(record) == 0 {
        return nil
    }

    var columns []string
    var placeholders []string
    var values []interface{}

    i := 1
    for key, value := range record {
        columns = append(columns, key)
        placeholders = append(placeholders, "$"+strconv.Itoa(i))
        
        // Convert complex types to JSON strings
        switch v := value.(type) {
        case []interface{}, map[string]interface{}:
            jsonData, err := json.Marshal(v)
            if err != nil {
                return fmt.Errorf("failed to marshal complex type: %w", err)
            }
            values = append(values, string(jsonData))
        default:
            values = append(values, value)
        }
        i++
    }

    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
        tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

    _, err := g.db.Exec(query, values...)
    return err
}

// ProcessRepository is the main method to process a GitHub repository
func (g *GitHubFunctionExtractor) ProcessRepository(repoURL string) (*ProcessingResult, error) {
    result := &ProcessingResult{
        ProcessedFunctions: []FunctionInfo{},
        CreatedTables:      []string{},
        Errors:             []string{},
        ExecutedFunctions:  []string{},
    }

    // Clone repository
    if err := g.CloneRepository(repoURL); err != nil {
        return result, fmt.Errorf("failed to clone repository: %w", err)
    }
    defer g.Cleanup()

    // Connect to database
    if err := g.ConnectToDB(); err != nil {
        return result, fmt.Errorf("failed to connect to database: %w", err)
    }
    defer g.CloseDB()

    // Find Go files
    goFiles, err := g.FindGoFiles()
    if err != nil {
        return result, fmt.Errorf("failed to find Go files: %w", err)
    }

    g.logger.Printf("Found %d Go files", len(goFiles))

    // Process each Go file
    for _, filePath := range goFiles {
        functions, err := g.ExtractFunctionsFromFile(filePath)
        if err != nil {
            result.Errors = append(result.Errors, 
                fmt.Sprintf("Failed to extract functions from %s: %v", filePath, err))
            continue
        }

        // Process each function
        for _, function := range functions {
            result.ProcessedFunctions = append(result.ProcessedFunctions, function)

            // Try to execute function
            data, err := g.ExecuteFunction(function)
            if err != nil {
                result.Errors = append(result.Errors, 
                    fmt.Sprintf("Failed to execute function %s: %v", function.Name, err))
                continue
            }

            if data != nil {
                // Create table and insert data
                if err := g.CreateTableFromData(function.Name, data); err != nil {
                    result.Errors = append(result.Errors, 
                        fmt.Sprintf("Failed to create table for %s: %v", function.Name, err))
                    continue
                }

                if err := g.InsertDataToTable(function.Name, data); err != nil {
                    result.Errors = append(result.Errors, 
                        fmt.Sprintf("Failed to insert data for %s: %v", function.Name, err))
                    continue
                }

                result.CreatedTables = append(result.CreatedTables, function.Name)
                result.ExecutedFunctions = append(result.ExecutedFunctions, function.Name)
            }
        }
    }

    return result, nil
}