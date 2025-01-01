.PHONY: build run test clean

# Build variables
BINARY_NAME=storage
BINARY_DIR=bin
CMD_DIR=cmd/api

build:
	@echo "Building..."
	@go build -o $(BINARY_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

run: build
	@echo "Running..."
	@./$(BINARY_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./internal/... ./pkg/...

clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)

# Development helpers
dev:
	@go run ./$(CMD_DIR)
