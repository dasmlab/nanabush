## OpenShift RHEL AI Enablement

This README tracks the work required to stand up Red Hat OpenShift AI (RHOAI) capabilities on top of our GPU-backed OpenShift footprint so the `nanabush` vLLM services can be hosted in-cluster and consumed by the Glooscap operator.

### 1. Baseline Prerequisites

- Confirm target cluster runs OpenShift 4.15+ with workload nodes on RHEL CoreOS/RHEL 9.
- Allocate at least one GPU-equipped worker (A100/H100/MI300) with SR-IOV enabled at the hypervisor.
- Ensure Red Hat subscription entitlements cover OpenShift AI, NCCL/NVIDIA GPU operator, and Compliance Operator.
- Install local tooling on the bastion jump host: `oc`, `kubectl`, `openshift-install`, `rosa` (if ROSA), `kustomize`, `helm`, and `podman`.

### 2. Host & Cluster Preparation

- Update BIOS/firmware and validate IOMMU passthrough for the GPU hosts.
- Enable RHEL repositories (`rhel-9-for-x86_64-baseos-rpms`, `rhel-9-for-x86_64-appstream-rpms`, `codeready-builder`) and sync errata via Satellite or `dnf`.
- Install GPU drivers on bare-metal RHEL nodes when not managed by GPU Operator; otherwise leave to operator.
- Configure OpenShift MachineSets/MachinePools dedicated to GPU workloads with proper taints/tolerations and labeling (`node-role.kubernetes.io/gpu=`, `nvidia.com/gpu.present=true`).
- Deploy Node Feature Discovery (NFD) operator to advertise GPU capabilities.

### 3. Core Operators

- Install NVIDIA GPU Operator (or AMD parallel) with driver toolkit enabled and pinned to the GPU MachineSets.
- Install Red Hat OpenShift AI operator (`redhat-aiservice-operator`) scoped to the `openshift-operators` namespace.
- Enable Data Science Pipelines (DSP) and Model Serving components inside RHOAI.
- Validate cluster monitoring stack exposes GPU metrics (DCGM exporter, Prometheus federation).

### 4. Storage & Networking Foundations

- Provision high-throughput storage class (OCS/ODF or RWX CSI) for model artifacts, pipelines, and notebooks.
- Configure internal registry or object storage for container/model versioning (Quay, S3-compatible).
- Ensure service mesh (Istio / OpenShift Service Mesh) is installed for secure, mTLS-protected service-to-service calls.
- Codify baseline NetworkPolicies and EgressFirewalls that match zero-exfil expectations from `tools/glooscap/docs/security.md`.

### 5. RHOAI Workspace Layout

- Create dedicated `nanabush-ai` or equivalent project with:
  - Namespace quotas and limit ranges sized for GPU workloads.
  - Service accounts and secrets for vLLM weight pulls.
  - ConfigMaps for model configuration and tokenizer assets.
- Stand up data science projects/workbenches needed for model fine-tuning or evaluation.

### 6. GPU-Backed vLLM Deployment Plan

- Package vLLM inference image via Helm (`helm/vllm/`) with Service Mesh annotations, OTEL sidecars, and GPU resource requests.
- Decide on execution mode(s):
  - `TektonJob` flow for on-demand translation (`tekton/translation-pipeline.yaml` scaffolding).
  - Always-on service with HorizontalPodAutoscaler and PodDisruptionBudgets.
- Wire Grafana dashboards to GPU metrics and create alert rules for thermal throttling, memory pressure, and inference errors.

### 7. Security & Compliance Tasks

- Apply Compliance Operator profiles (e.g., `ocp4-moderate`, `cis`) across GPU nodes and control plane.
- Enforce runtimeClass (Kata/gVisor) for translation jobs where compatible; document exceptions for GPU pods.
- Implement admission controls validating CRDs, resource quotas, and allowed container registries.
- Integrate OTEL tracing with centralized logging (Loki/Elastic) for immutable audit trails.

### 8. Integration with Glooscap Tooling

- Document RHOAI endpoints and credentials that the Glooscap operator (`tools/glooscap/operator/`) will consume.
- Provide Tekton pipeline bindings and secrets that bridge operator CRDs to RHOAI workloads.
- Align sanitized feedback loops with RHOAI model registry for retraining cadence.
- Update cross-repo docs (`tools/glooscap/docs/vllm-integration.md`) once endpoints are live.

### 9. Outstanding Questions / TODOs

- [ ] Confirm exact GPU SKU availability and driver/toolkit matrix.
- [ ] Decide between ROSA w/ hosted control plane vs. self-managed OCP footprint.
- [ ] Evaluate whether to offload notebook experiences to Red Hat-managed RHOAI instances.
- [ ] Map SLA/SLO targets for inference latency and throughput.
- [ ] Draft runbooks for incident response, GPU failure remediation, and capacity scaling.

Keep this README in sync as tasks progress and create subordinate documents (runbooks, manifests, pipelines) under `infra/openshift/` as each area matures.

