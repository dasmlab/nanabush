#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHART="${ROOT_DIR}/dist/vllm-0.1.0.tgz"
NS="${1:-nanabush}"

if ! helm status vllm -n "${NS}" >/dev/null 2>&1; then
  helm upgrade --install vllm "${CHART}" \
    --namespace "${NS}" \
    --create-namespace
else
  helm upgrade vllm "${CHART}" -n "${NS}"
fi

echo "[deployme] vLLM release applied to namespace ${NS}"

