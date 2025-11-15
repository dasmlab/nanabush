#!/usr/bin/env bash
set -euo pipefail

# Optimize CPU requests on SNO cluster to free up resources for workloads
# This reduces CPU requests from platform components that are over-provisioned

CONTEXT="${OC_CONTEXT:-default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm}"

echo "[optimize-resources] Optimizing CPU requests on SNO cluster..."
echo "Context: ${CONTEXT}"

# Function to patch deployment CPU requests
patch_cpu() {
    local namespace=$1
    local deployment=$2
    local container=$3
    local new_cpu=$4
    
    echo "[optimize-resources] Patching ${namespace}/${deployment} (${container}): ${new_cpu}"
    
    # Use oc set resources which is simpler and more reliable
    if oc --context="${CONTEXT}" set resources deployment -n "${namespace}" "${deployment}" -c "${container}" --requests="cpu=${new_cpu}" &>/dev/null; then
        echo "  ✅ Successfully patched ${namespace}/${deployment}"
    else
        echo "  ⚠️  Failed to patch ${namespace}/${deployment}"
    fi
}

# RHOAI Operator - Reduce from 500m to 200m
echo ""
echo "[optimize-resources] Optimizing RHOAI Operator..."
patch_cpu "redhat-ods-operator" "rhods-operator" "rhods-operator" "200m"

# RHOAI Applications - Reduce from 500m to 100m
echo ""
echo "[optimize-resources] Optimizing RHOAI Applications..."
patch_cpu "redhat-ods-applications" "rhods-dashboard" "rhods-dashboard" "100m"
patch_cpu "redhat-ods-applications" "odh-notebook-controller-manager" "manager" "100m"
patch_cpu "redhat-ods-applications" "notebook-controller-deployment" "manager" "100m"
patch_cpu "redhat-ods-applications" "kueue-controller-manager" "manager" "100m"

# Knative Serving - Reduce from 300m/100m to 50m
echo ""
echo "[optimize-resources] Optimizing Knative Serving..."
patch_cpu "knative-serving" "activator" "activator" "50m"
patch_cpu "knative-serving" "controller" "controller" "50m"
patch_cpu "knative-serving" "webhook" "webhook" "50m"
patch_cpu "knative-serving" "autoscaler-hpa" "autoscaler-hpa" "20m"

# Clean up pending pods that are stuck
echo ""
echo "[optimize-resources] Cleaning up pending pods..."
oc --context="${CONTEXT}" delete pods --all-namespaces --field-selector=status.phase=Pending --timeout=30s 2>&1 || echo "  ⚠️  Some pending pods couldn't be deleted"

# Wait for deployments to stabilize
echo ""
echo "[optimize-resources] Waiting for deployments to stabilize..."
sleep 10

# Show new resource usage
echo ""
echo "[optimize-resources] Current CPU requests by namespace:"
oc --context="${CONTEXT}" get pods --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}{"\t"}{.spec.containers[0].resources.requests.cpu}{"\n"}{end}' | \
    awk -F'\t' '{if ($2 != "") {cpu[$1]+=$2}} END {for (ns in cpu) printf "  %-40s %10s\n", ns, cpu[ns]"m"}' | \
    sort -k2 -rn | head -10

echo ""
echo "[optimize-resources] ✅ Optimization complete!"
echo "[optimize-resources] Monitor pod startup times and actual CPU usage to ensure components still function properly."

