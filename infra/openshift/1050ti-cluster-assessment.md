# 1050Ti Cluster Assessment - OpenShift SNO with GTX 1050 Ti Mobile

**Cluster:** `api.ocp-sno-1050ti.rh.dasmlab.org:6443`  
**Assessment Date:** 2025-11-11  
**Context:** `default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm`

## Cluster Metadata

- **Server Version:** OpenShift 4.20.2 (Kubernetes v1.33.5)
- **Client Version:** 4.8.11
- **Cluster Type:** Single Node OpenShift (SNO) - baremetal
- **Node Name:** `ocp-sno-1050ti`
- **Node Roles:** `control-plane,master,worker`
- **Cluster Age:** ~25 hours
- **Latest Update:** Completed 2025-11-14T01:27:11Z (version 4.20.2)

## GPU Hardware Detection ✅

**GPU Detected:** NVIDIA Corporation GP107M [GeForce GTX 1050 Ti Mobile] (rev a1)

### PCI Device Details
```bash
oc debug node/ocp-sno-1050ti -- chroot /host lspci | grep -iE 'nvidia|vga'
```

**Results:**
- `00:02.0 VGA compatible controller: Intel Corporation HD Graphics 630 (rev 04)` (integrated graphics)
- `01:00.0 3D controller: NVIDIA Corporation GP107M [GeForce GTX 1050 Ti Mobile] (rev a1)` ✅

### Driver Status ✅

**Current State:** NVIDIA drivers **INSTALLED and RUNNING**
- **Driver Version:** 580.95.05
- **CUDA Version:** 13.0
- **nvidia-smi:** ✅ Working (verified in driver container)
- **GPU Detected:** NVIDIA GeForce GTX 1050 Ti
  - VRAM: 4096 MiB
  - Bus ID: 00000000:01:00.0
  - Temperature: 52°C
  - Status: Ready, no processes running
- **Node Labels:** ✅ All NVIDIA labels present
  - `nvidia.com/gpu.present=true`
  - `nvidia.com/cuda.driver-version.full=580.95.05`
  - `nvidia.com/gpu.compute.major=6`

## Node Inventory

```bash
oc get nodes -L node-role.kubernetes.io/worker,node-role.kubernetes.io/gpu,nvidia.com/gpu.present,nvidia.com/gpu.product
```

**Current Node Labels:**
- Node: `ocp-sno-1050ti`
- Status: `Ready`
- GPU labels: **None** (will appear after GPU Operator install)

## Operator Subscriptions

### Available Catalogs
- ✅ `redhat-operators` - Available (Red Hat)
- ✅ `redhat-marketplace` - Available (Red Hat)

### Installed Operators ✅
- ✅ **NVIDIA GPU Operator** v25.10.0 - **INSTALLED and READY**
  - ClusterPolicy: `gpu-cluster-policy` (Status: ready)
  - All daemonsets running: driver, device-plugin, DCGM, GFD, container-toolkit
- ✅ **Node Feature Discovery (NFD)** v4.20.0 - **INSTALLED**
  - GPU hardware detected and labeled
- ✅ **LVMS Operator** v4.20.0 - **INSTALLED**
  - Storage class `lvms-vg1` ready
- ❌ Red Hat OpenShift AI (RHOAI) - **Not installed** (Next step)
- ❌ Service Mesh - **Not installed** (Optional)
- ❌ Compliance Operator - **Not installed** (Optional)

## Storage Baseline ✅

**Storage Classes:**
- ✅ `lvms-vg1` (default) - **Configured and Ready**
  - Provisioner: `topolvm.io`
  - Volume Binding Mode: `WaitForFirstConsumer`
  - Filesystem: `xfs`
  - Device Class: `vg1`
  - Reclaim Policy: `Delete`
  - Volume Expansion: Enabled
- **LVMCluster:** `my-lvmcluster` in `openshift-storage` namespace - Status: Ready
- **Volume Group:** `vg1` created using `/dev/sdb` (931.5GB available)
- **Thin Pool:** `thin-pool-1` configured with 90% size and 10x overprovision ratio

## Action Items

### Immediate Next Steps

1. **Install NVIDIA GPU Operator** ✅
   - [x] Create OperatorGroup for `nvidia-gpu-operator` namespace
   - [x] Subscribe to NVIDIA GPU Operator from `certified-operators` catalog
   - [x] Create ClusterPolicy to enable driver installation
   - [x] Install Node Feature Discovery (NFD) to detect GPU hardware
   - [x] Verify driver installation via `nvidia-smi` (Driver 580.95.05, CUDA 13.0)
   - [x] Confirm node labels appear (`nvidia.com/gpu.present=true`)
   - [x] All GPU Operator daemonsets running and ready

2. **Configure Node Feature Discovery (NFD)** ✅
   - [x] Install NFD operator from `redhat-operators` catalog
   - [x] Create NodeFeatureDiscovery resource
   - [x] Verify NFD worker daemonset is running
   - [x] Confirm GPU features are detected and labeled (`feature.node.kubernetes.io/pci-0302_10de.present=true`)

3. **Install Red Hat OpenShift AI (RHOAI)**
   - [ ] Verify prerequisites (GPU Operator, storage)
   - [ ] Subscribe to RHOAI operator from `redhat-operators`
   - [ ] Configure RHOAI components for single-node deployment

4. **Storage Setup** ✅
   - [x] Determine storage strategy (LVMS using `/dev/sdb`)
   - [x] Install LVMS operator from `redhat-operators` catalog
   - [x] Create LVMCluster configuration with auto-discovery
   - [x] Verify storage class `lvms-vg1` is created and ready
   - [x] Test PVC creation and binding behavior (WaitForFirstConsumer confirmed)

5. **Baseline Security & Compliance**
   - [ ] Install Compliance Operator (optional, per security requirements)
   - [ ] Apply baseline compliance profiles if needed
   - [ ] Review SCC requirements for GPU workloads

### Verification Commands

Once GPU Operator is installed, verify:
```bash
# Check GPU operator pods
oc get pods -n nvidia-gpu-operator

# Verify node labels
oc get node ocp-sno-1050ti --show-labels | grep nvidia

# Check GPU via nvidia-smi
oc debug node/ocp-sno-1050ti -- chroot /host nvidia-smi

# Verify GPU resources exposed to cluster
oc describe node ocp-sno-1050ti | grep -A 10 nvidia.com/gpu
```

## Notes

- **Single Node Deployment:** This is an SNO cluster, so all workloads (control plane + workers) run on the same node. Resource allocation needs careful planning.
- **GPU Model:** GTX 1050 Ti Mobile has ~4GB VRAM - suitable for smaller models, may need optimization for larger workloads.
- **Integrated Graphics:** Intel HD Graphics 630 is also present - consider workload scheduling to avoid conflicts.
- **OpenShift 4.20.2:** Very recent version - should have excellent GPU Operator and RHOAI support.

