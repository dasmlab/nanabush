#!/usr/bin/env bash
set -euo pipefail

app=nanabush-grpc-server
version=scratch
docker build -t  $app:$version .

