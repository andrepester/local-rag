#!/bin/sh
set -eu

. ./shell/lib.sh

compose_project_dir=${COMPOSE_PROJECT_DIR:-.}
compose_file=${COMPOSE_FILE:-docker/docker-compose.yml}

docker compose --project-directory "$compose_project_dir" -f "$compose_file" config >/dev/null

if ! docker compose --project-directory "$compose_project_dir" -f "$compose_file" exec -T rag-mcp true >/dev/null 2>&1; then
	printf '%s\n' 'doctor: rag-mcp container is not running. Start the stack first with make up.' >&2
	exit 1
fi

if ! docker compose --project-directory "$compose_project_dir" -f "$compose_file" exec -T chroma true >/dev/null 2>&1; then
	printf '%s\n' 'doctor: chroma container is not running. Start the stack first with make up.' >&2
	exit 1
fi

if ! docker compose --project-directory "$compose_project_dir" -f "$compose_file" exec -T ollama true >/dev/null 2>&1; then
	printf '%s\n' 'doctor: ollama container is not running. Start the stack first with make up.' >&2
	exit 1
fi

docker compose --project-directory "$compose_project_dir" -f "$compose_file" exec -T rag-mcp /app/rag-index
COMPOSE_PROJECT_DIR="$compose_project_dir" COMPOSE_FILE="$compose_file" sh ./shell/doctor-verify-index.sh

rag_http_port=$(resolve_host_override RAG_HTTP_PORT 8765)
if ! curl -fsS "http://127.0.0.1:${rag_http_port}/healthz" >/dev/null; then
	printf '%s\n' 'doctor: MCP health endpoint is not ready on localhost' >&2
	exit 1
fi

printf '%s\n' 'doctor: runtime checks passed'
