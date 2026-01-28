package render

import (
	"os"
	"path/filepath"
	"testing"

	"qr_generator/internal/config"
	"qr_generator/internal/qr"
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

func TestRenderSVGMatchesGolden(t *testing.T) {
	root := findRepoRoot(t)
	cfg, err := config.Load(filepath.Join(root, "variants.json"))
	if err != nil {
		t.Fatalf("load variants: %v", err)
	}

	matrix, err := qr.EncodeMatrix("https://example.com", "m")
	if err != nil {
		t.Fatalf("encode matrix: %v", err)
	}

	cases := []string{"classic", "rounded", "sunset"}
	update := os.Getenv("UPDATE_GOLDEN") != ""

	for _, name := range cases {
		variant, ok := cfg.Variants[name]
		if !ok {
			t.Fatalf("missing variant %s", name)
		}
		svg, err := RenderSVG(
			matrix,
			10,
			4,
			variant.Dark,
			variant.Light,
			variant.Shape,
			variant.Radius,
			variant.Gradient,
			nil,
			0,
			0,
			0,
			"after",
		)
		if err != nil {
			t.Fatalf("render svg %s: %v", name, err)
		}

		goldenPath := filepath.Join(root, "internal", "render", "testdata", name+".svg")
		if update {
			if err := os.WriteFile(goldenPath, []byte(svg), 0o644); err != nil {
				t.Fatalf("write golden %s: %v", name, err)
			}
			continue
		}

		golden, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if string(golden) != svg {
			t.Fatalf("svg mismatch for %s:\n%s", name, firstDiff(string(golden), svg))
		}
	}
}

// Diff helper lives in compat_test.go to reuse across tests.
