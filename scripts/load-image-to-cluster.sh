#!/usr/bin/env bash
set -euo pipefail

# Load nanabush gRPC server image to OpenShift cluster
# For local development - loads image directly to cluster

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="nanabush-grpc-server:latest"
CONTEXT="${OC_CONTEXT:-default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm}"

echo "[load-image] Loading ${IMAGE_NAME} to cluster (${CONTEXT})..."

# Save image to tar
echo "[load-image] Saving image to tar..."
docker save "${IMAGE_NAME}" -o /tmp/nanabush-grpc-server.tar

# Load image to cluster
echo "[load-image] Loading image to cluster..."
oc --context="${CONTEXT}" image import nanabush-grpc-server:latest --from=/tmp/nanabush-grpc-server.tar -n nanabush 2>&1 || {
    echo "[load-image] Image import failed, trying alternative method..."
    # Alternative: Use image stream and reference
    oc --context="${CONTEXT}" import-image nanabush-grpc-server:latest -n nanabush --from=docker-daemon:${IMAGE_NAME} --confirm 2>&1 || true
}

# Clean up
rm -f /tmp/nanabush-grpc-server.tar

echo "[load-image] Image loaded to cluster!"
echo "[load-image] Update deployment to use: imageStreamTag/nanabush-grpc-server:latest"

