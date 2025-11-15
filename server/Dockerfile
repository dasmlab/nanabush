# Multi-stage build for nanabush gRPC server
FROM registry.redhat.io/ubi9/go-toolset:latest AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy proto files (generated)
COPY pkg/proto/ ./pkg/proto/

# Copy source code
COPY cmd/ ./cmd/
COPY pkg/service/ ./pkg/service/

# Build the binary to /tmp (writable location)
RUN go build -o /tmp/nanabush-grpc-server ./cmd/server && \
    chmod +x /tmp/nanabush-grpc-server

# Final stage
FROM registry.redhat.io/ubi9/ubi-minimal:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /tmp/nanabush-grpc-server /app/nanabush-grpc-server

# Set executable permissions
RUN chmod +x /app/nanabush-grpc-server

# Non-root user (UID will be assigned by OpenShift SCC)
# Create a group 0 (root group) user - OpenShift will assign UID
RUN useradd -r -g 0 -u 1001 nanabush || true && \
    chown -R 1001:0 /app && \
    chmod -R g+w /app

# Don't set USER here - let OpenShift SCC handle it
# USER 1001

EXPOSE 50051

CMD ["/app/nanabush-grpc-server", "-port", "50051", "-insecure"]

