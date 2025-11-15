#!/usr/bin/env bash
set -euo pipefail

# Quick deploy script for nanabush gRPC server to 1050ti cluster
# For development/testing - uses direct image reference

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTEXT="${OC_CONTEXT:-default/api-ocp-sno-1050ti-rh-dasmlab-org:6443/dasm}"

echo "[quick-deploy] Deploying nanabush gRPC server to 1050ti cluster..."

cd "${ROOT_DIR}"

# For local development, we'll use the built binary via initContainer or
# better: push image to a registry or build in cluster

# Option 1: If you have a registry accessible by both clusters:
# 1. docker tag nanabush-grpc-server:latest <registry>/nanabush-grpc-server:latest
# 2. docker push <registry>/nanabush-grpc-server:latest
# 3. Update deployment with registry URL

# Option 2: Build in cluster using BuildConfig (requires git repo)
# oc --context=$CONTEXT apply -f kustomize/base/grpc-server-buildconfig.yaml
# oc --context=$CONTEXT start-build nanabush-grpc-server -n nanabush

# Option 3: For testing, use a simple initContainer to copy binary from host
# (not recommended for production)

echo "[quick-deploy] Applying deployment manifests..."
oc --context="${CONTEXT}" apply -f kustomize/base/grpc-server-deployment.yaml

echo "[quick-deploy] Checking deployment status..."
oc --context="${CONTEXT}" get pods -n nanabush -l app=nanabush-grpc-server

echo "[quick-deploy] Note: Image may need to be pushed to a registry accessible by the cluster"
echo "[quick-deploy] Or use BuildConfig to build directly in the cluster"

