.PHONY: build run docker-build docker-up docker-down clean test deps

# Build the application
build:
	go build -o main .

# Run the application
run:
	go run main.go

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Build Docker image
docker-build:
	docker-compose build

# Start Docker containers
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -f main
	go clean

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# View logs
logs:
	docker-compose logs -f api

