package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qr_generator/internal/config"
)

func TestListVariantsSorted(t *testing.T) {
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"zeta": {
				Name:  "zeta",
				Shape: "square",
				Dark:  "#000000",
			},
			"alpha": {
				Name:  "alpha",
				Shape: "square",
				Dark:  "#111111",
			},
		},
		AnimationVariants: []string{"wave"},
	}
	variants, err := ListVariants(cfg)
	if err != nil {
		t.Fatalf("list variants: %v", err)
	}
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(variants))
	}
	if variants[0].Name != "alpha" || variants[1].Name != "zeta" {
		t.Fatalf("expected sorted variants, got %v then %v", variants[0].Name, variants[1].Name)
	}
}

func TestBuildSVGUsesOverrides(t *testing.T) {
	light := "#ffffff"
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:   "classic",
				Shape:  "square",
				Dark:   "#000000",
				Light:  &light,
				Radius: 0,
			},
		},
		AnimationVariants: []string{"wave"},
	}
	req := RenderRequest{
		Data:         "hello",
		Variant:      "classic",
		ErrorLevel:   "m",
		Scale:        6,
		Border:       4,
		Dark:         "#ff0000",
		NoBackground: true,
	}
	svg, err := BuildSVG(cfg, req)
	if err != nil {
		t.Fatalf("build svg: %v", err)
	}
	if !strings.Contains(svg, "#ff0000") {
		t.Fatalf("expected override dark color")
	}
	if strings.Contains(svg, light) {
		t.Fatalf("expected no background color")
	}
}

func TestFindVariantsPath(t *testing.T) {
	root := t.TempDir()
	variantsPath := filepath.Join(root, "variants.json")
	if err := os.WriteFile(variantsPath, []byte(`{"variants":[],"animation_variants":[]}`), 0o644); err != nil {
		t.Fatalf("write variants: %v", err)
	}
	guiDir := filepath.Join(root, "gui")
	if err := os.MkdirAll(guiDir, 0o755); err != nil {
		t.Fatalf("mkdir gui: %v", err)
	}
	found, err := FindVariantsPath(guiDir, "")
	if err != nil {
		t.Fatalf("find variants: %v", err)
	}
	if found != variantsPath {
		t.Fatalf("expected %s, got %s", variantsPath, found)
	}
}
