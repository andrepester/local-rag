package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreatesEnvAndOpenCodeConfig(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, ".env.example"), []byte("RAG_HTTP_PORT=8090\n"), 0o644); err != nil {
		t.Fatalf("write .env.example: %v", err)
	}

	var out bytes.Buffer
	if err := run([]string{"--repo-root", repoRoot}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoRoot, ".env")); err != nil {
		t.Fatalf("expected .env to be created: %v", err)
	}
	assertFileMode(t, filepath.Join(repoRoot, ".env"), 0o600)
	raw, err := os.ReadFile(filepath.Join(repoRoot, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	assertFileMode(t, filepath.Join(repoRoot, "opencode.json"), 0o600)
	if !strings.Contains(string(raw), "http://127.0.0.1:8090/mcp") {
		t.Fatalf("opencode.json did not include expected URL: %s", string(raw))
	}

	for _, dir := range []string{
		filepath.Join(repoRoot, "data", "docs"),
		filepath.Join(repoRoot, "data", "code"),
		filepath.Join(repoRoot, "data", "index"),
		filepath.Join(repoRoot, "data", "models"),
	} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist: %v", dir, err)
		}
	}
}

func TestRunInteractiveSetsCustomSourceDirsInEnv(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, ".env.example"), []byte("RAG_HTTP_PORT=8090\nHOST_DOCS_DIR=./data/docs\nHOST_CODE_DIR=./data/code\n"), 0o644); err != nil {
		t.Fatalf("write .env.example: %v", err)
	}

	var out bytes.Buffer
	input := strings.NewReader("n\n./custom/docs\n./custom/code\n")
	if err := run([]string{"--repo-root", repoRoot, "--interactive"}, input, &out); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(repoRoot, ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	env := string(raw)
	if !strings.Contains(env, "HOST_DOCS_DIR=./custom/docs") {
		t.Fatalf("expected HOST_DOCS_DIR override in .env: %s", env)
	}
	if !strings.Contains(env, "HOST_CODE_DIR=./custom/code") {
		t.Fatalf("expected HOST_CODE_DIR override in .env: %s", env)
	}

	for _, dir := range []string{
		filepath.Join(repoRoot, "custom", "docs"),
		filepath.Join(repoRoot, "custom", "code"),
	} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist: %v", dir, err)
		}
	}
}

func TestRunInteractiveResetsToStandardSourceDirs(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, ".env.example"), []byte("RAG_HTTP_PORT=8090\nHOST_DOCS_DIR=./data/docs\nHOST_CODE_DIR=./data/code\n"), 0o644); err != nil {
		t.Fatalf("write .env.example: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte("RAG_HTTP_PORT=8090\nHOST_DOCS_DIR=./custom/docs\nHOST_CODE_DIR=./custom/code\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	var out bytes.Buffer
	if err := run([]string{"--repo-root", repoRoot, "--interactive"}, strings.NewReader("y\n"), &out); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(repoRoot, ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	env := string(raw)
	if !strings.Contains(env, "HOST_DOCS_DIR=./data/docs") {
		t.Fatalf("expected HOST_DOCS_DIR default in .env: %s", env)
	}
	if !strings.Contains(env, "HOST_CODE_DIR=./data/code") {
		t.Fatalf("expected HOST_CODE_DIR default in .env: %s", env)
	}
}

func TestRunInteractiveOverridesExistingProcessEnvForCurrentRun(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, ".env.example"), []byte("RAG_HTTP_PORT=8090\nHOST_DOCS_DIR=./data/docs\nHOST_CODE_DIR=./data/code\n"), 0o644); err != nil {
		t.Fatalf("write .env.example: %v", err)
	}

	t.Setenv("HOST_DOCS_DIR", "./stale/docs")
	t.Setenv("HOST_CODE_DIR", "./stale/code")

	var out bytes.Buffer
	input := strings.NewReader("n\n./fresh/docs\n./fresh/code\n")
	if err := run([]string{"--repo-root", repoRoot, "--interactive"}, input, &out); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	for _, dir := range []string{
		filepath.Join(repoRoot, "fresh", "docs"),
		filepath.Join(repoRoot, "fresh", "code"),
	} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist: %v", dir, err)
		}
	}

	if _, err := os.Stat(filepath.Join(repoRoot, "stale", "docs")); !os.IsNotExist(err) {
		t.Fatalf("expected stale docs directory to remain absent, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "stale", "code")); !os.IsNotExist(err) {
		t.Fatalf("expected stale code directory to remain absent, got err=%v", err)
	}
}

func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("permissions for %s = %o, want %o", path, got, want)
	}
}
