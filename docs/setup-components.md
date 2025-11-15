# Detailed Component Setup Plan for OCP SNO 1050ti

## Component Breakdown

### 1. MetalLB Operator & Configuration

**Why**: OCP SNO doesn't have a cloud LoadBalancer. MetalLB provides LoadBalancer services for bare metal.

**Installation Steps**:
1. Subscribe to MetalLB Operator from OperatorHub
2. Create MetalLB instance in `metallb-system` namespace
3. Configure IP address pool (ARP mode for single node)
4. Test with sample LoadBalancer service

**Configuration Details**:
```yaml
# MetalLB Instance
apiVersion: metallb.io/v1beta1
kind: MetalLB
metadata:
  name: metallb
  namespace: metallb-system
spec: {}

# IP Address Pool
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: default-pool
  namespace: metallb-system
spec:
  addresses:
  - 10.20.1.100-10.20.1.150  # Example range - adjust to your network

# L2 Advertisement (for ARP mode)
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: default
  namespace: metallb-system
spec:
  ipAddressPools:
  - default-pool
```

**Validation**:
- Create test LoadBalancer service
- Verify external IP assignment
- Test connectivity from external network

---

### 2. NVIDIA GPU Operator

**Why**: Required for GPU access, CUDA support, and nvidia-smi in containers.

**Components Installed**:
- NVIDIA Driver Container
- NVIDIA Device Plugin (for Kubernetes GPU scheduling)
- NVIDIA Container Toolkit
- GPU Feature Discovery
- Node Feature Discovery (optional)

**Installation Steps**:
1. Subscribe to NVIDIA GPU Operator from OperatorHub
2. Create ClusterPolicy custom resource
3. Wait for driver installation (may take 10-15 minutes)
4. Verify GPU detection

**Configuration Details**:
```yaml
# ClusterPolicy for NVIDIA components
apiVersion: nvidia.com/v1
kind: ClusterPolicy
metadata:
  name: cluster-policy
spec:
  driver:
    enabled: true
    repository: nvcr.io/nvidia
    version: "latest"
  devicePlugin:
    enabled: true
  toolkit:
    enabled: true
  operator:
    defaultRuntime: containerd
```

**Validation**:
- Check operator pods: `oc get pods -n gpu-operator-resources`
- Verify GPU in node: `oc describe node | grep nvidia.com/gpu`
- Test pod with GPU: Create pod requesting GPU resource
- Run `nvidia-smi` in pod to verify GPU access

**Hardware Requirements**:
- NVIDIA GPU must be physically present
- GPU must be compatible with NVIDIA drivers
- Check GPU model: `lspci | grep -i nvidia` (on node)

---

### 3. OpenShift AI (Red Hat AI/ML)

**Why**: Provides AI/ML platform, model serving, data science pipelines, and Jupyter support.

**Components**:
- OpenShift AI Operator (Red Hat AI/ML Operator)
- Data Science Pipelines Operator
- Model Serving Operator
- Jupyter Operator
- Dashboard components

**Installation Steps**:
1. Subscribe to "Red Hat AI/ML" operator from OperatorHub
2. Create DataScienceCluster custom resource
3. Configure storage for models and data
4. Set up authentication/authorization
5. Configure model serving backend

**Configuration Details**:
```yaml
# DataScienceCluster
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: default-dsc
spec:
  components:
    codeflare:
      managementState: Managed
    dashboard:
      managementState: Managed
    datasciencepipelines:
      managementState: Managed
    kserve:
      managementState: Managed
    modelmeshserving:
      managementState: Managed
    ray:
      managementState: Managed
    workbenches:
      managementState: Managed
```

**Storage Requirements**:
- Model storage: Large PVC (100GB+ recommended)
- Data storage: PVC for training/inference data
- Workspace storage: PVC for Jupyter notebooks

**GPU Integration**:
- Configure GPU node selectors for AI workloads
- Set resource limits for GPU allocation
- Configure vLLM serving runtime

**Validation**:
- Check operator status: `oc get dsc`
- Verify components are ready: `oc get pods -n redhat-ods-applications`
- Access dashboard: Route to OpenShift AI dashboard
- Test model serving endpoint

---

### 4. Storage Configuration

**Why**: Models, data, and logs need persistent storage.

**Options for SNO**:
- Local Storage Operator (for local PVs)
- NFS (if NFS server available)
- HostPath (development only)

**Installation Steps**:
1. Install Local Storage Operator (if using local storage)
2. Create StorageClass
3. Create PVCs for:
   - Model storage (large, ReadWriteMany preferred)
   - Data storage (ReadWriteOnce or ReadWriteMany)
   - Log storage (ReadWriteOnce)

**Configuration**:
```yaml
# Example StorageClass for local storage
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
```

---

### 5. Network Configuration

**Why**: Enable cross-cluster communication and service exposure.

**Components**:
- Routes (for HTTP/HTTPS services)
- LoadBalancer services (via MetalLB)
- Network Policies (for isolation)
- DNS configuration

**Configuration**:
- External access: Use Routes for HTTP services
- gRPC access: Use LoadBalancer service with MetalLB
- Internal access: ClusterIP services
- Cross-cluster: Configure network routes or VPN

**gRPC Service Example**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: nanabush-grpc
  namespace: nanabush
spec:
  type: LoadBalancer
  ports:
  - port: 50051
    targetPort: 50051
    protocol: TCP
    name: grpc
  selector:
    app: nanabush
```

---

### 6. Monitoring & Observability

**Why**: Track GPU usage, model performance, and cluster health.

**Components**:
- Prometheus (usually pre-installed)
- Grafana (for dashboards)
- GPU metrics exporter
- Custom metrics for model inference

**Configuration**:
- GPU metrics: NVIDIA GPU exporter (part of GPU operator)
- Model metrics: Custom Prometheus metrics
- Alerting: Set up alerts for GPU utilization, model errors

**Dashboards**:
- GPU utilization
- Model inference latency
- Request throughput
- Error rates

---

### 7. Security & Compliance

**Why**: Enforce security policies and compliance requirements.

**Components**:
- Security Context Constraints (SCCs)
- Network Policies
- Pod Security Standards
- RBAC configuration

**Configuration**:
- GPU workloads: May need privileged SCC
- Network isolation: Policies for vLLM namespace
- Audit logging: Enable for all operations
- Image security: Scan container images

**SCC Example**:
```yaml
# Allow GPU access (may need privileged)
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: gpu-workload
allowHostNetwork: false
allowPrivilegedContainer: true
allowedCapabilities:
- SYS_ADMIN
```

---

## Installation Script Structure

The `post-ocp-deployment.sh` script should:

1. **Pre-flight checks**:
   - Verify cluster access
   - Check node resources
   - Verify GPU hardware
   - Check network connectivity

2. **Install MetalLB**:
   - Subscribe to operator
   - Wait for CSV to be ready
   - Create MetalLB instance
   - Configure IP pool
   - Validate

3. **Install NVIDIA GPU Operator**:
   - Subscribe to operator
   - Wait for CSV to be ready
   - Create ClusterPolicy
   - Wait for driver installation
   - Validate GPU access

4. **Install OpenShift AI**:
   - Subscribe to operator
   - Wait for CSV to be ready
   - Create DataScienceCluster
   - Configure storage
   - Wait for components to be ready
   - Validate

5. **Configure Storage**:
   - Install Local Storage Operator (if needed)
   - Create StorageClasses
   - Create PVCs for models/data

6. **Configure Network**:
   - Create nanabush namespace
   - Set up Network Policies
   - Configure Routes (if needed)

7. **Post-installation**:
   - Run validation tests
   - Display connection information
   - Show next steps

## Dependencies & Order

```
MetalLB (first)
    ↓
Storage (early)
    ↓
NVIDIA GPU Operator
    ↓
OpenShift AI (depends on GPU + Storage)
    ↓
Network Configuration
    ↓
Monitoring (can run in parallel)
    ↓
Security Policies (apply throughout)
```

## Estimated Time

- MetalLB: ~5 minutes
- NVIDIA GPU Operator: ~15-20 minutes (driver installation)
- OpenShift AI: ~10-15 minutes
- Storage: ~5 minutes
- Total: ~35-45 minutes

## Rollback Plan

Each component should be idempotent and reversible:
- Operators can be uninstalled via OperatorHub
- Custom resources can be deleted
- Storage can be cleaned up (with caution)

## Next Steps After Script

1. Deploy Nanabush gRPC service
2. Configure Tekton pipelines
3. Load initial model
4. Test gRPC connectivity from Glooscap cluster
5. Configure Glooscap operator with Nanabush endpoint

