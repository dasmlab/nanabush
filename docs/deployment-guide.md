# Nanabush gRPC Server Deployment Guide

**Date:** 2025-11-15  
**Cluster:** `ocp-sno-1050ti` (1050Ti GPU cluster)  
**Purpose:** Deploy nanabush gRPC server to handle translation requests from glooscap

## Prerequisites

- ✅ OpenShift cluster `ocp-sno-1050ti` accessible
- ✅ RHOAI platform installed and operational
- ✅ GPU Operator installed and working
- ✅ LVMS storage available
- ✅ Container registry accessible (for pushing images)

## Build and Deploy Steps

### 1. Build Container Image

```bash
cd /home/dasm/org-dasmlab/infra/nanabush

# Build the gRPC server image
./scripts/build-grpc-server.sh

# Or manually:
docker build -f kustomize/base/Dockerfile.grpc-server -t nanabush-grpc-server:latest .
```

**Note:** The image will be built from source. For production, push to a registry:

```bash
# Tag for registry
docker tag nanabush-grpc-server:latest registry.example.com/rhoai/nanabush-grpc-server:latest

# Push (requires authentication)
docker push registry.example.com/rhoai/nanabush-grpc-server:latest
```

### 2. Update Deployment Manifest

Update `kustomize/base/grpc-server-deployment.yaml` with the correct image:

```yaml
image: registry.example.com/rhoai/nanabush-grpc-server:latest
```

Or for local development/testing:

```bash
# Load image to cluster (if using local cluster)
# For remote cluster, push to accessible registry
```

### 3. Deploy to 1050ti Cluster

```bash
# Set context for 1050ti cluster
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Apply all nanabush resources
oc --context=$CONTEXT apply -k kustomize/base

# Verify deployment
oc --context=$CONTEXT get pods -n nanabush -l app=nanabush-grpc-server
oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
```

### 4. Verify Server is Running

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Check pod status
oc --context=$CONTEXT get pods -n nanabush -l app=nanabush-grpc-server

# Check logs
oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server --tail=50

# Port-forward for testing (optional)
oc --context=$CONTEXT port-forward -n nanabush svc/nanabush-grpc-server 50051:50051
```

### 5. Test gRPC Health Check

```bash
# Using grpc_health_probe (if available)
grpc_health_probe -addr localhost:50051

# Or using grpcurl
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 nanabush.v1.TranslationService.CheckTitle
```

## Cross-Cluster Communication

### Network Requirements

**glooscap** (ocp-ai-sno-2) needs to reach **nanabush** (ocp-sno-1050ti).

**Options:**
1. **Service Mesh Federation** - If both clusters are in a service mesh federation
2. **External Service** - Expose server via LoadBalancer/Ingress with external IP
3. **Direct Network** - If clusters can reach each other via DNS/routing

### Option 1: Service Mesh (Recommended for Production)

If both clusters are part of a service mesh federation:
- Service Mesh will handle cross-cluster mTLS
- Use service mesh DNS: `nanabush-grpc-server.nanabush.svc.cluster.local:50051`
- Configure ServiceEntry on glooscap cluster to reach nanabush cluster

### Option 2: LoadBalancer/Ingress (Development/Testing)

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Expose via LoadBalancer (requires external IP)
oc --context=$CONTEXT expose svc nanabush-grpc-server -n nanabush --type=LoadBalancer --port=50051 --target-port=50051

# Get external IP
oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
```

Then configure glooscap client with the external IP:port.

### Option 3: Route/Ingress (HTTP/2 gRPC)

For HTTP/2 gRPC over Route (if supported):

```yaml
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: nanabush-grpc-server
  namespace: nanabush
spec:
  to:
    kind: Service
    name: nanabush-grpc-server
  port:
    targetPort: grpc
  tls:
    termination: passthrough  # For gRPC
```

## Update Glooscap Deployment

On **ocp-ai-sno-2** cluster, update glooscap operator deployment:

### Environment Variables

```yaml
env:
  - name: NANABUSH_GRPC_ADDR
    value: "nanabush-grpc-server.nanabush.svc.cluster.local:50051"
    # Or use external IP if using LoadBalancer:
    # value: "10.20.30.40:50051"
  - name: NANABUSH_SECURE
    value: "false"  # Set to "true" once TLS/mTLS is configured
```

### Namespace Labels

Ensure the glooscap namespace is labeled as trusted (for NetworkPolicy):

```bash
# On ocp-sno-1050ti cluster (nanabush side)
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Label glooscap namespace on 1050ti (if it exists) OR
# Create a namespace selector that allows glooscap from ocp-ai-sno-2

# Actually, since they're on different clusters, NetworkPolicy won't work across clusters
# We'll need Service Mesh or external exposure
```

## Network Policy Limitations

**Important:** Kubernetes NetworkPolicy only works within a single cluster. Since glooscap is on `ocp-ai-sno-2` and nanabush is on `ocp-sno-1050ti`, NetworkPolicy will not enforce cross-cluster traffic.

**Solutions:**
1. **Service Mesh** - Use service mesh policies for cross-cluster security
2. **External Firewall** - Use network-level firewalls/routers
3. **API Gateway** - Place an API gateway in front of nanabush

## Testing Cross-Cluster Communication

### From Glooscap Operator (ocp-ai-sno-2)

```bash
# Port-forward to test (if clusters can reach each other)
oc --context=<glooscap-context> port-forward -n glooscap deployment/glooscap-operator 8080:8080

# Create a test TranslationJob
# The operator should connect to nanabush and call CheckTitle() or Translate()
```

### Verify in Logs

**On nanabush server (1050ti):**
```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"
oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server -f
```

**On glooscap operator (ocp-ai-sno-2):**
```bash
# Get glooscap operator context (TBD)
oc --context=<glooscap-context> logs -n glooscap deployment/glooscap-operator -f
```

## Troubleshooting

### Server Not Starting

```bash
# Check pod events
oc --context=$CONTEXT describe pod -n nanabush -l app=nanabush-grpc-server

# Check logs
oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server
```

### Client Cannot Connect

1. **Verify server is accessible:**
   ```bash
   oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
   ```

2. **Check network connectivity:**
   - From glooscap cluster, test DNS resolution
   - Test TCP connection to server IP:port
   - Check firewall rules

3. **Verify service address:**
   - Ensure `NANABUSH_GRPC_ADDR` is correct
   - For cross-cluster, use external IP or service mesh DNS

### Port Already in Use

```bash
# Check what's using port 50051
oc --context=$CONTEXT get svc -n nanabush | grep 50051
```

## Next Steps After Deployment

1. ✅ Deploy server to 1050ti cluster
2. ✅ Update glooscap with server address
3. ✅ Test basic connection (CheckTitle)
4. ✅ Test translation (Translate)
5. ⏳ Implement vLLM backend integration
6. ⏳ Add TLS/mTLS
7. ⏳ Add observability (metrics, tracing)

