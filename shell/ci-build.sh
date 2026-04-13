#!/bin/sh
set -eu

. ./shell/lib.sh

setup_go_toolchain_env

docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$(pwd):/workspace" -w /workspace "$GO_IMAGE" "$GO_BIN" build ./...
