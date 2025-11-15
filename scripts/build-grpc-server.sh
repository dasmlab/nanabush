#!/usr/bin/env bash
set -euo pipefail

# Build script for nanabush gRPC server container image

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-nanabush-grpc-server}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
REGISTRY="${REGISTRY:-registry.redhat.io/rhoai}"

echo "[build-grpc-server] Building nanabush gRPC server container image..."
echo "Image: ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

cd "${ROOT_DIR}"

# Build using Dockerfile
docker build \
  -f kustomize/base/Dockerfile.grpc-server \
  -t "${IMAGE_NAME}:${IMAGE_TAG}" \
  -t "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" \
  .

echo "[build-grpc-server] Image built successfully!"
echo "To push to registry:"
echo "  docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

