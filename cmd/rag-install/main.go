package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrepester/rag-search-mcp/internal/bootstrap"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	var repoRoot string
	var interactive bool
	fs := flag.NewFlagSet("rag-install", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&repoRoot, "repo-root", ".", "repository root directory")
	fs.BoolVar(&interactive, "interactive", false, "prompt for docs/code source directories")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	created, err := bootstrap.EnsureEnvFile(absRoot)
	if err != nil {
		return err
	}
	if created {
		if _, err := fmt.Fprintln(stdout, "created .env from .env.example"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(stdout, "kept existing .env"); err != nil {
			return err
		}
	}

	if interactive {
		docsDir, codeDir, err := promptSourceDirs(stdin, stdout)
		if err != nil {
			return err
		}
		if err := bootstrap.UpsertHostSourceDirs(absRoot, docsDir, codeDir); err != nil {
			return err
		}
		if err := os.Setenv("HOST_DOCS_DIR", docsDir); err != nil {
			return fmt.Errorf("set HOST_DOCS_DIR for current run: %w", err)
		}
		if err := os.Setenv("HOST_CODE_DIR", codeDir); err != nil {
			return fmt.Errorf("set HOST_CODE_DIR for current run: %w", err)
		}
		if _, err := fmt.Fprintf(stdout, "updated .env host source paths: %s, %s\n", docsDir, codeDir); err != nil {
			return err
		}
	}

	port, err := bootstrap.ResolvePort(absRoot)
	if err != nil {
		return err
	}
	if err := bootstrap.UpsertOpenCodeConfig(absRoot, port); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "updated opencode.json for rag-search-mcp MCP alias at http://127.0.0.1:%d/mcp\n", port); err != nil {
		return err
	}

	if err := bootstrap.EnsureHostDataDirs(absRoot); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "ensured host mount directories for docs, code, index, and models"); err != nil {
		return err
	}

	return nil
}

func promptSourceDirs(stdin io.Reader, stdout io.Writer) (string, string, error) {
	reader := bufio.NewReader(stdin)
	if _, err := fmt.Fprint(stdout, "Use standard source directories (./data/docs and ./data/code)? [Y/n]: "); err != nil {
		return "", "", err
	}
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", "", fmt.Errorf("read standard source selection: %w", err)
	}
	selection := strings.ToLower(strings.TrimSpace(line))
	if selection == "" || selection == "y" || selection == "yes" {
		return "./data/docs", "./data/code", nil
	}
	if selection != "n" && selection != "no" {
		return "", "", fmt.Errorf("invalid selection %q, expected y/yes or n/no", strings.TrimSpace(line))
	}

	docsDir, err := promptRequiredPath(reader, stdout, "Enter HOST_DOCS_DIR: ")
	if err != nil {
		return "", "", err
	}
	codeDir, err := promptRequiredPath(reader, stdout, "Enter HOST_CODE_DIR: ")
	if err != nil {
		return "", "", err
	}

	return docsDir, codeDir, nil
}

func promptRequiredPath(reader *bufio.Reader, stdout io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(stdout, prompt); err != nil {
		return "", err
	}
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read input: %w", err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	return value, nil
}
