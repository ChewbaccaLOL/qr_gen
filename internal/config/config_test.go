package config

import (
	"os"
	"path/filepath"
	"testing"
)

func findRepoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := cwd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "variants.json")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate variants.json from %s", cwd)
	return ""
}

func TestLoadDefaultVariants(t *testing.T) {
	root := findRepoRoot(t)
	cfg, err := Load(filepath.Join(root, "variants.json"))
	if err != nil {
		t.Fatalf("load variants: %v", err)
	}
	if _, ok := cfg.Variants["classic"]; !ok {
		t.Fatalf("expected classic variant")
	}
	if len(cfg.AnimationVariants) == 0 {
		t.Fatalf("expected animation variants")
	}
}
