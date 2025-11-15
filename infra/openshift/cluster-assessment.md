## OpenShift & GPU Baseline Assessment

Use this checklist to confirm the current OpenShift environment satisfies the prerequisites for enabling Red Hat OpenShift AI (RHOAI) with GPU-backed workloads.

### Cluster Metadata

- Capture cluster version and channel:
  ```bash
  oc version
  oc get clusterversion/version -o jsonpath='{.status.history[0]}'
  ```
- Result: client `4.8.11`, server `4.18.26`, Kubernetes `v1.31.13`. Latest history entry shows image `quay.io/openshift-release-dev/ocp-release@sha256:dcd5fce7701d1e568ffb1065800a4aa34c911910400209224e702b951412171d`, completed `2025-10-23T13:22:05Z`.
- Record installer footprint (`rosa`, self-managed, hosted control plane, etc.).
- Note base domain, API endpoints, and authentication providers.

### Node & GPU Inventory

- Enumerate nodes and labels:
  ```bash
  oc get nodes -L node-role.kubernetes.io/worker,node-role.kubernetes.io/gpu,nvidia.com/gpu.present
  ```
- Result (2025-11-11): single node `00-0c-29-37-41-c9` with roles `control-plane,master,worker` and no GPU labels present.
- Gather GPU details per node (requires DCGM or `oc debug`):
  ```bash
  for node in $(oc get nodes -l nvidia.com/gpu.present=true -o name); do
    oc adm debug "$node" -- chroot /host nvidia-smi
  done
  ```
- `oc debug node/00-0c-29-37-41-c9 -- chroot /host nvidia-smi` returns `No such file or directory`.
- `oc debug node/00-0c-29-37-41-c9 -- chroot /host lspci` shows VMware virtual GPU (`VGA controller: VMware SVGA II`) and no physical NVIDIA/AMD/Intel PCI IDs, indicating the guest OS currently lacks a passed-through accelerator (likely needs host passthrough configuration).
- Confirm SR-IOV/IOMMU state for bare-metal hosts.
- Identify firmware/BIOS versions; record pending updates.

### Subscription & Operator Entitlements

- List installed operators and sources:
  ```bash
  oc get operators.operators.coreos.com -A
  oc get catalogsources -n openshift-marketplace
  ```
- Verify Red Hat subscriptions cover GPU Operator, RHOAI, Compliance Operator, Service Mesh.
- Confirm access to required registries (`registry.redhat.io`, `quay.io`, vendor GPU registries).

### Storage & Networking Baseline

- Document available storage classes and performance characteristics:
  ```bash
  oc get sc
  ```
- Result: only `lvms-vg1` storage class (provisioner `topolvm.io`, `WaitForFirstConsumer`, expansion enabled).
- Check internal registry status:
  ```bash
  oc get configs.imageregistry.operator.openshift.io/cluster -o yaml
  ```
- Review service mesh presence (`istio-system` or `openshift-operators` namespace).
- Audit existing NetworkPolicies/EgressFirewalls affecting GPU namespaces.

### Compliance & Monitoring

- Determine if Compliance Operator is deployed:
  ```bash
  oc get complianceoperator -n openshift-compliance
  ```
- Result: API endpoint `complianceoperator` not registered (v1.7.1 AmI). Operator is installed (`compliance-operator.v1.7.1` CSV in `openshift-compliance` namespace), but CRDs exposed are `compliancesuites`, `scansettings`, `profilebundles`, etc. Use those resources for status.
- Inspect cluster monitoring for DCGM exporters or gpu-metrics:
  ```bash
  oc get pods -n openshift-monitoring | grep gpu
  ```
- Result: no GPU-specific pods running; standard monitoring stack only.
- Review logging/tracing stack alignment with `tools/glooscap/docs/security.md`.

### Action Items

- [x] Populate this file with live cluster data once commands are executed.
- [ ] File follow-up issues for any gaps (missing GPU nodes, operator subscriptions, storage throughput, etc.).


