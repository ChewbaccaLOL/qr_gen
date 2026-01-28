package renderpng

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"math"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"qr_generator/internal/config"
)

func RenderCatalogPNG(
	matrix [][]bool,
	scale int,
	border int,
	variants []config.Variant,
	columns int,
	background string,
	labelSize int,
) (*image.RGBA, error) {
	if len(variants) == 0 {
		return nil, errors.New("no variants available for catalog")
	}
	if scale <= 0 {
		return nil, errors.New("scale must be greater than 0")
	}
	size := len(matrix)
	if size == 0 {
		return nil, errors.New("matrix is empty")
	}

	tileDim := (size + border*2) * scale
	padding := maxInt(8, int(float64(scale)*1.2))
	if labelSize <= 0 {
		labelSize = maxInt(10, int(float64(scale)*1.4))
	}
	labelHeight := int(float64(labelSize) * 1.6)
	tileTotalHeight := tileDim + labelHeight + padding
	if columns < 1 {
		columns = 1
	}
	rows := int(math.Ceil(float64(len(variants)) / float64(columns)))
	width := columns*tileDim + (columns+1)*padding
	height := rows*tileTotalHeight + padding

	backgroundColor, err := parseColor(background)
	if err != nil {
		return nil, fmt.Errorf("invalid catalog background: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	stddraw.Draw(img, img.Bounds(), &image.Uniform{C: backgroundColor}, image.Point{}, stddraw.Src)

	labelColor, err := parseColor("#1a1a1a")
	if err != nil {
		return nil, fmt.Errorf("invalid label color: %w", err)
	}

	for index, variant := range variants {
		col := index % columns
		row := index / columns
		originX := padding + col*(tileDim+padding)
		originY := padding + row*tileTotalHeight

		tileBg := backgroundColor
		if variant.Light != nil {
			parsed, err := parseColor(*variant.Light)
			if err != nil {
				return nil, fmt.Errorf("invalid variant background: %w", err)
			}
			tileBg = parsed
		}
		fillRect(img, originX, originY, tileDim, tileDim, tileBg, nil)

		if err := drawMatrixAt(img, matrix, scale, border, variant, originX, originY); err != nil {
			return nil, err
		}

		labelX := originX + tileDim/2
		labelY := originY + tileDim + int(float64(labelHeight)*0.75)
		drawLabel(img, variant.Name, labelX, labelY, labelSize, labelColor)
	}

	return img, nil
}

func drawMatrixAt(img *image.RGBA, matrix [][]bool, scale int, border int, variant config.Variant, offsetX int, offsetY int) error {
	darkColor, err := parseColor(variant.Dark)
	if err != nil {
		return fmt.Errorf("invalid dark color: %w", err)
	}

	var gradientLUT []color.RGBA
	if variant.Gradient != nil {
		from, err := parseColor(variant.Gradient.From)
		if err != nil {
			return fmt.Errorf("invalid gradient from color: %w", err)
		}
		to, err := parseColor(variant.Gradient.To)
		if err != nil {
			return fmt.Errorf("invalid gradient to color: %w", err)
		}
		gradientLUT = buildGradientLUT(scale, from, to)
	}

	for y, row := range matrix {
		for x, cell := range row {
			if !cell {
				continue
			}
			originX := offsetX + (x+border)*scale
			originY := offsetY + (y+border)*scale
			if err := drawModule(img, originX, originY, scale, variant.Shape, variant.Radius, darkColor, gradientLUT); err != nil {
				return err
			}
		}
	}

	return nil
}

func drawLabel(img *image.RGBA, text string, xCenter int, yBaseline int, size int, color color.RGBA) {
	if text == "" {
		return
	}
	if size <= 0 {
		return
	}
	face, err := fontFace(size)
	if err != nil {
		return
	}
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color),
		Face: face,
	}
	width := drawer.MeasureString(text).Ceil()
	if width <= 0 {
		return
	}
	drawer.Dot = fixed.P(xCenter-width/2, yBaseline)
	drawer.DrawString(text)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	fontOnce sync.Once
	fontErr  error
	fontData *opentype.Font
	fontMu   sync.Mutex
	fonts    = map[int]font.Face{}
)

func fontFace(size int) (font.Face, error) {
	fontOnce.Do(func() {
		fontData, fontErr = opentype.Parse(goregular.TTF)
	})
	if fontErr != nil {
		return nil, fontErr
	}
	if size <= 0 {
		return nil, errors.New("font size must be greater than 0")
	}
	fontMu.Lock()
	defer fontMu.Unlock()
	if face, ok := fonts[size]; ok {
		return face, nil
	}
	face, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	fonts[size] = face
	return face, nil
}
