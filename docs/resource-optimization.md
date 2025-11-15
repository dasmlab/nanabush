# Resource Optimization for SNO Cluster

**Issue:** 7492m / 7500m CPU (99.9%) is already requested by platform components, leaving almost no capacity for workloads.

## Current CPU Request Breakdown

Based on analysis, the main CPU consumers are:

### High CPU Request Namespaces:
1. **redhat-ods-operator** - RHOAI operator components
2. **redhat-ods-applications** - RHOAI application components  
3. **knative-serving** - Serverless components
4. **openshift-kube-*** - Core OpenShift components

## Optimization Strategies

### 1. Reduce RHOAI Component Requests

Many RHOAI components have 500m CPU requests that may not be needed on a SNO cluster.

**Components to review:**
- `rhods-operator` pods (500m each)
- `rhods-dashboard` pods (500m each)
- `odh-notebook-controller-manager` (500m)
- `notebook-controller-deployment` (500m)
- `kueue-controller-manager` (500m)

**Recommended changes:**
- Reduce operator CPU requests from 500m to 100-200m
- Reduce dashboard CPU requests from 500m to 100m
- Reduce controller CPU requests from 500m to 100-200m

### 2. Reduce Knative Serving Requests

**Components to review:**
- `activator` pods (300m each)
- `controller` pods (100m each)
- `webhook` pods (100m each)
- `autoscaler-hpa` pods (30m each)

**Recommended changes:**
- Reduce activator CPU requests from 300m to 50-100m
- Reduce controller/webhook CPU requests from 100m to 50m

### 3. Review Pending/Failed Pods

Several pods are in Pending state, consuming CPU requests without actually running:
- Multiple `activator` pods (Pending)
- Multiple `controller` pods (Pending)
- Multiple `webhook` pods (Pending)

**Action:** Clean up pending pods that are stuck.

### 4. Use HorizontalPodAutoscaler with Lower Minimums

Instead of high static requests, use HPA with lower minimums and let it scale based on actual usage.

### 5. Priority Classes

Implement PriorityClasses to allow important workloads (like nanabush-grpc-server) to preempt lower priority pods when resources are constrained.

## Implementation

### Patch RHOAI Operator Deployment

```bash
CONTEXT="default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm"

# Patch rhods-operator deployment to reduce CPU requests
oc --context=$CONTEXT patch deployment -n redhat-ods-operator rhods-operator \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","resources":{"requests":{"cpu":"100m"}}}]}}}}'

# Patch rhods-dashboard deployment
oc --context=$CONTEXT patch deployment -n redhat-ods-applications rhods-dashboard \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"dashboard","resources":{"requests":{"cpu":"100m"}}}]}}}}'
```

### Patch Knative Serving

```bash
# Patch activator
oc --context=$CONTEXT patch deployment -n knative-serving activator \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"activator","resources":{"requests":{"cpu":"50m"}}}]}}}}'

# Patch controller
oc --context=$CONTEXT patch deployment -n knative-serving controller \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","resources":{"requests":{"cpu":"50m"}}}]}}}}'
```

### Clean Up Pending Pods

```bash
# Delete pending pods that are consuming resources
oc --context=$CONTEXT delete pods --all-namespaces --field-selector=status.phase=Pending
```

## Expected Results

After optimization:
- **Current:** 7492m / 7500m requested (99.9%)
- **Target:** ~4000-5000m / 7500m requested (53-67%)
- **Available:** ~2500-3500m for workloads

This will allow:
- Multiple nanabush-grpc-server replicas
- vLLM workloads
- Other AI/ML workloads
- Room for scaling

## Monitoring

After changes, monitor:
- Pod startup times (should not be significantly impacted)
- Actual CPU usage vs requests (ensure not over-committed)
- Scheduler pressure (should decrease)

## Notes

- These are resource **requests**, not limits. Limits can remain higher for burst capacity.
- On SNO clusters, components should be more conservative with requests.
- Some components may need higher requests - verify actual usage before reducing.

