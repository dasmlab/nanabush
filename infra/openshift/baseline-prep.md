## Baseline Host & Cluster Preparation

These steps should be executed before installing GPU/RHOAI operators. They bring RHEL/OpenShift nodes to a predictable state and align with the zero-exfil posture described in `tools/glooscap/docs/security.md`.

### 1. Firmware & BIOS Validation

- Update motherboard BIOS, BMC, and GPU firmware to vendor-recommended versions.
- Enable the following settings for GPU nodes:
  - SR-IOV / VT-d (Intel) or AMD-Vi / IOMMU.
  - Above 4G decoding / Resizable BAR (for PCIe GPU address space).
  - NUMA alignment settings per vendor best practices.
- Document firmware revisions in `infra/openshift/cluster-assessment.md`.

### 2. RHEL Repository Configuration

- Ensure Satellite or RHSM provides:
  - `rhel-9-for-x86_64-baseos-rpms`
  - `rhel-9-for-x86_64-appstream-rpms`
  - `codeready-builder-for-rhel-9-x86_64-rpms`
  - Optional: `rhocp-4.15-for-rhel-9-x86_64-rpms` if managing workers manually.
- Mirror NVIDIA/AMD driver channels internally when possible.
- Use Ansible or Satellite activation keys to enforce repository consistency.

### 3. OS Patching & Kernel Tuning

- Apply latest errata:
  ```bash
  sudo dnf update -y && sudo reboot
  ```
- Install additional packages (only when not fully managed by Machine Config Operator):
  - `kernel-devel`, `kernel-headers` (required for GPU driver builds).
  - `pciutils`, `ipmitool`, `ethtool` for diagnostics.
- Configure tuned profiles for high-performance workloads, e.g. `openshift-node-performance` if available.

### 4. Node Feature Discovery (NFD)

- Install NFD Operator from OperatorHub (namespace `openshift-operators`).
- Create an `NFD` custom resource targeting GPU nodes:
  ```yaml
  apiVersion: nfd.openshift.io/v1
  kind: NodeFeatureDiscovery
  metadata:
    name: nfd-instance
    namespace: openshift-operators
  spec:
    workerConfig:
      configData: |
        sources:
          pci:
            deviceLabelFields: ["vendor", "device"]
  ```
- Verify GPU labels appear (`feature.node.kubernetes.io/pci-10de.present=true` for NVIDIA).

### 5. MachineSet / MachineConfig Updates

- Define dedicated GPU MachineSets (or MachinePools for managed offerings) with:
  - Taints: `nvidia.com/gpu=true:NoSchedule`.
  - Labels: `node-role.kubernetes.io/gpu=""`, `nvidia.com/gpu.present=true`.
  - Instance types sized for chosen GPUs, ensuring adequate vCPU/RAM.
- Create MachineConfig / MachineConfigPool to apply kernel args (e.g., `intel_iommu=on`), hugepages if needed, and optional chrony configuration aligned with low-latency requirements.

### 6. Security Baselines

- Enforce SELinux in enforcing mode (`getenforce`).
- Configure `chronyd` to trusted NTP sources; record in compliance documentation.
- Ensure `fips-mode-setup` status matches compliance requirements.
- Validate bastion hosts use hardened SSH policies (no password auth, strong ciphers).

### 7. Validation Checklist

- [ ] GPU nodes show correct labels/taints.
- [ ] NFD advertising GPU vendor/device IDs.
- [ ] Kernel args applied (verify via `oc debug node/<name> -- chroot /host cat /proc/cmdline`).
- [ ] Required repos enabled and accessible.
- [ ] Firmware versions documented and within vendor support matrix.
- [ ] Baseline security controls (SELinux enforcing, FIPS if required) confirmed.

Update this document as procedures evolve or additional configuration items emerge.

