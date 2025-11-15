# Post-Deployment Setup Plan for OCP SNO 1050ti

This document outlines the complete setup plan for preparing the `ocp-sno-1050ti` cluster to host Nanabush (vLLM platform) and OCP AI components.

## Overview

The `ocp-sno-1050ti.rh.dasmlab.org` cluster needs to be configured with:
1. **MetalLB** - LoadBalancer service support (critical for external access and future second NIC)
2. **NVIDIA GPU Support** - Drivers, operator, and CUDA toolkit for GPU acceleration
3. **OCP AI Bundle** - OpenShift AI operator and components for model serving
4. **Infrastructure Prerequisites** - Storage, networking, monitoring

## Prerequisites

- OCP SNO cluster deployed and accessible
- Cluster admin access (`dasm` user with cluster-admin role)
- Network connectivity to:
  - Red Hat registries (for operators)
  - Internet (for NVIDIA driver downloads if needed)
- GPU hardware detected on the node

## Setup Components

### 1. MetalLB Setup

**Purpose**: Enable LoadBalancer services on bare metal/SNO cluster

**Components**:
- MetalLB Operator (via OperatorHub)
- MetalLB instance configuration
- IP address pool configuration
- BGP/ARP mode selection (ARP for single node)

**Steps**:
1. Install MetalLB Operator from OperatorHub
2. Create MetalLB instance
3. Configure IP address pool (for LoadBalancer services)
4. Verify LoadBalancer service creation works

**Configuration**:
- Mode: ARP (Layer 2) for single node
- IP Pool: Define range based on network topology
- Namespace: `metallb-system`

**Validation**:
- Create test LoadBalancer service
- Verify external IP assignment
- Test connectivity

### 2. NVIDIA GPU Support

**Purpose**: Enable CUDA workloads and GPU access for vLLM inference

**Components**:
- NVIDIA GPU Operator (via OperatorHub)
- NVIDIA Driver Container
- NVIDIA Device Plugin
- NVIDIA Container Toolkit
- GPU Feature Discovery
- Node Feature Discovery (optional, for auto-detection)

**Steps**:
1. Install NVIDIA GPU Operator from OperatorHub
2. Configure ClusterPolicy for NVIDIA components
3. Verify GPU detection (`nvidia-smi` in pods)
4. Test GPU workload scheduling

**Configuration**:
- Driver version: Latest compatible with hardware
- Toolkit version: Latest CUDA toolkit
- Device plugin: Enable for GPU scheduling
- MIG mode: Disabled (single GPU) or enabled if supported

**Validation**:
- Check GPU operator pods are running
- Verify GPU is visible to cluster (`oc describe node | grep nvidia`)
- Run test pod with GPU access
- Execute `nvidia-smi` in pod

### 3. OpenShift AI Bundle

**Purpose**: Provide AI/ML platform capabilities for model serving and pipelines

**Components**:
- OpenShift AI Operator (Red Hat AI/ML)
- Data Science Pipelines Operator
- Model Serving components
- Jupyter notebook support
- vLLM integration components

**Steps**:
1. Install OpenShift AI Operator from OperatorHub
2. Create DataScienceCluster custom resource
3. Configure model serving backend
4. Set up storage for models and data
5. Configure authentication/authorization

**Configuration**:
- Namespace: `redhat-ods-applications` (default)
- Storage: PVC for model storage
- Serving runtime: vLLM or compatible
- Resource limits: GPU allocation for AI workloads

**Validation**:
- Verify operator installation
- Check DataScienceCluster status
- Test model serving endpoint
- Verify GPU allocation to AI workloads

### 4. Storage Setup

**Purpose**: Persistent storage for models, data, and logs

**Components**:
- StorageClass configuration
- PVC creation for model storage
- Backup storage configuration

**Configuration**:
- StorageClass: Local storage or NFS (for SNO)
- Model storage: Large PVC for model weights
- Data storage: PVC for training data
- Log storage: PVC for audit logs

### 5. Network Configuration

**Purpose**: Enable cross-cluster communication and service exposure

**Components**:
- Service mesh (optional, for secure inter-service communication)
- Routes for external access
- Network policies for isolation
- DNS configuration

**Configuration**:
- External access: Routes via MetalLB LoadBalancer
- Internal access: ClusterIP services
- Cross-cluster: VPN or exposed endpoints

### 6. Monitoring & Observability

**Purpose**: Track GPU usage, model performance, and cluster health

**Components**:
- Prometheus operator (if not already installed)
- Grafana dashboards
- GPU metrics collection
- Alerting rules

**Configuration**:
- GPU metrics: NVIDIA GPU exporter
- Model metrics: Custom metrics for inference
- Alerting: GPU utilization, model errors

### 7. Security & Compliance

**Purpose**: Enforce security policies and compliance

**Components**:
- Security Context Constraints (SCCs)
- Network Policies
- Pod Security Standards
- Compliance Operator (optional)

**Configuration**:
- GPU access: Privileged SCC for GPU workloads
- Network isolation: Policies for vLLM namespace
- Audit logging: Enable for all operations

## Installation Order

Recommended sequence to avoid dependency issues:

1. **MetalLB** (first - needed for LoadBalancer services)
2. **Storage** (early - needed by operators)
3. **NVIDIA GPU Operator** (before AI - GPU must be available)
4. **OpenShift AI Bundle** (last - depends on GPU and storage)
5. **Monitoring** (can run in parallel)
6. **Security Policies** (apply throughout)

## Validation Checklist

After setup, verify:

- [ ] MetalLB operator installed and running
- [ ] LoadBalancer service gets external IP
- [ ] NVIDIA GPU Operator installed
- [ ] GPU visible via `nvidia-smi` in test pod
- [ ] OpenShift AI Operator installed
- [ ] DataScienceCluster created and ready
- [ ] Storage classes available
- [ ] Model serving endpoint accessible
- [ ] Cross-cluster connectivity (from Glooscap cluster)
- [ ] Monitoring dashboards functional

## Troubleshooting

### Common Issues

1. **MetalLB not assigning IPs**
   - Check IP pool configuration
   - Verify ARP mode for single node
   - Check node network configuration

2. **GPU not detected**
   - Verify hardware is present
   - Check NVIDIA operator logs
   - Verify driver installation

3. **OCP AI not starting**
   - Check GPU availability
   - Verify storage is provisioned
   - Check operator logs

4. **Cross-cluster connectivity**
   - Verify network routes
   - Check firewall rules
   - Test DNS resolution

## Next Steps After Setup

1. Deploy Nanabush gRPC service
2. Configure Tekton pipelines for translation
3. Set up model storage and loading
4. Configure Glooscap to connect to Nanabush
5. Test end-to-end translation workflow

## References

- [MetalLB Documentation](https://metallb.universe.tf/)
- [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/)
- [OpenShift AI Documentation](https://access.redhat.com/documentation/en-us/red_hat_openshift_ai/)
- [OCP SNO Best Practices](https://docs.openshift.com/container-platform/latest/installing/installing_sno/)

