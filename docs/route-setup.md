# Nanabush gRPC Server Route Setup

**Date:** 2025-11-15  
**Cluster:** `ocp-sno-1050ti` (1050Ti GPU cluster)  
**Route:** `nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443`

## Route Configuration

The nanabush gRPC server is exposed via OpenShift Route using **edge TLS termination**:

```yaml
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: nanabush-grpc
  namespace: nanabush
spec:
  host: nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org
  to:
    kind: Service
    name: nanabush-grpc-server
  port:
    targetPort: grpc
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
```

## Route Details

- **Host:** `nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org`
- **Port:** `443` (HTTPS via edge termination)
- **TLS Termination:** `edge` (OpenShift Router terminates TLS)
- **Backend Service:** `nanabush-grpc-server` (port 50051)
- **Protocol:** gRPC over HTTP/2

## HAProxy Integration

The Route is configured to be picked up by HAProxy (HAP) for cross-cluster access:
- HAProxy monitors OpenShift Routes
- Routes exposed as `*.apps.ocp-sno-1050ti.rh.dasmlab.org` are automatically discovered
- HAProxy forwards gRPC traffic (HTTP/2) to the OpenShift Router
- OpenShift Router terminates TLS and forwards to backend service

## Glooscap Client Configuration

**Cluster:** `ocp-ai-sno-2` (glooscap cluster)

Update glooscap operator deployment with Route address:

```yaml
env:
  - name: NANABUSH_GRPC_ADDR
    value: "nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443"
  - name: NANABUSH_SECURE
    value: "true"  # Route uses edge TLS termination
```

## Testing the Route

### From Local Machine (if DNS is configured)

```bash
# Test Route connectivity
curl -k https://nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443

# Test with grpcurl (if gRPC over HTTP/2 works)
grpcurl -insecure nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443 list
```

### From Glooscap Cluster (ocp-ai-sno-2)

Once HAProxy is configured and DNS is in place, glooscap operator can connect via the Route.

## Troubleshooting

### Route Not Accessible

1. **Check Route status:**
   ```bash
   CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"
   oc --context=$CONTEXT get route -n nanabush nanabush-grpc
   ```

2. **Check Service:**
   ```bash
   oc --context=$CONTEXT get svc -n nanabush nanabush-grpc-server
   ```

3. **Check Pods:**
   ```bash
   oc --context=$CONTEXT get pods -n nanabush -l app=nanabush-grpc-server
   ```

4. **Check Route endpoints:**
   ```bash
   oc --context=$CONTEXT get endpoints -n nanabush nanabush-grpc-server
   ```

### gRPC Over Route

**Important:** OpenShift Routes support gRPC over HTTP/2, but there may be limitations:
- Some gRPC features may not work through edge termination
- Streaming may have limitations
- If issues occur, consider LoadBalancer/Service Mesh instead

### DNS Resolution

Ensure DNS is configured for `*.apps.ocp-sno-1050ti.rh.dasmlab.org`:
- HAProxy should handle DNS resolution
- For local testing, add to `/etc/hosts` or use `oc get route` to get the router IP

## Next Steps

1. ✅ Route created on 1050ti cluster
2. ⏳ Update glooscap deployment on ocp-ai-sno-2 cluster
3. ⏳ Verify HAProxy picks up the Route
4. ⏳ Test connection from glooscap
5. ⏳ Test gRPC calls (CheckTitle, Translate)

