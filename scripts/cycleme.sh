#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NS="${1:-nanabush}"

echo "[cycleme] Rebuilding chart..."
"${ROOT_DIR}/scripts/buildme.sh"

echo "[cycleme] Redeploying release..."
"${ROOT_DIR}/scripts/deployme.sh" "${NS}"

echo "[cycleme] Triggering Tekton translation pipeline dry-run..."
kubectl create -n "${NS}" -f "${ROOT_DIR}/tekton/translation-pipeline.yaml" --dry-run=client -o yaml | kubectl apply -n "${NS}" -f -

echo "[cycleme] Done."

