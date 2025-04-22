# Project settings
APP_NAME = vote-api
MAIN_FILE = cmd/main.go

# Go settings
GO_CMD = go
SWAG_CMD = swag

.PHONY: all build run test clean swag help

all: build

## Build the Go application
build:
	@echo "🔨 Building $(APP_NAME)..."
	$(GO_CMD) build -o $(APP_NAME) $(MAIN_FILE)

## Run the application
run:
	@echo "🚀 Running $(APP_NAME)..."
	$(GO_CMD) run $(MAIN_FILE)

## Run tests
test:
	@echo "🧪 Running tests..."
	$(GO_CMD) test ./... -v

## Generate Swagger docs
swag:
	@echo "📚 Generating Swagger docs..."
	$(SWAG_CMD) init --generalInfo cmd/main.go --output docs

## Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f $(APP_NAME)

## Show help
help:
	@echo "🛠️  Available commands:"
	@echo "  make build     - Build the project"
	@echo "  make run       - Run the application"
	@echo "  make test      - Run unit tests"
	@echo "  make swag      - Generate Swagger documentation"
	@echo "  make clean     - Remove build artifacts"
