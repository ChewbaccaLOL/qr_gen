package render

import (
	"os"
	"path/filepath"
	"strings"
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
			variant.BackgroundGradient,
			false,
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

func TestRenderSVGCutoutMask(t *testing.T) {
	matrix := [][]bool{
		{true, false},
		{false, true},
	}
	light := "#ffffff"
	svg, err := RenderSVG(
		matrix,
		4,
		1,
		"#000000",
		&light,
		"square",
		0,
		nil,
		nil,
		true,
		nil,
		0,
		0,
		0,
		"after",
	)
	if err != nil {
		t.Fatalf("render svg cutout: %v", err)
	}
	if !strings.Contains(svg, "<mask id=\"qr-cutout\"") {
		t.Fatalf("expected cutout mask definition")
	}
	if !strings.Contains(svg, "mask=\"url(#qr-cutout)\"") {
		t.Fatalf("expected background rect to use cutout mask")
	}
}

func TestRenderSVGCutoutRequiresBackground(t *testing.T) {
	matrix := [][]bool{{true}}
	_, err := RenderSVG(
		matrix,
		4,
		1,
		"#000000",
		nil,
		"square",
		0,
		nil,
		nil,
		true,
		nil,
		0,
		0,
		0,
		"after",
	)
	if err == nil {
		t.Fatalf("expected error when cutout has no background")
	}
}

// Diff helper lives in compat_test.go to reuse across tests.
