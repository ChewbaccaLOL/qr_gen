package gui

import (
	"os"
	"path/filepath"
	"testing"

	"qr_generator/internal/config"
)

func TestCustomVariantsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "variants.custom.json")
	light := "#ffffff"
	input := map[string]config.Variant{
		"custom-alpha": {
			Name:   "custom-alpha",
			Shape:  "square",
			Dark:   "#111111",
			Light:  &light,
			Radius: 0.2,
		},
	}
	if err := SaveCustomVariants(path, input); err != nil {
		t.Fatalf("save custom variants: %v", err)
	}
	loaded, err := LoadCustomVariants(path)
	if err != nil {
		t.Fatalf("load custom variants: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(loaded))
	}
	if _, ok := loaded["custom-alpha"]; !ok {
		t.Fatalf("expected custom-alpha variant")
	}
}

func TestCustomVariantsMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")
	loaded, err := LoadCustomVariants(path)
	if err != nil {
		t.Fatalf("load custom variants: %v", err)
	}
	if len(loaded) != 0 {
		t.Fatalf("expected no variants, got %d", len(loaded))
	}
}

func TestSaveCustomVariantsEmptyRemovesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "variants.custom.json")
	if err := SaveCustomVariants(path, map[string]config.Variant{}); err != nil {
		t.Fatalf("save custom variants empty: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected custom variants file to be removed")
	}
}

func TestMergeVariantsRejectsDuplicate(t *testing.T) {
	base := map[string]config.Variant{
		"classic": {
			Name:  "classic",
			Shape: "square",
			Dark:  "#000000",
		},
	}
	custom := map[string]config.Variant{
		"classic": {
			Name:  "classic",
			Shape: "square",
			Dark:  "#111111",
		},
	}
	if _, err := MergeVariants(base, custom); err == nil {
		t.Fatalf("expected duplicate variants error")
	}
}
