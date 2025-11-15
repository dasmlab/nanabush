# gRPC Setup - Next Steps

**Status:** Server implementation complete, proto compilation pending

## Completed ✅

1. **Server Implementation** (`server/`)
   - ✅ `TranslationService` implementation with all three RPC methods
   - ✅ Server main entry point with health checks
   - ✅ Placeholder backend interface for vLLM integration
   - ✅ Graceful shutdown handling
   - ✅ Logging and error handling

2. **Deployment Manifests** (`kustomize/base/`)
   - ✅ Deployment with 2 replicas
   - ✅ Service (ClusterIP, port 50051)
   - ✅ NetworkPolicy (ingress from trusted namespaces)
   - ✅ Health checks (gRPC liveness/readiness probes)

3. **Build System**
   - ✅ Makefile updated for proto generation (client + server)
   - ✅ Server Makefile for building/running
   - ✅ go.mod created for server dependencies

## Pending ⏳

### 1. Proto Compilation

**Required:** `protoc` compiler must be installed

```bash
# Install protoc
sudo apt install protobuf-compiler  # Ubuntu/Debian
brew install protobuf              # macOS

# Verify
protoc --version

# Install Go plugins (already in Makefile)
cd /home/dasm/org-dasmlab/infra/nanabush
make install-protoc

# Generate proto stubs
make proto  # Generates both client and server stubs
```

**After compilation:**
- Client stubs: `../tools/glooscap/operator/pkg/nanabush/proto/v1/`
- Server stubs: `server/pkg/proto/v1/`

### 2. Update Client Code (glooscap)

**Location:** `/home/dasm/org-dasmlab/tools/glooscap/operator/pkg/nanabush/client.go`

**Changes needed:**
1. Uncomment client stub initialization
2. Update `CheckTitle()` to use generated types
3. Update `Translate()` to use generated types
4. Import generated proto package

**Note:** Client runs on `ocp-ai-sno-2`, server runs on `ocp-sno-1050ti`

### 3. Fix Server Imports

After proto compilation, update server imports in:
- `server/pkg/service/translation_service.go`
- `server/cmd/server/main.go`

Ensure imports point to: `github.com/dasmlab/nanabush/server/pkg/proto/v1`

### 4. Build and Deploy

**Server (nanabush on ocp-sno-1050ti):**

```bash
# Context for 1050ti cluster
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Build server binary
cd /home/dasm/org-dasmlab/infra/nanabush/server
make deps
make build

# Deploy to cluster
cd /home/dasm/org-dasmlab/infra/nanabush
oc --context=$CONTEXT apply -k kustomize/base
```

**Client (glooscap on ocp-ai-sno-2):**

```bash
# Context for ocp-ai-sno-2 cluster
# TODO: Get correct context name for ocp-ai-sno-2

# Build operator with updated client
cd /home/dasm/org-dasmlab/tools/glooscap/operator
make build

# Deploy operator
# TODO: Update deployment to use new client
```

### 5. Test Basic Communication

**Before TLS/mTLS** (using insecure mode):

1. **Deploy server** on 1050ti cluster
2. **Deploy client** on ocp-ai-sno-2 cluster
3. **Test connection:**
   ```bash
   # From ocp-ai-sno-2 cluster (glooscap namespace)
   # Port-forward or use service mesh to reach 1050ti cluster
   # Address: nanabush-grpc-server.nanabush.svc.cluster.local:50051
   ```
4. **Verify:**
   - Server health checks pass
   - Client can call `CheckTitle()`
   - Client can call `Translate()`

## Important Notes

### Cluster Separation

- **glooscap** runs on: `ocp-ai-sno-2`
- **nanabush** runs on: `ocp-sno-1050ti`
- **Communication:** Cross-cluster (requires service mesh or network routing)

### Network Connectivity

**Options:**
1. **Service Mesh** - If both clusters are part of a service mesh federation
2. **External Service** - Expose server via LoadBalancer/Ingress
3. **Direct Network** - If clusters can reach each other via cluster DNS

**Current NetworkPolicy:**
- Server accepts ingress from namespaces with label: `glooscap.dasmlab.org/trusted: "true"`
- Client namespace on ocp-ai-sno-2 must have this label

### Insecure Mode

- Currently using insecure gRPC (no TLS)
- **TODO:** Implement TLS/mTLS before production
- Server flag: `-insecure=true` (default)

## After Basic Setup Works

1. **Implement vLLM Backend Integration**
   - Connect to RHOAI/KServe InferenceService
   - Implement `TranslatorBackend` interface
   - Handle model serving requests

2. **Add TLS/mTLS**
   - Generate certificates
   - Configure server TLS
   - Configure client mTLS
   - Update deployment secrets

3. **Add Observability**
   - OTEL tracing
   - Prometheus metrics
   - Structured logging

4. **Add Rate Limiting**
   - Per-client limits
   - Global rate limits
   - Service mesh integration

