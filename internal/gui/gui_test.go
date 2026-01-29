package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qr_generator/internal/config"
)

func TestListVariantsSorted(t *testing.T) {
	base := map[string]config.Variant{
		"zeta": {
			Name:  "zeta",
			Shape: "square",
			Dark:  "#000000",
		},
		"beta": {
			Name:     "beta",
			Shape:    "square",
			Dark:     "#111111",
			Disabled: true,
		},
		"alpha": {
			Name:  "alpha",
			Shape: "square",
			Dark:  "#111111",
		},
	}
	variants, err := ListVariants(base, nil)
	if err != nil {
		t.Fatalf("list variants: %v", err)
	}
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(variants))
	}
	if variants[0].Name != "alpha" || variants[1].Name != "zeta" {
		t.Fatalf("expected sorted variants, got %v then %v", variants[0].Name, variants[1].Name)
	}
	if variants[0].IsCustom || variants[1].IsCustom {
		t.Fatalf("expected base variants to not be custom")
	}
}

func TestListVariantsWithCustom(t *testing.T) {
	base := map[string]config.Variant{
		"classic": {
			Name:  "classic",
			Shape: "square",
			Dark:  "#000000",
		},
	}
	custom := map[string]config.Variant{
		"custom": {
			Name:  "custom",
			Shape: "square",
			Dark:  "#111111",
		},
	}
	variants, err := ListVariants(base, custom)
	if err != nil {
		t.Fatalf("list variants: %v", err)
	}
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(variants))
	}
	if variants[0].Name != "classic" || variants[1].Name != "custom" {
		t.Fatalf("unexpected variant order")
	}
	if variants[1].IsCustom != true {
		t.Fatalf("expected custom variant to be flagged")
	}
}

func TestListVariantsIncludesCutout(t *testing.T) {
	base := map[string]config.Variant{
		"cutout": {
			Name:   "cutout",
			Shape:  "square",
			Dark:   "#000000",
			Cutout: true,
		},
	}
	variants, err := ListVariants(base, nil)
	if err != nil {
		t.Fatalf("list variants: %v", err)
	}
	if len(variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(variants))
	}
	if !variants[0].Cutout {
		t.Fatalf("expected cutout flag to be true")
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

func TestBuildSVGDefaults(t *testing.T) {
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#123456",
			},
		},
		AnimationVariants: []string{"wave"},
	}
	req := RenderRequest{
		Data:   "hello",
		Scale:  6,
		Border: 4,
	}
	svg, err := BuildSVG(cfg, req)
	if err != nil {
		t.Fatalf("build svg: %v", err)
	}
	if !strings.Contains(svg, "#123456") {
		t.Fatalf("expected default variant color")
	}
}

func TestBuildSVGCutoutRequiresBackground(t *testing.T) {
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#123456",
			},
		},
		AnimationVariants: []string{"wave"},
	}
	req := RenderRequest{
		Data:    "hello",
		Variant: "classic",
		Scale:   6,
		Border:  4,
		Cutout:  true,
	}
	_, err := BuildSVG(cfg, req)
	if err == nil {
		t.Fatalf("expected error for cutout without background")
	}
}

func TestBuildPNGDefaultScale(t *testing.T) {
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#000000",
			},
		},
		AnimationVariants: []string{"wave"},
	}
	req := RenderRequest{
		Data:     "hello",
		Variant:  "classic",
		Scale:    6,
		Border:   4,
		PngScale: 0,
	}
	imageOut, err := BuildPNG(cfg, req)
	if err != nil {
		t.Fatalf("build png: %v", err)
	}
	if imageOut == nil {
		t.Fatalf("expected png image output")
	}
}

func TestBuildSVGGradientOverrides(t *testing.T) {
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#000000",
				Gradient: &config.Gradient{
					ID:   "fg",
					From: "#111111",
					To:   "#222222",
				},
			},
		},
		AnimationVariants: []string{"wave"},
	}
	req := RenderRequest{
		Data:       "hello",
		Variant:    "classic",
		ErrorLevel: "m",
		Scale:      6,
		Border:     4,
		Gradient:   true,
	}
	svg, err := BuildSVG(cfg, req)
	if err != nil {
		t.Fatalf("build svg: %v", err)
	}
	if !strings.Contains(svg, "linearGradient") {
		t.Fatalf("expected gradient definition")
	}
	if !strings.Contains(svg, "#111111") || !strings.Contains(svg, "#222222") {
		t.Fatalf("expected default gradient colors")
	}

	req.GradientFrom = "#abcdef"
	req.GradientTo = "#123456"
	svg, err = BuildSVG(cfg, req)
	if err != nil {
		t.Fatalf("build svg: %v", err)
	}
	if !strings.Contains(svg, "#abcdef") || !strings.Contains(svg, "#123456") {
		t.Fatalf("expected override gradient colors")
	}
}

func TestBuildSVGShapeOverride(t *testing.T) {
	cfg := &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#000000",
			},
		},
		AnimationVariants: []string{"wave"},
	}
	req := RenderRequest{
		Data:       "hello",
		Variant:    "classic",
		ErrorLevel: "m",
		Scale:      6,
		Border:     4,
		Shape:      "dot",
	}
	svg, err := BuildSVG(cfg, req)
	if err != nil {
		t.Fatalf("build svg: %v", err)
	}
	if !strings.Contains(svg, "circle") {
		t.Fatalf("expected dot shape override")
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

func TestFindVariantsPathUsesExeAncestors(t *testing.T) {
	root := t.TempDir()
	variantsPath := filepath.Join(root, "variants.json")
	if err := os.WriteFile(variantsPath, []byte(`{"variants":[],"animation_variants":[]}`), 0o644); err != nil {
		t.Fatalf("write variants: %v", err)
	}
	exeDir := filepath.Join(root, "gui", "build", "bin")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatalf("mkdir exe dir: %v", err)
	}
	exePath := filepath.Join(exeDir, "qr-generator")
	found, err := FindVariantsPath(exeDir, exePath)
	if err != nil {
		t.Fatalf("find variants: %v", err)
	}
	if found != variantsPath {
		t.Fatalf("expected %s, got %s", variantsPath, found)
	}
}

func TestDefaultExportDir(t *testing.T) {
	root := t.TempDir()
	expected := filepath.Join(root, "out")
	if got := DefaultExportDir(root); got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	if got := DefaultExportDir(""); got != "" {
		t.Fatalf("expected empty default dir, got %s", got)
	}
}
