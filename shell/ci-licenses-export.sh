#!/bin/sh
set -eu

. ./shell/lib.sh

setup_go_toolchain_env
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$(pwd):/workspace" -w /workspace "$GO_IMAGE" sh -lc 'set -eu; PATH="/usr/local/go/bin:$PATH"; toolbin=/tmp/bin; mkdir -p "$toolbin"; GOBIN="$toolbin" /usr/local/go/bin/go install github.com/google/go-licenses@v1.6.0; "$toolbin"/go-licenses report ./... > licenses.csv'
