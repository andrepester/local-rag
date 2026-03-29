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

	originalArgs := os.Args
	os.Args = []string{"rag-install", "--repo-root", repoRoot}
	t.Cleanup(func() { os.Args = originalArgs })

	var out bytes.Buffer
	if err := run(&out); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoRoot, ".env")); err != nil {
		t.Fatalf("expected .env to be created: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(repoRoot, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	if !strings.Contains(string(raw), "http://127.0.0.1:8090/mcp") {
		t.Fatalf("opencode.json did not include expected URL: %s", string(raw))
	}
}
