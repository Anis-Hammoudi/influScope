# InfluScope Automation

.PHONY: all build up down test clean logs

# Default target
all: test build up

# 🐳 Run the full stack (Detached)
up:
	@echo " Starting InfluScope..."
	docker-compose up --build -d
	@echo "Services running! Search API: http://localhost:8080/search?q=tech"

# 🛑 Stop everything
down:
	@echo "Stopping services..."
	docker-compose down

# 🧪 Run all unit tests
test:
	@echo " Running Unit Tests..."
	cd scraper && go test -v ./...
	cd indexer && go test -v ./...
	cd api && go test -v ./...
	@echo "All tests passed!"

# 🏗️ Build binaries locally (checks for compilation errors without Docker)
build:
	@echo "Building binaries..."
	cd scraper && go build -v ./...
	cd indexer && go build -v ./...
	cd api && go build -v ./...

# 📜 Tail logs
logs:
	docker-compose logs -f