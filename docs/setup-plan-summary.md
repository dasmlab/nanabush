# Setup Plan Summary: OCP SNO 1050ti Post-Deployment

## Executive Summary

This document outlines the complete setup plan for `ocp-sno-1050ti.rh.dasmlab.org` to host Nanabush (vLLM platform) and enable cross-cluster communication with Glooscap running on `ocp-ai-sno-2`.

## Cluster Information

- **Cluster**: ocp-sno-1050ti.rh.dasmlab.org
- **Type**: Single Node OpenShift (SNO)
- **Node**: ocp-sno-1050ti
- **Resources**: 8 CPU, ~24GB RAM
- **GPU**: NVIDIA GPU (model to be verified)
- **Network**: 10.20.1.0/24 (node IP: 10.20.1.20)

## Required Components

### 1. MetalLB ⚡ (Priority: CRITICAL)

**Purpose**: Enable LoadBalancer services (required for external gRPC access)

**What it does**:
- Provides LoadBalancer service type on bare metal
- Assigns external IPs from configured pool
- Uses ARP mode for single node (Layer 2)

**Installation**:
- Operator: MetalLB Operator (from OperatorHub)
- Namespace: `metallb-system`
- Configuration: IP pool (e.g., 10.20.1.100-10.20.1.150)

**Why first**: Needed before other services that require LoadBalancer

**Time**: ~5 minutes

---

### 2. NVIDIA GPU Operator 🎮 (Priority: CRITICAL)

**Purpose**: Enable GPU access, CUDA support, and nvidia-smi in containers

**What it installs**:
- NVIDIA Driver Container (kernel module)
- NVIDIA Device Plugin (Kubernetes GPU scheduling)
- NVIDIA Container Toolkit (CUDA runtime)
- GPU Feature Discovery

**Installation**:
- Operator: NVIDIA GPU Operator (from OperatorHub)
- Custom Resource: ClusterPolicy
- Namespace: `gpu-operator-resources`

**Why before AI**: OpenShift AI requires GPU to be available

**Time**: ~15-20 minutes (driver installation is slow)

**Validation**: 
- `oc describe node | grep nvidia.com/gpu`
- Test pod with `nvidia-smi` command

---

### 3. OpenShift AI Bundle 🤖 (Priority: HIGH)

**Purpose**: AI/ML platform for model serving, pipelines, and Jupyter

**What it provides**:
- Model Serving (KServe/ModelMesh)
- Data Science Pipelines
- Jupyter Notebooks
- Dashboard
- vLLM integration support

**Installation**:
- Operator: "Red Hat AI/ML" (from OperatorHub)
- Custom Resource: DataScienceCluster
- Namespace: `redhat-ods-applications`

**Dependencies**: GPU + Storage must be ready

**Time**: ~10-15 minutes

**Configuration**:
- Enable model serving components
- Configure GPU node selectors
- Set up storage for models

---

### 4. Storage 📦 (Priority: MEDIUM)

**Purpose**: Persistent storage for models, data, and logs

**Options for SNO**:
- Local Storage Operator (recommended)
- NFS (if available)
- HostPath (development only)

**What to create**:
- StorageClass for local storage
- PVC for model storage (100GB+)
- PVC for data storage
- PVC for logs/audit

**Time**: ~5 minutes

---

### 5. Network Configuration 🌐 (Priority: MEDIUM)

**Purpose**: Enable cross-cluster communication

**What to configure**:
- Create `nanabush` namespace
- Network Policies for isolation
- LoadBalancer service for gRPC (via MetalLB)
- Routes for HTTP services (if needed)

**gRPC Service**:
- Type: LoadBalancer
- Port: 50051
- External IP: Assigned by MetalLB

**Time**: ~5 minutes

---

### 6. Monitoring & Observability 📊 (Priority: LOW)

**Purpose**: Track GPU usage and model performance

**Components**:
- GPU metrics (from NVIDIA operator)
- Prometheus (usually pre-installed)
- Custom dashboards

**Time**: ~5 minutes (mostly configuration)

---

### 7. Security & Compliance 🔒 (Priority: MEDIUM)

**Purpose**: Enforce security policies

**Components**:
- Security Context Constraints (SCCs) for GPU workloads
- Network Policies
- Pod Security Standards

**Time**: ~5 minutes

---

## Installation Order

```
1. MetalLB (5 min)
   ↓
2. Storage (5 min)
   ↓
3. NVIDIA GPU Operator (15-20 min) ⏳
   ↓
4. OpenShift AI (10-15 min)
   ↓
5. Network Configuration (5 min)
   ↓
6. Monitoring (5 min, parallel)
   ↓
7. Security Policies (5 min, parallel)
```

**Total Estimated Time**: ~45-55 minutes

## Script Structure

The `post-ocp-deployment.sh` script will:

1. **Pre-flight Checks**
   - Verify cluster connectivity
   - Check node resources
   - Detect GPU hardware
   - Verify network configuration

2. **Install MetalLB**
   - Subscribe to operator
   - Wait for CSV ready
   - Create MetalLB instance
   - Configure IP pool
   - Validate with test service

3. **Install NVIDIA GPU Operator**
   - Subscribe to operator
   - Wait for CSV ready
   - Create ClusterPolicy
   - Wait for driver installation (long wait)
   - Validate GPU access

4. **Configure Storage**
   - Install Local Storage Operator (if needed)
   - Create StorageClasses
   - Create PVCs for models/data

5. **Install OpenShift AI**
   - Subscribe to operator
   - Wait for CSV ready
   - Create DataScienceCluster
   - Wait for components ready
   - Validate

6. **Configure Network**
   - Create nanabush namespace
   - Set up Network Policies
   - Configure service exposure

7. **Post-Installation**
   - Run validation tests
   - Display connection info
   - Show next steps

## Configuration Parameters

The script will need these configurable parameters:

```bash
# MetalLB Configuration
METALLB_IP_POOL_START="10.20.1.100"
METALLB_IP_POOL_END="10.20.1.150"

# Storage Configuration
MODEL_STORAGE_SIZE="200Gi"
DATA_STORAGE_SIZE="100Gi"

# Network Configuration
NANABUSH_NAMESPACE="nanabush"
GRPC_PORT="50051"

# GPU Configuration (auto-detected)
GPU_DRIVER_VERSION="latest"  # or specific version
```

## Validation Checklist

After script completion, verify:

- [ ] MetalLB operator running
- [ ] LoadBalancer service gets external IP
- [ ] NVIDIA GPU Operator installed
- [ ] GPU visible: `oc describe node | grep nvidia`
- [ ] `nvidia-smi` works in test pod
- [ ] OpenShift AI operator installed
- [ ] DataScienceCluster status: Ready
- [ ] Storage classes available
- [ ] PVCs created and bound
- [ ] nanabush namespace created
- [ ] Network policies applied

## Troubleshooting

### MetalLB Issues
- Check IP pool range doesn't conflict
- Verify ARP mode for single node
- Check node network interface

### GPU Issues
- Verify GPU hardware: `lspci | grep -i nvidia`
- Check operator logs: `oc logs -n gpu-operator-resources`
- Driver installation can take 15+ minutes

### OpenShift AI Issues
- Ensure GPU is available first
- Check storage is provisioned
- Review operator logs

## Next Steps After Script

1. Deploy Nanabush gRPC service
2. Configure Tekton pipelines
3. Load initial translation model
4. Test gRPC from Glooscap cluster
5. Configure Glooscap operator endpoint

## Questions to Resolve

Before writing the script, please confirm:

1. **GPU Hardware**: What NVIDIA GPU model is in the 1050ti machine?
   - Check: `lspci | grep -i nvidia` on the node
   - This affects driver compatibility

2. **IP Pool Range**: What IP range should MetalLB use?
   - Current node IP: 10.20.1.20
   - Suggested: 10.20.1.100-10.20.1.150 (adjust if needed)

3. **Storage**: Do you have NFS available, or should we use local storage?
   - Local storage is simpler for SNO
   - NFS allows ReadWriteMany (better for models)

4. **Network Access**: How will Glooscap cluster reach Nanabush?
   - Direct network route?
   - VPN?
   - Exposed endpoint?

5. **Model Storage Size**: How large are the translation models?
   - Typical: 7B model ~14GB, 13B model ~26GB
   - Recommend: 200GB+ for multiple models

## Files to Create

1. `scripts/post-ocp-deployment.sh` - Main installation script
2. `kustomize/metallb/` - MetalLB manifests
3. `kustomize/nvidia/` - NVIDIA GPU operator manifests
4. `kustomize/ai/` - OpenShift AI manifests
5. `kustomize/storage/` - Storage configuration

## Review Checklist

Please review and confirm:

- [ ] Installation order is correct
- [ ] All required components are listed
- [ ] Configuration parameters are appropriate
- [ ] Validation steps are clear
- [ ] Troubleshooting section is helpful
- [ ] Questions are answered before script writing

---

**Ready to proceed?** Once you've reviewed and confirmed the plan, I'll write the `post-ocp-deployment.sh` script with all the installation steps, error handling, and validation.

