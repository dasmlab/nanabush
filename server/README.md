# Nanabush gRPC Server

gRPC server implementation for the Nanabush translation service.

## Prerequisites

1. **Go 1.21+** installed
2. **protoc compiler** installed (see below)
3. **protoc plugins:**
   - `protoc-gen-go`
   - `protoc-gen-go-grpc`

## Setup

### Install protoc

```bash
# Ubuntu/Debian
sudo apt install protobuf-compiler

# macOS
brew install protobuf

# Verify installation
protoc --version
```

### Install protoc plugins

From the project root (`/home/dasm/org-dasmlab/infra/nanabush`):

```bash
make install-protoc
```

Or manually:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generate proto stubs

From the project root:

```bash
# Generate for both client and server
make proto

# Or generate separately:
make proto-server   # For server (nanabush)
make proto-client   # For client (glooscap)
```

This will generate:
- `server/pkg/proto/v1/*.pb.go` - Protocol buffer types
- `server/pkg/proto/v1/*_grpc.pb.go` - gRPC service stubs

### Build

```bash
cd server
make deps   # Download dependencies
make build  # Build binary
```

Binary will be at: `server/bin/nanabush-grpc-server`

### Run locally (development)

```bash
cd server
make run
```

Server will start on port `50051` in insecure mode.

## Configuration

### Environment Variables

- `NANABUSH_BACKEND_URL` - vLLM backend URL (default: `http://vllm.nanabush.svc:8000`)

### Command-line Flags

- `-port` - gRPC server port (default: `50051`)
- `-insecure` - Run in insecure mode, no TLS (default: `true`)
- `-tls-cert` - Path to TLS server certificate (future)
- `-tls-key` - Path to TLS server private key (future)
- `-tls-ca` - Path to CA certificate for client verification/mTLS (future)

## Deployment

### Kubernetes Deployment

The server is deployed via Kustomize:

```bash
# From project root
kubectl apply -k kustomize/base
```

Or for a specific overlay:

```bash
kubectl apply -k kustomize/overlays/dev
```

### Service

The deployment creates a Kubernetes Service:

- **Name:** `nanabush-grpc-server`
- **Type:** ClusterIP
- **Port:** `50051`
- **Target:** gRPC server pods

### Network Policy

Ingress is restricted to:
- Pods from trusted namespaces (label: `glooscap.dasmlab.org/trusted: "true"`)
- Pods within the `nanabush` namespace (for health checks)

## Service Methods

### CheckTitle

Pre-flight check with title only:

```go
req := &nanabushv1.TitleCheckRequest{
    Title:          "Hello World",
    LanguageTag:    "fr-CA",
    SourceLanguage: "EN",
}
resp, err := client.CheckTitle(ctx, req)
```

### Translate

Full document translation:

```go
req := &nanabushv1.TranslateRequest{
    JobId:          "job-123",
    Primitive:      nanabushv1.PrimitiveType_PRIMITIVE_DOC_TRANSLATE,
    SourceLanguage: "EN",
    TargetLanguage: "fr-CA",
    Source: &nanabushv1.TranslateRequest_Doc{
        Doc: &nanabushv1.DocumentContent{
            Title:    "Example",
            Markdown: "# Example\n\nContent...",
        },
    },
}
resp, err := client.Translate(ctx, req)
```

### TranslateStream

Streaming translation for large documents:

```go
stream, err := client.TranslateStream(ctx)
// Send chunks...
// Receive translated chunks...
```

## Health Checks

The server implements the gRPC health checking protocol:

```bash
grpc_health_probe -addr localhost:50051
```

Kubernetes liveness/readiness probes use gRPC health checks.

## Next Steps

1. **vLLM Backend Integration** - Implement `TranslatorBackend` interface to connect to RHOAI/KServe
2. **TLS/mTLS** - Add certificate-based authentication
3. **Metrics** - Add Prometheus metrics
4. **Tracing** - Integrate with OTEL
5. **Rate Limiting** - Add per-client rate limits

## Notes

- Server currently runs in insecure mode (no TLS)
- Backend integration is placeholder (returns mock responses)
- Proto compilation must happen before building

