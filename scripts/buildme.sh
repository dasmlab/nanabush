#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHART_DIR="${ROOT_DIR}/helm/vllm"

echo "[buildme] Packaging vLLM Helm chart..."
rm -f "${ROOT_DIR}"/dist/*.tgz 2>/dev/null || true
mkdir -p "${ROOT_DIR}/dist"

helm lint "${CHART_DIR}"
helm package "${CHART_DIR}" --destination "${ROOT_DIR}/dist"

echo "[buildme] Chart packaged under dist/"

