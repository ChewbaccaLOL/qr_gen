package renderpng

import (
	"testing"

	"qr_generator/internal/config"
)

func TestRenderCatalogPNGDimensions(t *testing.T) {
	matrix := [][]bool{{true}}
	variants := []config.Variant{{
		Name:  "classic",
		Shape: "square",
		Dark:  "#000000",
		Light: ptr("#ffffff"),
	}}
	scale := 10
	border := 4
	columns := 2
	labelSize := 0
	background := "#ffffff"

	img, err := RenderCatalogPNG(matrix, scale, border, variants, columns, background, labelSize)
	if err != nil {
		t.Fatalf("render catalog png: %v", err)
	}

	tileDim := (len(matrix) + border*2) * scale
	padding := maxInt(8, int(float64(scale)*1.2))
	expectedLabelSize := maxInt(10, int(float64(scale)*1.4))
	labelHeight := int(float64(expectedLabelSize) * 1.6)
	tileTotalHeight := tileDim + labelHeight + padding
	width := columns*tileDim + (columns+1)*padding
	height := tileTotalHeight + padding

	if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
		t.Fatalf("unexpected dimensions: %v", img.Bounds())
	}
}

func TestRenderCatalogPNGLabelDrawn(t *testing.T) {
	matrix := [][]bool{{true}}
	variants := []config.Variant{{
		Name:  "classic",
		Shape: "square",
		Dark:  "#000000",
		Light: ptr("#ffffff"),
	}}
	scale := 10
	border := 4
	columns := 1
	labelSize := 12
	background := "#ffffff"

	img, err := RenderCatalogPNG(matrix, scale, border, variants, columns, background, labelSize)
	if err != nil {
		t.Fatalf("render catalog png: %v", err)
	}

	tileDim := (len(matrix) + border*2) * scale
	padding := maxInt(8, int(float64(scale)*1.2))
	labelHeight := int(float64(labelSize) * 1.6)
	originX := padding
	originY := padding
	labelX := originX + tileDim/2
	labelY := originY + tileDim + int(float64(labelHeight)*0.75)

	bg, err := parseColor(background)
	if err != nil {
		t.Fatalf("parse background: %v", err)
	}

	found := false
	for dy := -2; dy <= 2; dy++ {
		for dx := -20; dx <= 20; dx++ {
			x := labelX + dx
			y := labelY + dy
			if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
				continue
			}
			if img.RGBAAt(x, y) != bg {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Fatalf("expected label pixels to differ from background")
	}
}
