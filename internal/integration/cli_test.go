package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var cliPath string

func TestMain(m *testing.M) {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tmpDir, err := os.MkdirTemp("", "qr-generator-cli-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "qr-generator"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	cliPath = filepath.Join(tmpDir, binaryName)

	cmd := exec.Command("go", "build", "-o", cliPath, "./cmd/qr_generator")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build cli: %v\n%s\n", err, string(output))
		os.Exit(1)
	}

	exit := m.Run()
	os.Exit(exit)
}

func TestCLIGeneratesSVG(t *testing.T) {
	root := mustRepoRoot(t)
	outDir := t.TempDir()
	svgPath := filepath.Join(outDir, "qr.svg")

	cmd := exec.Command(cliPath, "-o", svgPath, "https://example.com")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run cli: %v\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "Saved") {
		t.Fatalf("expected saved message, got: %s", string(output))
	}
	data, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("read svg: %v", err)
	}
	trimmed := bytes.TrimSpace(data)
	if !bytes.HasPrefix(trimmed, []byte("<svg")) && !bytes.HasPrefix(trimmed, []byte("<?xml")) {
		t.Fatalf("expected svg output")
	}
}

func TestCLIGeneratesPNG(t *testing.T) {
	root := mustRepoRoot(t)
	outDir := t.TempDir()
	svgPath := filepath.Join(outDir, "qr.svg")
	pngPath := filepath.Join(outDir, "qr.png")

	cmd := exec.Command(cliPath, "--png", "--png-output", pngPath, "-o", svgPath, "Hello PNG")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run cli: %v\n%s", err, string(output))
	}
	data, err := os.ReadFile(pngPath)
	if err != nil {
		t.Fatalf("read png: %v", err)
	}
	if len(data) < 8 || !bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
		t.Fatalf("expected png signature")
	}
}

func mustRepoRoot(t *testing.T) string {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	dir := cwd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "variants.json")
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not locate variants.json from %s", cwd)
}
