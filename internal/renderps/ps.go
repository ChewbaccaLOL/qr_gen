package renderps

import (
	"bytes"
	"fmt"
	"math"
	"os"

	"qr_generator/internal/config"
)

type PSDocument struct {
	buf    bytes.Buffer
	width  float64
	height float64
}

func RenderPS(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
) (*PSDocument, error) {
	size := len(matrix)
	if size == 0 {
		return nil, fmt.Errorf("matrix is empty")
	}
	dim := float64((size + border*2) * scale)
	doc := newPS(dim, dim)

	if light != nil {
		color, err := parseColor(*light)
		if err != nil {
			return nil, err
		}
		writeRectFill(&doc.buf, 0, 0, dim, dim, color)
	}

	darkColor, err := parseColor(dark)
	if err != nil {
		return nil, err
	}

	var grad *gradientColors
	if gradient != nil {
		grad, err = parseGradient(gradient)
		if err != nil {
			return nil, err
		}
	}

	corner := clampRadius(radius, float64(scale))

	for y, row := range matrix {
		for x, cell := range row {
			if !cell {
				continue
			}
			px := float64(x+border) * float64(scale)
			py := float64(y+border) * float64(scale)
			drawModulePS(&doc.buf, px, py, float64(scale), shape, corner, darkColor, grad)
		}
	}

	doc.finish()
	return doc, nil
}

func RenderCatalogPS(
	matrix [][]bool,
	scale int,
	border int,
	variants []config.Variant,
	columns int,
	background string,
	labelSize int,
) (*PSDocument, error) {
	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants available for catalog")
	}
	size := len(matrix)
	if size == 0 {
		return nil, fmt.Errorf("matrix is empty")
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
	width := float64(columns*tileDim + (columns+1)*padding)
	height := float64(rows*tileTotalHeight + padding)

	bg, err := parseColor(background)
	if err != nil {
		return nil, err
	}

	doc := newPS(width, height)
	writeRectFill(&doc.buf, 0, 0, width, height, bg)

	for index, variant := range variants {
		col := index % columns
		row := index / columns
		originX := float64(padding + col*(tileDim+padding))
		originY := float64(padding + row*tileTotalHeight)
		tileBg := bg
		if variant.Light != nil {
			parsed, err := parseColor(*variant.Light)
			if err != nil {
				return nil, err
			}
			tileBg = parsed
		}
		writeRectFill(&doc.buf, originX, originY, float64(tileDim), float64(tileDim), tileBg)

		darkColor, err := parseColor(variant.Dark)
		if err != nil {
			return nil, err
		}
		var grad *gradientColors
		if variant.Gradient != nil {
			grad, err = parseGradient(variant.Gradient)
			if err != nil {
				return nil, err
			}
		}
		corner := clampRadius(variant.Radius, float64(scale))

		for y, rowVals := range matrix {
			for x, cell := range rowVals {
				if !cell {
					continue
				}
				px := originX + float64(x+border)*float64(scale)
				py := originY + float64(y+border)*float64(scale)
				drawModulePS(&doc.buf, px, py, float64(scale), variant.Shape, corner, darkColor, grad)
			}
		}

		labelX := originX + float64(tileDim)/2
		labelY := originY + float64(tileDim) + float64(labelHeight)*0.75
		writeTextCentered(&doc.buf, labelX, labelY, float64(labelSize), variant.Name)
	}

	doc.finish()
	return doc, nil
}

func (d *PSDocument) Write(path string) error {
	return os.WriteFile(path, d.buf.Bytes(), 0o644)
}

func newPS(width float64, height float64) *PSDocument {
	doc := &PSDocument{width: width, height: height}
	fmt.Fprintf(&doc.buf, "%%!PS-Adobe-3.0\n")
	fmt.Fprintf(&doc.buf, "%%%%BoundingBox: 0 0 %d %d\n", int(math.Ceil(width)), int(math.Ceil(height)))
	fmt.Fprintf(&doc.buf, "%%%%Pages: 1\n")
	fmt.Fprintf(&doc.buf, "%%%%EndComments\n")
	fmt.Fprintf(&doc.buf, "0 %.4f translate\n1 -1 scale\n", height)
	return doc
}

func (d *PSDocument) finish() {
	fmt.Fprintf(&d.buf, "showpage\n%%%%EOF\n")
}

type rgb struct {
	R int
	G int
	B int
}

type gradientColors struct {
	From rgb
	To   rgb
}

func parseGradient(gradient *config.Gradient) (*gradientColors, error) {
	from, err := parseColor(gradient.From)
	if err != nil {
		return nil, err
	}
	to, err := parseColor(gradient.To)
	if err != nil {
		return nil, err
	}
	return &gradientColors{From: from, To: to}, nil
}

func parseColor(value string) (rgb, error) {
	color, err := parseHexColor(value)
	if err != nil {
		return rgb{}, err
	}
	return rgb{R: color.R, G: color.G, B: color.B}, nil
}

func drawModulePS(buf *bytes.Buffer, x, y, size float64, shape string, corner float64, color rgb, gradient *gradientColors) {
	if gradient != nil {
		writeGradientFill(buf, x, y, size, shape, corner, gradient)
		return
	}
	switch shape {
	case "square":
		writeRectFill(buf, x, y, size, size, color)
	case "rounded":
		writeRoundedRectFill(buf, x, y, size, size, corner, color)
	case "dot":
		writeCircleFill(buf, x+size/2, y+size/2, size*0.45, color)
	default:
		writeRectFill(buf, x, y, size, size, color)
	}
}

func writeRectFill(buf *bytes.Buffer, x, y, w, h float64, color rgb) {
	fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
	fmt.Fprintf(buf, "%.4f %.4f %.4f %.4f rectfill\n", x, y, w, h)
}

func writeRoundedRectFill(buf *bytes.Buffer, x, y, w, h, r float64, color rgb) {
	if r <= 0 {
		writeRectFill(buf, x, y, w, h, color)
		return
	}
	fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
	writeRoundedRectPath(buf, x, y, w, h, r)
	fmt.Fprintf(buf, "fill\n")
}

func writeRoundedRectPath(buf *bytes.Buffer, x, y, w, h, r float64) {
	x0 := x
	y0 := y
	x1 := x + w
	y1 := y + h
	fmt.Fprintf(buf, "newpath\n")
	fmt.Fprintf(buf, "%.4f %.4f moveto\n", x0+r, y0)
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x1-r, y0)
	fmt.Fprintf(buf, "%.4f %.4f %.4f 270 360 arc\n", x1-r, y0+r, r)
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x1, y1-r)
	fmt.Fprintf(buf, "%.4f %.4f %.4f 0 90 arc\n", x1-r, y1-r, r)
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x0+r, y1)
	fmt.Fprintf(buf, "%.4f %.4f %.4f 90 180 arc\n", x0+r, y1-r, r)
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x0, y0+r)
	fmt.Fprintf(buf, "%.4f %.4f %.4f 180 270 arc\n", x0+r, y0+r, r)
	fmt.Fprintf(buf, "closepath\n")
}

func writeCircleFill(buf *bytes.Buffer, cx, cy, r float64, color rgb) {
	fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
	fmt.Fprintf(buf, "newpath\n")
	fmt.Fprintf(buf, "%.4f %.4f %.4f 0 360 arc\n", cx, cy, r)
	fmt.Fprintf(buf, "closepath fill\n")
}

func writeGradientFill(buf *bytes.Buffer, x, y, size float64, shape string, corner float64, gradient *gradientColors) {
	fmt.Fprintf(buf, "gsave\n")
	switch shape {
	case "square":
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f %.4f rectclip\n", x, y, size, size)
	case "rounded":
		writeRoundedRectPath(buf, x, y, size, size, corner)
		fmt.Fprintf(buf, "clip\nnewpath\n")
	case "dot":
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f 0 360 arc\nclip\nnewpath\n", x+size/2, y+size/2, size*0.45)
	default:
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f %.4f rectclip\n", x, y, size, size)
	}

	steps := 10
	stepSize := size / float64(steps)
	for iy := 0; iy < steps; iy++ {
		for ix := 0; ix < steps; ix++ {
			px := x + float64(ix)*stepSize
			py := y + float64(iy)*stepSize
			cx := (float64(ix) + 0.5) * stepSize
			cy := (float64(iy) + 0.5) * stepSize
			t := (cx + cy) / (2 * size)
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			color := mixColor(gradient.From, gradient.To, t)
			fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
			fmt.Fprintf(buf, "%.4f %.4f %.4f %.4f rectfill\n", px, py, stepSize, stepSize)
		}
	}

	fmt.Fprintf(buf, "grestore\n")
}

func mixColor(a rgb, b rgb, t float64) rgb {
	return rgb{
		R: int(math.Round(float64(a.R) + (float64(b.R)-float64(a.R))*t)),
		G: int(math.Round(float64(a.G) + (float64(b.G)-float64(a.G))*t)),
		B: int(math.Round(float64(a.B) + (float64(b.B)-float64(a.B))*t)),
	}
}

func clampRadius(radius float64, scale float64) float64 {
	if radius < 0 {
		radius = 0
	}
	if radius > 0.5 {
		radius = 0.5
	}
	return radius * scale
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func writeTextCentered(buf *bytes.Buffer, x, y, size float64, text string) {
	fmt.Fprintf(buf, "/Helvetica findfont %.4f scalefont setfont\n", size)
	fmt.Fprintf(buf, "gsave\n")
	fmt.Fprintf(buf, "%.4f %.4f moveto\n", x, y)
	fmt.Fprintf(buf, "(%s) dup stringwidth pop 2 div neg 0 rmoveto show\n", escapeText(text))
	fmt.Fprintf(buf, "grestore\n")
}

func escapeText(text string) string {
	escaped := bytes.NewBuffer(nil)
	for i := 0; i < len(text); i++ {
		ch := text[i]
		switch ch {
		case '(', ')', '\\':
			escaped.WriteByte('\\')
		}
		escaped.WriteByte(ch)
	}
	return escaped.String()
}
