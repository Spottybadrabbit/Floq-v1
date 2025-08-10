# Floq-v1

A Go application that clones GitHub repositories, extracts Go functions, executes them, and populates PostgreSQL tables with their outputs.

## 🚀 Features

- **Automatic Repository Cloning**: Clone GitHub repositories with a single URL
- **Go Function Extraction**: Parse Go source code and extract function metadata using AST
- **Dynamic Function Execution**: Execute exported functions and capture their outputs
- **Database Integration**: Automatically create PostgreSQL tables based on function outputs
- **Batch Processing**: Process multiple repositories in sequence
- **Flexible Configuration**: Support for environment variables and JSON config files
- **Comprehensive Logging**: Detailed logging for debugging and monitoring
- **Error Handling**: Robust error handling with detailed error reporting

## 📋 Prerequisites

- Go 1.21 or later
- PostgreSQL database
- Git (for cloning repositories)

## 🛠️ Installation

1. Clone this repository:
```bash
git clone https://github.com/Spottybadrabbit/Floq-v1.git
cd Floq-v1
```

2. Install dependencies:
```bash
make install
```

3. Set up your database configuration:
```bash
cp .env.template .env
# Edit .env with your database credentials
```

## 🚀 Quick Start

### Using Environment Variables
```bash
export DB_HOST=localhost
export DB_NAME=mydb
export DB_USER=myuser
export DB_PASSWORD=mypass
make run
```

### Using Configuration File
```bash
cp config.json.template config.json
# Edit config.json with your settings
make run-config
```

## 📖 Usage

See [docs/USAGE.md](docs/USAGE.md) for detailed usage instructions.

## 🏗️ How It Works

1. **Repository Cloning**: Clones the specified GitHub repository to a temporary directory
2. **Function Extraction**: Parses Go source files using Go's AST package
3. **Function Execution**: Creates temporary main.go files to execute functions
4. **Table Creation**: Analyzes outputs and creates PostgreSQL tables with appropriate schemas
5. **Data Insertion**: Inserts function outputs into corresponding tables

## 🔧 Configuration

### Environment Variables
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_NAME`: Database name
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_SSLMODE`: SSL mode (default: disable)

### Supported Function Types
Functions must be:
- Exported (start with capital letter)
- Have no parameters
- Return serializable data

Example:
```go
func GetUserData() map[string]interface{} { ... }
func ListProducts() []Product { ... }
func GenerateReport() string { ... }
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⚠️ Limitations

- Only processes exported Go functions with no parameters
- Functions must return serializable data
- Requires Go toolchain for function execution
- Limited to publicly accessible repositories

## 📞 Support

If you encounter any issues or have questions, please open an issue on GitHub.