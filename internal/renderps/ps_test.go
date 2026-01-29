package renderps

import (
	"strings"
	"testing"

	"qr_generator/internal/config"
)

func TestRenderPSWritesHeader(t *testing.T) {
	matrix := [][]bool{
		{true, false},
		{false, true},
	}
	doc, err := RenderPS(matrix, 10, 1, "#000000", nil, "square", 0, nil, nil)
	if err != nil {
		t.Fatalf("RenderPS error: %v", err)
	}
	output := doc.buf.String()
	if !strings.HasPrefix(output, "%!PS-Adobe-3.0") {
		t.Fatalf("expected PS header, got: %q", output[:20])
	}
	if !strings.Contains(output, "%%BoundingBox:") {
		t.Fatalf("expected BoundingBox, got: %q", output[:80])
	}
}

func TestRenderCatalogPSIncludesLabels(t *testing.T) {
	matrix := [][]bool{
		{true, true},
		{true, false},
	}
	variants := []config.Variant{
		{Name: "classic", Dark: "#000000", Shape: "square"},
		{Name: "dotty", Dark: "#000000", Shape: "dot"},
	}
	doc, err := RenderCatalogPS(matrix, 8, 1, variants, 2, "#ffffff", 10)
	if err != nil {
		t.Fatalf("RenderCatalogPS error: %v", err)
	}
	output := doc.buf.String()
	if !strings.Contains(output, "classic") || !strings.Contains(output, "dotty") {
		t.Fatalf("expected catalog labels, got: %q", output[:120])
	}
}
