.PHONY: build run proto deps test clean

# Build the gRPC server
build:
	@echo "Building nanabush gRPC server..."
	@go build -o bin/nanabush-grpc-server ./cmd/server
	@echo "Build complete: bin/nanabush-grpc-server"

# Run the server locally (for development)
run: build
	@echo "Running nanabush gRPC server..."
	@./bin/nanabush-grpc-server -port 50051 -insecure

# Generate proto stubs (must be run from project root: make proto-server)
proto:
	@echo "Note: Run 'make proto-server' from project root to generate proto stubs"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf pkg/proto/v1/*.pb.go pkg/proto/v1/*_grpc.pb.go

