#!/bin/sh
set -eu

dockerfile_path=${DOCKERFILE_PATH:-docker/Dockerfile}
test_target=${DOCKER_TEST_TARGET:-test-runner}
test_image=${DOCKER_TEST_IMAGE:-rag-search-mcp-test-runner:local}

if [ "$#" -eq 0 ]; then
	set -- test -count=1 ./...
fi

docker build -f "$dockerfile_path" --target "$test_target" -t "$test_image" .
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$(pwd):/workspace" -w /workspace "$test_image" "$@"
