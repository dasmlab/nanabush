#!/usr/bin/env bash
set -euo pipefail

# cycleme.sh - Build and push nanabush gRPC server image
# This script builds and pushes the image in one go

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[cycleme] Building nanabush gRPC server..."
./buildme.sh

echo "[cycleme] Pushing nanabush gRPC server..."
./pushme.sh

echo "[cycleme] ✅ Done!"

