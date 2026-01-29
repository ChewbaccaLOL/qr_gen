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
