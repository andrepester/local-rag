#!/bin/sh
set -eu

. ./shell/lib.sh

setup_go_toolchain_env

docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$(pwd):/workspace" -w /workspace "$GO_IMAGE" sh -lc "set -eu; out=\"\$($GOFMT_BIN -l .)\"; if [ -n \"\$out\" ]; then printf '%s\n' 'Go files are not formatted:' >&2; printf '%s\n' \"\$out\" >&2; exit 1; fi"
