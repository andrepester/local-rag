#!/bin/sh
set -eu

. ./shell/lib.sh

host_repo=$(pwd -P)
host_parent=$(dirname "$host_repo")
repo_name=$(basename "$host_repo")

docs_value=$(resolve_host_override HOST_DOCS_DIR ./data/docs)
code_value=$(resolve_host_override HOST_CODE_DIR ./data/code)
persist_source_dirs=0

if [ ! -f .env ]; then
	cp .env.example .env
	chmod 600 .env
fi

if [ -t 0 ] && [ -t 1 ]; then
	persist_source_dirs=1
	printf '%s' 'Use standard source directories (./data/docs and ./data/code)? [Y/n]: '
	IFS= read -r selection || true
	selection_normalized=$(printf '%s' "$selection" | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')
	case "$selection_normalized" in
		''|y|yes)
			docs_value=./data/docs
			code_value=./data/code
			;;
		n|no)
			printf '%s' 'Enter HOST_DOCS_DIR: '
			IFS= read -r custom_docs
			if ! is_non_empty_non_ws "$custom_docs"; then
				printf '%s\n' 'HOST_DOCS_DIR must not be empty' >&2
				exit 2
			fi
			printf '%s' 'Enter HOST_CODE_DIR: '
			IFS= read -r custom_code
			if ! is_non_empty_non_ws "$custom_code"; then
				printf '%s\n' 'HOST_CODE_DIR must not be empty' >&2
				exit 2
			fi
			docs_value="$custom_docs"
			code_value="$custom_code"
			;;
		*)
			printf '%s\n' "invalid selection '$selection', expected y/yes or n/no" >&2
			exit 2
			;;
	esac
fi

if [ "$persist_source_dirs" -eq 1 ]; then
	upsert_env_value HOST_DOCS_DIR "$docs_value"
	upsert_env_value HOST_CODE_DIR "$code_value"
fi

set -- docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e RAG_HTTP_PORT -e HOST_DOCS_DIR -e HOST_CODE_DIR -e HOST_INDEX_DIR -e HOST_MODELS_DIR -v "$host_parent:/workspace-parent" -w "/workspace-parent/$repo_name"
for key in HOST_DOCS_DIR HOST_CODE_DIR HOST_INDEX_DIR HOST_MODELS_DIR; do
	resolved=$(resolve_host_override "$key" "")
	if [ -n "$resolved" ]; then
		resolved_abs=$(ensure_abs_dir "$host_repo" "$resolved")
		set -- "$@" -e "$key=$resolved_abs" -v "$resolved_abs:$resolved_abs"
	fi
done

"$@" "${GO_IMAGE:?GO_IMAGE is required}" "${GO_BIN:?GO_BIN is required}" run ./cmd/rag-install --repo-root "/workspace-parent/$repo_name"
