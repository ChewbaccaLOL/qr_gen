package renderpng

import (
	"image/color"
	"testing"

	"qr_generator/internal/config"
)

func TestRenderPNGDimensions(t *testing.T) {
	matrix := [][]bool{{true}}
	img, err := RenderPNG(matrix, 10, 4, "#000000", ptr("#ffffff"), "square", 0, nil, nil)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	if img.Bounds().Dx() != 90 || img.Bounds().Dy() != 90 {
		t.Fatalf("unexpected dimensions: %v", img.Bounds())
	}
}

func TestRenderPNGTransparentBackground(t *testing.T) {
	matrix := [][]bool{{true}}
	img, err := RenderPNG(matrix, 10, 4, "#000000", nil, "square", 0, nil, nil)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	if got := img.RGBAAt(0, 0); got.A != 0 {
		t.Fatalf("expected transparent background, got alpha %d", got.A)
	}
}

func TestGradientLUTVariation(t *testing.T) {
	matrix := [][]bool{{true}}
	gradient := &config.Gradient{From: "#000000", To: "#ffffff"}
	img, err := RenderPNG(matrix, 10, 0, "#000000", ptr("#ffffff"), "square", 0, gradient, nil)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	left := img.RGBAAt(0, 0)
	right := img.RGBAAt(9, 9)
	if left == right {
		t.Fatalf("expected gradient variation")
	}
}

func TestRenderPNGRotateTilesClamped(t *testing.T) {
	matrix := [][]bool{{true}}
	img, err := RenderPNGWithOffsets(matrix, 10, 4, "#000000", ptr("#ffffff"), "square", 0, nil, nil, nil, 0, 0, 20, "after", true)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	if img.Bounds().Dx() != 90 || img.Bounds().Dy() != 90 {
		t.Fatalf("unexpected dimensions: %v", img.Bounds())
	}
}

func TestRenderPNGRotateClamped(t *testing.T) {
	matrix := [][]bool{{true}}
	img, err := RenderPNGWithOffsets(matrix, 10, 4, "#000000", ptr("#ffffff"), "square", 0, nil, nil, nil, 0, 0, 20, "after", false)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	if img.Bounds().Dx() != 90 || img.Bounds().Dy() != 90 {
		t.Fatalf("unexpected dimensions: %v", img.Bounds())
	}
}

func TestRenderPNGRotateTilesKeepsBackground(t *testing.T) {
	matrix := [][]bool{{true}}
	bg := &config.Gradient{From: "#ffffff", To: "#ffffff"}
	img, err := RenderPNGWithOffsets(matrix, 10, 4, "#000000", nil, "square", 0, nil, bg, nil, 0, 0, 20, "after", true)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	got := img.RGBAAt(0, 0)
	if got.A == 0 {
		t.Fatalf("expected background to be opaque")
	}
	if got.R == 0 && got.G == 0 && got.B == 0 {
		t.Fatalf("expected background to stay visible, got black")
	}
}

func TestRenderPNGIslandGradient(t *testing.T) {
	matrix := [][]bool{{true, true}}
	gradient := &config.Gradient{From: "#000000", To: "#ffffff", Scope: "module"}
	img, err := RenderPNG(matrix, 10, 0, "#000000", ptr("#ffffff"), "island-4", 0.3, gradient, nil)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	left := img.RGBAAt(5, 5)
	right := img.RGBAAt(15, 5)
	if left == right {
		t.Fatalf("expected island gradient to span across modules")
	}
}

func TestParseColorHex(t *testing.T) {
	parsed, err := parseColor("#0f0")
	if err != nil {
		t.Fatalf("parse color: %v", err)
	}
	if parsed != (color.RGBA{0, 255, 0, 255}) {
		t.Fatalf("unexpected color: %v", parsed)
	}
}

func ptr(value string) *string {
	return &value
}
