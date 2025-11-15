# Cross-Cluster gRPC Communication Setup

**Date:** 2025-11-15  
**Purpose:** Configure communication between glooscap (ocp-ai-sno-2) and nanabush (ocp-sno-1050ti)

## Architecture

```
┌────────────────────────────────────────┐
│  ocp-ai-sno-2 (glooscap cluster)       │
│  ┌──────────────────────────────────┐  │
│  │  glooscap-operator                │  │
│  │  - TranslationJob controller      │  │
│  │  - nanabush.Client                │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
                   │
                   │ gRPC (port 50051)
                   │ (insecure for now)
                   │
                   ▼
┌────────────────────────────────────────┐
│  ocp-sno-1050ti (nanabush cluster)     │
│  ┌──────────────────────────────────┐  │
│  │  nanabush-grpc-server             │  │
│  │  - TranslationService             │  │
│  │  - gRPC endpoint (50051)          │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
```

## Network Connectivity Options

Since glooscap and nanabush are on **different clusters**, we need to establish network connectivity:

### Option 1: Service Mesh Federation (Recommended for Production)

If both clusters are part of a service mesh federation:
- Use service mesh DNS: `nanabush-grpc-server.nanabush.svc.cluster.local:50051`
- Service mesh handles mTLS automatically
- Configure ServiceEntry on glooscap cluster

**Requirements:**
- Both clusters in same service mesh federation
- Service Mesh operator installed on both clusters

### Option 2: LoadBalancer Service (Development/Testing)

Expose nanabush server via LoadBalancer to get external IP:

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Change service type to LoadBalancer
oc --context=$CONTEXT patch svc -n nanabush nanabush-grpc-server -p '{"spec":{"type":"LoadBalancer"}}'

# Get external IP
oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server

# Update glooscap env var with external IP:port
```

**Limitation:** Requires LoadBalancer support in cluster

### Option 3: NodePort + External IP (Alternative)

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Change to NodePort
oc --context=$CONTEXT patch svc -n nanabush nanabush-grpc-server -p '{"spec":{"type":"NodePort"}}'

# Get node IP and port
oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
oc --context=$CONTEXT get nodes -o wide
```

### Option 4: Route/Ingress (if gRPC over HTTP/2 supported)

Some OpenShift versions support gRPC over Route:

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
    termination: passthrough  # Required for gRPC
```

**Limitation:** May not work for all gRPC implementations

## Deployment Steps

### 1. Build and Push nanabush gRPC Server Image

```bash
cd /home/dasm/org-dasmlab/infra/nanabush

# Build image
./scripts/build-grpc-server.sh

# Or manually:
docker build -f kustomize/base/Dockerfile.grpc-server -t nanabush-grpc-server:latest .

# Push to registry accessible by both clusters
docker tag nanabush-grpc-server:latest registry.example.com/rhoai/nanabush-grpc-server:latest
docker push registry.example.com/rhoai/nanabush-grpc-server:latest
```

### 2. Deploy nanabush Server (1050ti cluster)

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Update deployment with correct image
oc --context=$CONTEXT apply -k kustomize/base

# Or apply just the server deployment:
oc --context=$CONTEXT apply -f kustomize/base/grpc-server-deployment.yaml
oc --context=$CONTEXT apply -f kustomize/base/grpc-server-networkpolicy.yaml

# Verify deployment
oc --context=$CONTEXT get pods -n nanabush -l app=nanabush-grpc-server
oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
```

### 3. Expose Server for Cross-Cluster Access

**Option A: LoadBalancer (if available)**

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Change service to LoadBalancer
oc --context=$CONTEXT patch svc -n nanabush nanabush-grpc-server -p '{"spec":{"type":"LoadBalancer"}}'

# Wait for external IP
oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server -w

# Get external IP
EXTERNAL_IP=$(oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo "External IP: ${EXTERNAL_IP}:50051"
```

**Option B: Service Mesh (if configured)**

If service mesh federation is set up:
- DNS: `nanabush-grpc-server.nanabush.svc.cluster.local:50051`
- Service mesh handles routing automatically

### 4. Update Glooscap Deployment (ocp-ai-sno-2)

**Get the context for ocp-ai-sno-2 cluster:**
```bash
# TODO: Get actual context name for ocp-ai-sno-2
# oc config get-contexts | grep ocp-ai-sno-2
```

**Update environment variables:**

```bash
# Set context for glooscap cluster
# GLOOSCAP_CONTEXT="<ocp-ai-sno-2-context>"

# Update deployment with nanabush address
oc --context=$GLOOSCAP_CONTEXT patch deployment -n glooscap controller-manager -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "manager",
          "env": [
            {"name": "NANABUSH_GRPC_ADDR", "value": "<external-ip>:50051"},
            {"name": "NANABUSH_SECURE", "value": "false"}
          ]
        }]
      }
    }
  }
}'
```

Or manually edit `operator/config/manager/manager.yaml` and redeploy.

### 5. Test Connection

**From glooscap operator logs:**
```bash
# On ocp-ai-sno-2 cluster
oc --context=$GLOOSCAP_CONTEXT logs -n glooscap deployment/controller-manager -f

# Create a test TranslationJob to trigger connection
# Watch for logs showing gRPC calls to nanabush
```

**From nanabush server logs:**
```bash
# On ocp-sno-1050ti cluster
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"
oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server -f

# Should see incoming gRPC requests from glooscap
```

## Testing Without Deploying

### Test Locally (Port Forward)

**1. Start nanabush server locally:**
```bash
cd /home/dasm/org-dasmlab/infra/nanabush/server
./bin/nanabush-grpc-server -port 50051 -insecure
```

**2. Port-forward from glooscap operator:**
```bash
# On ocp-ai-sno-2 cluster
oc --context=$GLOOSCAP_CONTEXT port-forward -n glooscap deployment/controller-manager 50051:50051
```

**3. Update glooscap env to point to localhost:**
```bash
oc --context=$GLOOSCAP_CONTEXT set env deployment/controller-manager -n glooscap NANABUSH_GRPC_ADDR=localhost:50051
```

### Test with grpcurl

```bash
# Test CheckTitle
grpcurl -plaintext localhost:50051 \
  nanabush.v1.TranslationService.CheckTitle \
  -d '{"title":"Test","language_tag":"fr-CA","source_language":"EN"}'

# Test Translate
grpcurl -plaintext localhost:50051 \
  nanabush.v1.TranslationService.Translate \
  -d '{
    "job_id":"test-123",
    "primitive":"PRIMITIVE_TITLE",
    "source":{"title":"Hello World"},
    "source_language":"EN",
    "target_language":"fr-CA"
  }'
```

## Troubleshooting

### Server Not Accessible

1. **Check service:**
   ```bash
   CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"
   oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
   ```

2. **Check pods:**
   ```bash
   oc --context=$CONTEXT get pods -n nanabush -l app=nanabush-grpc-server
   oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server
   ```

3. **Test connectivity from glooscap cluster:**
   ```bash
   # From a pod on ocp-ai-sno-2
   oc --context=$GLOOSCAP_CONTEXT run test-connectivity --image=registry.redhat.io/ubi9/ubi:latest --rm -it --restart=Never -- \
     bash -c "nc -zv <nanabush-ip> 50051"
   ```

### Client Cannot Connect

1. **Verify address:**
   - Check `NANABUSH_GRPC_ADDR` environment variable
   - Ensure it's reachable from glooscap cluster
   - Test DNS resolution if using DNS name

2. **Check firewall/routing:**
   - Ensure port 50051 is open between clusters
   - Check network policies (though they won't work cross-cluster)
   - Verify routing tables

3. **Check logs:**
   ```bash
   # Glooscap operator logs
   oc --context=$GLOOSCAP_CONTEXT logs -n glooscap deployment/controller-manager | grep nanabush
   
   # Nanabush server logs
   oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server | grep -i error
   ```

## Security Considerations

### Current Setup (Insecure)

- Using insecure gRPC (no TLS)
- **DO NOT** use in production
- Only for development/testing

### Future: TLS/mTLS

1. Generate certificates
2. Configure server TLS
3. Configure client mTLS
4. Update deployments with secrets

## Next Steps

1. ✅ Build container image
2. ✅ Deploy to 1050ti cluster
3. ✅ Expose service (LoadBalancer/Service Mesh)
4. ✅ Update glooscap with server address
5. ✅ Test connection
6. ⏳ Implement vLLM backend integration
7. ⏳ Add TLS/mTLS
8. ⏳ Add observability

