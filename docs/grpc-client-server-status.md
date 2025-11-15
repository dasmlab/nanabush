# gRPC Client-Server Interface Status

**Date:** 2025-11-15  
**Purpose:** Summary of current state of gRPC communication between glooscap (client) and nanabush (server)

## Current State

### ✅ What Exists

#### 1. **Protocol Definition (proto/translation.proto)**
- **Location:** `/home/dasm/org-dasmlab/infra/nanabush/proto/translation.proto`
- **Package:** `nanabush.v1`
- **Service:** `TranslationService` with three RPC methods:
  - `CheckTitle(TitleCheckRequest) returns (TitleCheckResponse)` - Pre-flight validation
  - `Translate(TranslateRequest) returns (TranslateResponse)` - Full document translation
  - `TranslateStream(stream TranslateChunk) returns (stream TranslateChunk)` - Streaming for large docs

**Status:** ✅ **Defined but not compiled to Go stubs**

#### 2. **Client Implementation (glooscap side)**
- **Location:** `/home/dasm/org-dasmlab/tools/glooscap/operator/pkg/nanabush/client.go`
- **Structure:**
  - `Client` struct with `conn *grpc.ClientConn`, `addr string`, `secure bool`
  - `Config` struct with TLS fields (commented as TODO):
    - `TLSCertPath` - Client certificate path
    - `TLSKeyPath` - Client private key path
    - `TLSCAPath` - CA certificate for server verification
  - Methods:
    - `NewClient(cfg Config)` - Creates client (currently uses insecure credentials)
    - `CheckTitle()` - **Placeholder** (returns mock response)
    - `Translate()` - **Placeholder** (returns error indicating proto not compiled)
    - `Close()` - Connection cleanup

**Status:** ⚠️ **Skeleton exists, TLS/mTLS not implemented, proto not compiled**

#### 3. **Client Integration (glooscap operator)**
- **Location:** `/home/dasm/org-dasmlab/tools/glooscap/operator/cmd/main.go`
- **Configuration:**
  - Environment variable: `NANABUSH_GRPC_ADDR` (e.g., `nanabush-service.nanabush.svc:50051`)
  - Environment variable: `NANABUSH_SECURE` (boolean, default false)
  - Timeout: 30 seconds
- **Usage:**
  - Initialized in `TranslationJobReconciler` struct
  - Used in `translationjob_controller.go` for:
    - Pre-flight checks via `CheckTitle()`
    - Full translation via `Translate()`

**Status:** ⚠️ **Wired but not functional (proto not compiled, TLS not configured)**

### ❌ What's Missing

#### 1. **gRPC Server Implementation**
- **Status:** ❌ **Does not exist**
- **Required:**
  - Go server implementation of `TranslationService`
  - Server startup code (main.go)
  - TLS/mTLS server configuration
  - Integration with vLLM backend (RHOAI/KServe)

#### 2. **Proto Compilation**
- **Status:** ❌ **Not compiled**
- **Required:**
  - Compile `proto/translation.proto` to Go stubs
  - Generate `nanabushv1` package
  - Update client code to use generated stubs

#### 3. **TLS/mTLS Implementation**
- **Status:** ❌ **Not implemented (TODOs present)**
- **Client Side (glooscap):**
  - Load client certificate from `TLSCertPath`
  - Load client private key from `TLSKeyPath`
  - Load CA certificate from `TLSCAPath`
  - Configure `grpc.WithTransportCredentials()` with TLS credentials
  - Currently hardcoded to `insecure.NewCredentials()`

- **Server Side (nanabush):**
  - Load server certificate and key
  - Load CA certificate for client verification (mTLS)
  - Configure gRPC server with TLS credentials
  - Require client certificate authentication

#### 4. **Certificate Management**
- **Status:** ❌ **Not implemented**
- **Required:**
  - Kubernetes Secrets for certificates
  - Certificate rotation strategy
  - Integration with cert-manager or OpenShift cert operator
  - Service Mesh certificates (if using Istio)

#### 5. **Service Deployment**
- **Status:** ❌ **No gRPC service defined**
- **Required:**
  - Kubernetes Service (ClusterIP) exposing gRPC port
  - Deployment manifest for gRPC server
  - Service Mesh configuration (if using Istio for mTLS)
  - Health checks and readiness probes

#### 6. **Security Primitives**
- **Status:** ⚠️ **Partially defined**
- **What exists:**
  - NetworkPolicy: Egress-only, namespace-based filtering
  - SCC: Security context constraints (seccomp, read-only rootfs)
  - Kata runtime class mentioned in values.yaml
  
- **What's missing:**
  - Client authentication/authorization (beyond mTLS certs)
  - Rate limiting
  - Request validation
  - Audit logging for gRPC calls

## Security Requirements (from documentation)

### Mentioned in README and docs:
1. **Service Mesh (Istio) for mTLS** - ✅ Installed (RHOAI prerequisites)
   - Service Mesh Operator v2.6.11 installed
   - Istio system namespace created
   - PeerAuthentication policies needed

2. **Zero-Trust Network Policies** - ✅ Defined
   - NetworkPolicy restricts egress to trusted namespaces
   - Namespace label: `glooscap.dasmlab.org/trusted: "true"`

3. **Kata Runtime** - ⚠️ Mentioned but not verified
   - Values.yaml references `runtimeClassName: kata`
   - Requires Kata Containers runtime installed

4. **OTEL Tracing** - ✅ Configured
   - Client-side tracing mentioned
   - Server-side tracing needed

5. **Audit Logging** - ❌ Not implemented
   - Mentioned in objectives
   - No implementation yet

## Architecture Intent (from docs)

### Service-Oriented Model:
- **Always-on vLLM deployment** exposes gRPC/REST endpoint
- **Operator uses service mesh** for mTLS, rate limiting, zero-trust
- **Hardened egress guardrails** (see security doc)

### Primitives (from translation.proto):
1. **Title-Only Translation** (`PRIMITIVE_TITLE`)
   - Lightweight pre-flight check
   - Fast validation

2. **Full Document Translation** (`PRIMITIVE_DOC_TRANSLATE`)
   - Complete document processing
   - Includes metadata (template helper, wiki URI, page info)

3. **Streaming Translation** (`TranslateStream`)
   - For large documents
   - Chunked processing

## Next Steps Required

### Priority 1: Foundation
1. **Compile proto file** to Go stubs
   - Add Makefile target or build script
   - Generate `nanabushv1` package
   - Update client to use generated types

2. **Implement gRPC server** in nanabush
   - Create server package/struct
   - Implement `TranslationService` interface
   - Add server startup code
   - Integrate with vLLM backend (RHOAI/KServe)

3. **Basic TLS implementation**
   - Certificate generation/management
   - Server TLS configuration
   - Client TLS configuration (beyond insecure)

### Priority 2: Security
4. **Mutual TLS (mTLS)**
   - Server requires client certificates
   - Client presents certificates
   - CA verification on both sides

5. **Service Mesh Integration**
   - PeerAuthentication policy (STRICT mode)
   - Service Mesh mTLS (if using Istio)
   - Alternative: Application-level mTLS

6. **Certificate Management**
   - Kubernetes Secrets for certs
   - Cert-manager or OpenShift cert operator
   - Rotation strategy

### Priority 3: Production Hardening
7. **Authentication/Authorization**
   - Beyond mTLS certs (optional)
   - Service account tokens
   - RBAC integration

8. **Rate Limiting**
   - Per-client limits
   - Global rate limits
   - Service Mesh rate limiting

9. **Audit Logging**
   - Log all gRPC calls
   - Include request metadata
   - Sanitize sensitive data

10. **Health Checks**
    - gRPC health service
    - Kubernetes liveness/readiness probes
    - Service mesh health checks

## Current Configuration Examples

### Client Configuration (glooscap):
```go
client, err := nanabush.NewClient(nanabush.Config{
    Address: "nanabush-service.nanabush.svc:50051",
    Secure:  true,  // Currently ignored (uses insecure)
    Timeout: 30 * time.Second,
})
```

### Environment Variables:
```bash
NANABUSH_GRPC_ADDR=nanabush-service.nanabush.svc:50051
NANABUSH_SECURE=true  # Currently not implemented
```

### Network Policy (existing):
```yaml
# Allows egress to trusted namespaces only
# Port: 8080 (HTTP) - needs gRPC port (50051)
```

## Integration Points

### With RHOAI/KServe:
- gRPC server needs to:
  - Submit inference requests to KServe InferenceService
  - Handle model serving endpoints
  - Process vLLM responses

### With Service Mesh:
- Istio PeerAuthentication (STRICT mode)
- Service Mesh mTLS certificates
- Traffic policies and rate limiting

### With OTEL:
- Trace context propagation (`traceparent`)
- Distributed tracing across client-server
- Metrics collection

## Summary

**Current State:** 🟡 **Early Development - Foundation Laid**
- ✅ Protocol defined
- ✅ Client skeleton exists
- ⚠️ Proto not compiled
- ⚠️ TLS/mTLS not implemented
- ❌ Server does not exist
- ❌ Certificate management not set up
- ⚠️ Security primitives partially defined

**Blockers:**
1. Proto compilation needed to proceed
2. Server implementation required
3. TLS/mTLS implementation critical for production

**Ready for:**
- Proto compilation and stub generation
- Server implementation
- Basic TLS setup (can evolve to mTLS)
- Certificate management integration

