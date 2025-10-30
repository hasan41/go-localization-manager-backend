.PHONY: help build run test docker-up docker-down clean install

help:
	@echo "Available commands:"
	@echo "  make install     - Install Go dependencies"
	@echo "  make build       - Build the server binary"
	@echo "  make run         - Run the server in development mode"
	@echo "  make test        - Run the test script"
	@echo "  make docker-up   - Start Redis using Docker Compose"
	@echo "  make docker-down - Stop Redis"
	@echo "  make clean       - Remove binary and cache files"

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build:
	@echo "Building server..."
	go build -o localization-server .

run:
	@echo "Starting server..."
	go run main.go

test:
	@echo "Running tests..."
	./test.sh

docker-up:
	@echo "Starting Redis..."
	docker-compose up -d
	@echo "Waiting for Redis to be ready..."
	@sleep 2
	@echo "Redis is ready at localhost:6379"

docker-down:
	@echo "Stopping Redis..."
	docker-compose down

clean:
	@echo "Cleaning up..."
	rm -f localization-server
	@echo "Done!"

all: install build

