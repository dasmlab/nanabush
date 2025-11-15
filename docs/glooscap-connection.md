# Glooscap Connection to Nanabush

**Date:** 2025-11-15  
**Purpose:** Connect glooscap (ocp-ai-sno-2) to nanabush (ocp-sno-1050ti)

## Current Status

✅ **Nanabush Server (1050ti cluster):**
- Pod: `nanabush-grpc-server-866f798c65-p9gvh` - Running
- Service: `nanabush-grpc-server` - ClusterIP 172.30.246.228:50051
- Route: `nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443` - Edge TLS
- Server: Listening on port 50051

✅ **Glooscap Client Configuration:**
- Updated: `/home/dasm/org-dasmlab/tools/glooscap/operator/config/manager/manager.yaml`
- `NANABUSH_GRPC_ADDR`: `nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443`
- `NANABUSH_SECURE`: `true`

⏳ **Next Steps:**
1. Apply updated glooscap deployment on ocp-ai-sno-2 cluster
2. Verify connection
3. Test gRPC calls

## Update Glooscap Deployment (ocp-ai-sno-2)

The glooscap operator deployment manifest has been updated with the Route address. Apply it to the ocp-ai-sno-2 cluster:

```bash
# Set context for glooscap cluster
# GLOOSCAP_CONTEXT="<ocp-ai-sno-2-context>"

# Apply updated deployment
oc --context=$GLOOSCAP_CONTEXT apply -f /home/dasm/org-dasmlab/tools/glooscap/operator/config/manager/manager.yaml

# Or patch deployment directly:
oc --context=$GLOOSCAP_CONTEXT patch deployment -n glooscap-system controller-manager \
  -p '{
    "spec": {
      "template": {
        "spec": {
          "containers": [{
            "name": "manager",
            "env": [
              {"name": "NANABUSH_GRPC_ADDR", "value": "nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443"},
              {"name": "NANABUSH_SECURE", "value": "true"}
            ]
          }]
        }
      }
    }
  }'
```

## Verify Connection

### From Glooscap Operator (ocp-ai-sno-2)

```bash
# Check environment variables
oc --context=$GLOOSCAP_CONTEXT exec -n glooscap-system deployment/controller-manager -- env | grep NANABUSH

# Check logs for connection attempts
oc --context=$GLOOSCAP_CONTEXT logs -n glooscap-system deployment/controller-manager -f | grep nanabush
```

### From Nanabush Server (ocp-sno-1050ti)

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Check logs for incoming connections
oc --context=$CONTEXT logs -n nanabush -l app=nanabush-grpc-server -f
```

## Test gRPC Connection

### Using grpcurl (if available)

```bash
# From a pod on ocp-ai-sno-2 (if Route is accessible)
grpcurl -insecure nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443 list

# Test CheckTitle
grpcurl -insecure nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443 \
  nanabush.v1.TranslationService.CheckTitle \
  -d '{"title":"Test","language_tag":"fr-CA","source_language":"EN"}'
```

### From Glooscap Operator

Create a test TranslationJob to trigger connection from glooscap to nanabush.

## Troubleshooting

### Connection Refused

1. **Verify Route is accessible:**
   ```bash
   curl -k https://nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org:443
   ```

2. **Check DNS resolution:**
   ```bash
   nslookup nanabush-grpc.apps.ocp-sno-1050ti.rh.dasmlab.org
   ```

3. **Verify HAProxy is forwarding traffic**

### TLS Errors

1. **Verify Route TLS configuration:**
   ```bash
   CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"
   oc --context=$CONTEXT get route -n nanabush nanabush-grpc -o yaml | grep -A 5 tls
   ```

2. **Check if gRPC client supports TLS:**
   - Ensure `NANABUSH_SECURE=true` is set
   - Verify client code handles TLS correctly

### gRPC Over HTTP/2 Issues

If gRPC over Route doesn't work:
1. Consider LoadBalancer instead
2. Use Service Mesh if available
3. Check OpenShift Router gRPC support

## Notes

- Route uses **edge TLS termination** - OpenShift Router terminates TLS
- gRPC over HTTP/2 may have limitations through edge termination
- HAProxy should automatically discover Routes in the cluster
- For production, consider mTLS via Service Mesh

