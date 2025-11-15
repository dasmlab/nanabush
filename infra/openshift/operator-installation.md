## Operator Installation Plan

This playbook records the sequence and configuration needed to install the core operators that underpin the RHOAI + GPU stack for `nanabush`.

### 1. NVIDIA GPU Operator (or AMD Equivalent)

#### Prerequisites
- GPU nodes labeled as described in `baseline-prep.md`.
- Node Feature Discovery operator running and reporting PCI capabilities.
- Cluster has access to `registry.redhat.io/nvidia` (or vendor mirror).

#### Installation Steps
1. Create OperatorGroup in `openshift-operators` (if not already present):
   ```yaml
   apiVersion: operators.coreos.com/v1
   kind: OperatorGroup
   metadata:
     name: openshift-operators
     namespace: openshift-operators
   spec:
     targetNamespaces:
       - openshift-operators
   ```
2. Subscribe to GPU Operator:
   ```yaml
   apiVersion: operators.coreos.com/v1alpha1
   kind: Subscription
   metadata:
     name: nvidia-gpu-operator
     namespace: openshift-operators
   spec:
     channel: stable
     name: gpu-operator-certified
     source: redhat-operators
     sourceNamespace: openshift-marketplace
     installPlanApproval: Automatic
   ```
3. Configure `ClusterPolicy` for GPU Operator:
   ```yaml
   apiVersion: nvidia.com/v1
   kind: ClusterPolicy
   metadata:
     name: gpu-cluster-policy
   spec:
     mig:
       strategy: single
     driver:
       enabled: true
     cdi:
       enabled: true
     dcgmExporter:
       serviceMonitor:
         enabled: true
   ```
4. Verify deployment:
   ```bash
   oc get pods -n nvidia-gpu-operator
   oc get node -L nvidia.com/gpu.present
   ```

### 2. Red Hat OpenShift AI Operator

#### Prerequisites
- Entitlement to RHOAI catalog (`redhat-aiservice-operator`).
- Storage class with RWX support for workbenches/pipelines.
- Namespace prepared for AI workloads (`nanabush-ai` or similar).

#### Installation Steps
1. Create CatalogSource if disconnected:
   ```yaml
   apiVersion: operators.coreos.com/v1alpha1
   kind: CatalogSource
   metadata:
     name: rhoai-catalog
     namespace: openshift-marketplace
   spec:
     sourceType: grpc
     image: registry.redhat.io/openshift-ai/aiservice-catalog:latest
   ```
2. OperatorGroup scoped to AI namespace:
   ```yaml
   apiVersion: operators.coreos.com/v1
   kind: OperatorGroup
   metadata:
     name: nanabush-ai-og
     namespace: nanabush-ai
   spec:
     targetNamespaces:
       - nanabush-ai
   ```
3. Subscription:
   ```yaml
   apiVersion: operators.coreos.com/v1alpha1
   kind: Subscription
   metadata:
     name: openshift-ai-sub
     namespace: nanabush-ai
   spec:
     channel: stable
     name: rhods-operator
     source: rhoai-catalog
     sourceNamespace: openshift-marketplace
     installPlanApproval: Manual
   ```
4. Approve InstallPlan, then create a `DataScienceCluster`:
   ```yaml
   apiVersion: dscinitialization.opendatahub.io/v1
   kind: DSCI
   metadata:
     name: default-dsci
     namespace: nanabush-ai
   spec:
     applications:
       datasciencepipelines:
         managementState: Managed
       modelmeshserving:
         managementState: Managed
       workbenches:
         managementState: Managed
   ```
5. Verify endpoints (`oc get routes -n nanabush-ai`).

### 3. OpenShift Service Mesh (Istio)

#### Prerequisites
- Cluster monitoring operators healthy.
- Sufficient capacity for control plane (3 replicas of istiod).

#### Installation Steps
1. Install Operators in order:
   - Red Hat OpenShift Elasticsearch (if using logging).
   - Red Hat OpenShift distributed tracing (Jaeger).
   - Red Hat OpenShift Service Mesh (Maistra).
2. Create OperatorGroup in `openshift-operators` (similar to GPU operator if not present).
3. Subscribe to Service Mesh operator:
   ```yaml
   apiVersion: operators.coreos.com/v1alpha1
   kind: Subscription
   metadata:
     name: servicemeshoperator
     namespace: openshift-operators
   spec:
     channel: stable
     name: servicemeshoperator
     source: redhat-operators
     sourceNamespace: openshift-marketplace
   ```
4. Deploy a `ServiceMeshControlPlane` in a dedicated namespace (e.g., `istio-system`):
   ```yaml
   apiVersion: maistra.io/v2
   kind: ServiceMeshControlPlane
   metadata:
     name: basic
     namespace: istio-system
   spec:
     version: v2.5
     policy:
       type: Istiod
     tracing:
       type: Jaeger
     addons:
       kiali:
         enabled: true
   ```
5. Define `ServiceMeshMemberRoll` to onboard application namespaces (`nanabush-ai`, `nanabush`):
   ```yaml
   apiVersion: maistra.io/v1
   kind: ServiceMeshMemberRoll
   metadata:
     name: default
     namespace: istio-system
   spec:
     members:
       - nanabush-ai
       - nanabush
   ```

### 4. Compliance & Logging Operators

- Ensure Red Hat Compliance Operator runs cluster-wide:
  ```bash
  oc apply -f https://raw.githubusercontent.com/ComplianceAsCode/content/master/openshift/compliance-operator-sub.yaml
  ```
- Configure `ScanSettingBinding` for `ocp4-moderate` or `cis`.
- Deploy OpenShift Logging (LokiStack/ClusterLogging) to capture GPU operator and RHOAI logs.

### 5. Post-Install Validation

- Confirm operators report `Succeeded` or `Ready` status.
- Validate GPU pods can schedule by launching a sample workload:
  ```bash
  oc apply -f samples/gpu-smoke-test.yaml
  ```
- Test RHOAI workbench and ModelMesh endpoints for end-to-end connectivity.
- Ensure Service Mesh enforces mTLS (`oc get peerauthentication -A`) and traffic policies align with zero-exfil requirements.

Keep manifests under `infra/openshift/manifests/` as they are refined. Update this plan as catalog channel versions change.

