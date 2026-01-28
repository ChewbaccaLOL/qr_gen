package renderpng

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"qr_generator/internal/config"
)

type Gradient struct {
	From color.RGBA
	To   color.RGBA
}

func RenderPNG(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
) (*image.RGBA, error) {
	if scale <= 0 {
		return nil, errors.New("scale must be greater than 0")
	}
	size := len(matrix)
	if size == 0 {
		return nil, errors.New("matrix is empty")
	}
	dim := (size + border*2) * scale
	if dim <= 0 {
		return nil, errors.New("image dimensions must be positive")
	}

	darkColor, err := parseColor(dark)
	if err != nil {
		return nil, fmt.Errorf("invalid dark color: %w", err)
	}

	var lightColor *color.RGBA
	if light != nil {
		parsed, err := parseColor(*light)
		if err != nil {
			return nil, fmt.Errorf("invalid light color: %w", err)
		}
		lightColor = &parsed
	}

	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	if lightColor != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{C: *lightColor}, image.Point{}, draw.Src)
	}

	var gradientDef *Gradient
	if gradient != nil {
		from, err := parseColor(gradient.From)
		if err != nil {
			return nil, fmt.Errorf("invalid gradient from color: %w", err)
		}
		to, err := parseColor(gradient.To)
		if err != nil {
			return nil, fmt.Errorf("invalid gradient to color: %w", err)
		}
		gradientDef = &Gradient{From: from, To: to}
	}

	var gradientLUT []color.RGBA
	if gradientDef != nil {
		gradientLUT = buildGradientLUT(scale, gradientDef.From, gradientDef.To)
	}

	for y, row := range matrix {
		for x, cell := range row {
			if !cell {
				continue
			}
			originX := (x + border) * scale
			originY := (y + border) * scale
			if err := drawModule(img, originX, originY, scale, shape, radius, darkColor, gradientLUT); err != nil {
				return nil, err
			}
		}
	}

	return img, nil
}

func drawModule(
	img *image.RGBA,
	originX int,
	originY int,
	scale int,
	shape string,
	radius float64,
	dark color.RGBA,
	gradientLUT []color.RGBA,
) error {
	switch shape {
	case "square":
		return fillRect(img, originX, originY, scale, scale, dark, gradientLUT)
	case "rounded":
		return fillRoundedRect(img, originX, originY, scale, scale, radius, dark, gradientLUT)
	case "dot":
		return fillCircle(img, originX, originY, scale, dark, gradientLUT)
	default:
		return fmt.Errorf("unknown shape: %s", shape)
	}
}

func fillRect(img *image.RGBA, x, y, w, h int, dark color.RGBA, gradientLUT []color.RGBA) error {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			img.SetRGBA(x+px, y+py, pickColor(dark, gradientLUT, w, px, py))
		}
	}
	return nil
}

func fillRoundedRect(img *image.RGBA, x, y, w, h int, radius float64, dark color.RGBA, gradientLUT []color.RGBA) error {
	r := math.Max(0, math.Min(radius, 0.5)) * float64(w)
	r2 := r * r
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			fx := float64(px) + 0.5
			fy := float64(py) + 0.5
			if insideRoundedRect(fx, fy, float64(w), float64(h), r, r2) {
				img.SetRGBA(x+px, y+py, pickColor(dark, gradientLUT, w, px, py))
			}
		}
	}
	return nil
}

func insideRoundedRect(x, y, w, h, r, r2 float64) bool {
	if r <= 0 {
		return true
	}
	if x >= r && x <= w-r {
		return true
	}
	if y >= r && y <= h-r {
		return true
	}
	if x < r && y < r {
		dx := x - r
		dy := y - r
		return dx*dx+dy*dy <= r2
	}
	if x > w-r && y < r {
		dx := x - (w - r)
		dy := y - r
		return dx*dx+dy*dy <= r2
	}
	if x < r && y > h-r {
		dx := x - r
		dy := y - (h - r)
		return dx*dx+dy*dy <= r2
	}
	if x > w-r && y > h-r {
		dx := x - (w - r)
		dy := y - (h - r)
		return dx*dx+dy*dy <= r2
	}
	return false
}

func fillCircle(img *image.RGBA, x, y, scale int, dark color.RGBA, gradientLUT []color.RGBA) error {
	r := float64(scale) * 0.45
	r2 := r * r
	cx := float64(scale) / 2
	cy := float64(scale) / 2
	for py := 0; py < scale; py++ {
		for px := 0; px < scale; px++ {
			dx := float64(px) + 0.5 - cx
			dy := float64(py) + 0.5 - cy
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(x+px, y+py, pickColor(dark, gradientLUT, scale, px, py))
			}
		}
	}
	return nil
}

func buildGradientLUT(size int, from color.RGBA, to color.RGBA) []color.RGBA {
	if size <= 0 {
		return nil
	}
	lut := make([]color.RGBA, size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			u := (float64(x) + 0.5) / float64(size)
			v := (float64(y) + 0.5) / float64(size)
			t := (u + v) / 2
			lut[y*size+x] = lerpColor(from, to, t)
		}
	}
	return lut
}

func pickColor(base color.RGBA, lut []color.RGBA, size int, px int, py int) color.RGBA {
	if lut == nil {
		return base
	}
	return lut[py*size+px]
}

func lerpColor(from color.RGBA, to color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(math.Round(float64(from.R) + (float64(to.R)-float64(from.R))*t)),
		G: uint8(math.Round(float64(from.G) + (float64(to.G)-float64(from.G))*t)),
		B: uint8(math.Round(float64(from.B) + (float64(to.B)-float64(from.B))*t)),
		A: uint8(math.Round(float64(from.A) + (float64(to.A)-float64(from.A))*t)),
	}
}

func parseColor(value string) (color.RGBA, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return color.RGBA{}, errors.New("color is empty")
	}
	if trimmed == "transparent" {
		return color.RGBA{0, 0, 0, 0}, nil
	}
	if strings.HasPrefix(trimmed, "#") {
		hex := trimmed[1:]
		switch len(hex) {
		case 3:
			r, err := parseHexNibble(hex[0])
			if err != nil {
				return color.RGBA{}, err
			}
			g, err := parseHexNibble(hex[1])
			if err != nil {
				return color.RGBA{}, err
			}
			b, err := parseHexNibble(hex[2])
			if err != nil {
				return color.RGBA{}, err
			}
			return color.RGBA{r * 17, g * 17, b * 17, 255}, nil
		case 6:
			r, err := parseHexByte(hex[0:2])
			if err != nil {
				return color.RGBA{}, err
			}
			g, err := parseHexByte(hex[2:4])
			if err != nil {
				return color.RGBA{}, err
			}
			b, err := parseHexByte(hex[4:6])
			if err != nil {
				return color.RGBA{}, err
			}
			return color.RGBA{r, g, b, 255}, nil
		case 8:
			r, err := parseHexByte(hex[0:2])
			if err != nil {
				return color.RGBA{}, err
			}
			g, err := parseHexByte(hex[2:4])
			if err != nil {
				return color.RGBA{}, err
			}
			b, err := parseHexByte(hex[4:6])
			if err != nil {
				return color.RGBA{}, err
			}
			a, err := parseHexByte(hex[6:8])
			if err != nil {
				return color.RGBA{}, err
			}
			return color.RGBA{r, g, b, a}, nil
		default:
			return color.RGBA{}, fmt.Errorf("unsupported hex length: %d", len(hex))
		}
	}
	if trimmed == "black" {
		return color.RGBA{0, 0, 0, 255}, nil
	}
	if trimmed == "white" {
		return color.RGBA{255, 255, 255, 255}, nil
	}
	return color.RGBA{}, fmt.Errorf("unsupported color format: %s", value)
}

func parseHexByte(value string) (uint8, error) {
	if len(value) != 2 {
		return 0, errors.New("invalid hex byte")
	}
	hi, err := parseHexNibble(value[0])
	if err != nil {
		return 0, err
	}
	lo, err := parseHexNibble(value[1])
	if err != nil {
		return 0, err
	}
	return hi*16 + lo, nil
}

func parseHexNibble(ch byte) (uint8, error) {
	switch {
	case ch >= '0' && ch <= '9':
		return uint8(ch - '0'), nil
	case ch >= 'a' && ch <= 'f':
		return uint8(ch-'a') + 10, nil
	case ch >= 'A' && ch <= 'F':
		return uint8(ch-'A') + 10, nil
	default:
		return 0, fmt.Errorf("invalid hex digit: %c", ch)
	}
}
