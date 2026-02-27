.PHONY: build run test clean docker-build docker-run migrate

# Build the application
build:
	go build -o bin/renderowl2-api ./cmd/api/main.go

# Run the application
run:
	go run ./cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t renderowl2-api .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker containers
docker-stop:
	docker-compose down

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run with hot reload (requires air)
dev:
	air
