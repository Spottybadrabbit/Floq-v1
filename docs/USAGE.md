# Usage Guide

## Configuration

### Environment Variables

Create a `.env` file from the template:

```bash
cp .env.template .env
```

Edit the `.env` file with your database credentials:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=your_database_name
DB_USER=your_username
DB_PASSWORD=your_password
DB_SSLMODE=disable
```

### JSON Configuration File

Alternatively, use a JSON configuration file:

```bash
cp config.json.template config.json
```

Edit `config.json`:

```json
{
  "host": "localhost",
  "port": "5432",
  "database": "your_database_name",
  "user": "your_username",
  "password": "your_password",
  "sslmode": "disable"
}
```

To use the JSON config, set the environment variable:

```bash
export CONFIG_FILE=config.json
```

## Running the Application

### Basic Usage

```bash
# Using environment variables
./floq-v1 https://github.com/username/repository.git

# Or with go run
go run . https://github.com/username/repository.git
```

### Using Make Commands

```bash
# Install dependencies
make install

# Run with environment variables
make run

# Run with config file
make run-config

# Build the application
make build

# Run tests
make test

# Format and check code
make check
```

## Supported Function Types

The application will process Go functions that meet these criteria:

- ✅ **Exported functions** (start with capital letter)
- ✅ **No parameters required**
- ✅ **Return serializable data**

### Examples of Supported Functions

```go
// Returns a map - creates table with columns for each key
func GetUserData() map[string]interface{} {
    return map[string]interface{}{
        "name": "John Doe",
        "age": 30,
        "active": true,
    }
}

// Returns a slice of structs - creates table with columns for struct fields
func ListProducts() []Product {
    return []Product{
        {Name: "Laptop", Price: 999.99},
        {Name: "Mouse", Price: 29.99},
    }
}

// Returns a simple string - creates table with single data column
func GenerateReport() string {
    return "Monthly sales report: $50,000"
}

// Returns JSON-serializable data
func GetConfiguration() interface{} {
    return map[string]interface{}{
        "version": "1.0.0",
        "features": []string{"auth", "api", "db"},
    }
}
```

### Examples of Unsupported Functions

```go
// ❌ Has parameters
func ProcessUser(id string) User { ... }

// ❌ Not exported (starts with lowercase)
func internalHelper() string { ... }

// ❌ Method (has receiver)
func (u *User) GetName() string { ... }
```

## Database Table Creation

The application automatically creates PostgreSQL tables based on function output:

### Map Output
```go
func GetUser() map[string]interface{} {
    return map[string]interface{}{
        "name": "John",
        "age": 30,
        "active": true,
    }
}
```

Creates table:
```sql
CREATE TABLE GetUser (
    id SERIAL PRIMARY KEY,
    name TEXT,
    age INTEGER,
    active BOOLEAN
);
```

### Array of Maps
```go
func ListUsers() []map[string]interface{} {
    return []map[string]interface{}{
        {"name": "John", "age": 30},
        {"name": "Jane", "age": 25},
    }
}
```

Creates table with multiple rows.

### Simple Values
```go
func GetMessage() string {
    return "Hello World"
}
```

Creates table:
```sql
CREATE TABLE GetMessage (
    id SERIAL PRIMARY KEY,
    data JSONB
);
```

## Error Handling

The application provides detailed error reporting:

- **Repository cloning errors**: Invalid URLs, network issues
- **Function extraction errors**: Syntax errors, parsing issues  
- **Function execution errors**: Runtime panics, import issues
- **Database errors**: Connection issues, table creation failures

## Limitations

1. **Function Parameters**: Only functions with no parameters are supported
2. **Return Types**: Must return JSON-serializable data
3. **Dependencies**: Target repository must have all dependencies available
4. **Go Modules**: Repository must be a valid Go module
5. **Public Repositories**: Currently only supports publicly accessible repositories

## Troubleshooting

### Common Issues

**Database Connection Failed**
```
Failed to connect to database: connection refused
```
- Check database is running
- Verify credentials in config
- Check network connectivity

**Function Execution Failed**
```
Failed to execute function GetData: exit status 2
```
- Function may have dependencies not available
- Check for missing imports
- Function may panic during execution

**Table Creation Failed**
```
Failed to create table for GetData: syntax error
```
- Function name may contain invalid characters
- Output data structure may be incompatible

### Debug Mode

Set log level for more detailed output:
```bash
export LOG_LEVEL=debug
```

## Advanced Usage

### Processing Multiple Repositories

```bash
#!/bin/bash
repos=(
    "https://github.com/user/repo1.git"
    "https://github.com/user/repo2.git"
    "https://github.com/user/repo3.git"
)

for repo in "${repos[@]}"; do
    echo "Processing $repo"
    ./floq-v1 "$repo"
done
```

### Custom Database Schema

The application creates tables in the connected database. To organize tables:

```sql
-- Create a dedicated schema
CREATE SCHEMA extracted_functions;

-- Update connection string to use schema
-- Note: Currently requires manual table moving
```

### Monitoring Progress

Check application logs:
```bash
./floq-v1 repo.git 2>&1 | tee processing.log
```

## Examples

See the `examples/` directory for sample repositories and expected outputs.