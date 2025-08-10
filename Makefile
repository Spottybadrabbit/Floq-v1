.PHONY: build run clean install test fmt vet check

# Application name
APP_NAME=floq-v1

# Build the application
build:
	go build -o $(APP_NAME) .

# Install dependencies
install:
	go mod tidy
	go mod download

# Run the application
run:
	go run .

# Run with config file
run-config:
	CONFIG_FILE=config.json go run .

# Clean build artifacts
clean:
	rm -f $(APP_NAME)
	rm -f $(APP_NAME)-*
	go clean

# Test the application
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all quality checks
check: fmt vet test

# Build for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(APP_NAME).exe .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-mac .

build-all: build-linux build-windows build-mac

# Development helpers
dev-setup:
	go mod tidy
	cp .env.template .env
	cp config.json.template config.json
	@echo "Please edit .env and config.json with your database credentials"

# Database helpers (requires psql)
db-test:
	@echo "Testing database connection..."
	@psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -c "SELECT version();"

# Release build
release: clean check build-all
	@echo "Release build complete"