.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker build -t renderowl-api:latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 --env-file .env renderowl-api:latest

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run all checks
check: fmt lint test
