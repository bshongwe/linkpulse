.PHONY: help build run test clean docker-build docker-run fmt

help:
	@echo "BFF Service Commands"
	@echo "===================="
	@echo "make build        - Build BFF binary"
	@echo "make run          - Run BFF locally"
	@echo "make test         - Run tests"
	@echo "make clean        - Clean build artifacts"
	@echo "make docker-build - Build Docker image"
	@echo "make docker-run   - Run Docker container"

build:
	@echo "Building BFF service..."
	go build -o ./bin/bff ./cmd/api

run: build
	@echo "Running BFF service..."
	LINKPULSE_ENVIRONMENT=development ./bin/bff

test:
	@echo "Running tests..."
	go test -v -cover ./...

clean:
	@echo "Cleaning..."
	rm -rf ./bin
	go clean

docker-build:
	@echo "Building Docker image..."
	docker build -t linkpulse/bff:latest .

docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-e LINKPULSE_ENVIRONMENT=development \
		-e LINKPULSE_SERVER_PORT=8080 \
		linkpulse/bff:latest

fmt:
	@echo "Formatting code..."
	go fmt ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
