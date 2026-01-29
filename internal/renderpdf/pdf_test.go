package renderpdf

import (
	"bytes"
	"strings"
	"testing"

	"qr_generator/internal/config"
)

func TestRenderPDFWritesHeader(t *testing.T) {
	matrix := [][]bool{
		{true, false},
		{false, true},
	}
	doc, err := RenderPDF(matrix, 10, 1, "#000000", nil, "square", 0, nil, nil)
	if err != nil {
		t.Fatalf("RenderPDF error: %v", err)
	}
	buf := bytes.NewBuffer(nil)
	if err := doc.pdf.Output(buf); err != nil {
		t.Fatalf("Output error: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "%PDF") {
		t.Fatalf("expected PDF header, got: %q", buf.String()[:10])
	}
}

func TestRenderCatalogPDFSmoke(t *testing.T) {
	matrix := [][]bool{
		{true, true},
		{true, false},
	}
	variants := []config.Variant{
		{Name: "classic", Dark: "#000000", Shape: "square"},
		{Name: "dotty", Dark: "#000000", Shape: "dot"},
	}
	doc, err := RenderCatalogPDF(matrix, 8, 1, variants, 2, "#ffffff", 10)
	if err != nil {
		t.Fatalf("RenderCatalogPDF error: %v", err)
	}
	buf := bytes.NewBuffer(nil)
	if err := doc.pdf.Output(buf); err != nil {
		t.Fatalf("Output error: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "%PDF") {
		t.Fatalf("expected PDF header, got: %q", buf.String()[:10])
	}
}
