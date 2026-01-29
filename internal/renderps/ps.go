package renderps

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"

	"qr_generator/internal/config"
	"qr_generator/internal/islands"
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
	backgroundGradient *config.Gradient,
) (*PSDocument, error) {
	size := len(matrix)
	if size == 0 {
		return nil, fmt.Errorf("matrix is empty")
	}
	dim := float64((size + border*2) * scale)
	doc := newPS(dim, dim)

	if backgroundGradient != nil {
		spec, err := parseGradientSpec(backgroundGradient)
		if err != nil {
			return nil, err
		}
		area := rect{X: 0, Y: 0, W: dim, H: dim}
		writeGradientFillRect(&doc.buf, area, spec, area)
	} else if light != nil {
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

	var grad *gradientSpec
	if gradient != nil {
		grad, err = parseGradientSpec(gradient)
		if err != nil {
			return nil, err
		}
	}

	corner := clampRadius(radius, float64(scale))

	if isIslandShape(shape) {
		if err := drawIslandsPS(&doc.buf, matrix, scale, border, radius, darkColor, grad, islandConnectivity(shape), rect{X: 0, Y: 0, W: dim, H: dim}, 0, 0); err != nil {
			return nil, err
		}
		doc.finish()
		return doc, nil
	}

	for y, row := range matrix {
		for x, cell := range row {
			if !cell {
				continue
			}
			px := float64(x+border) * float64(scale)
			py := float64(y+border) * float64(scale)
			drawModulePS(&doc.buf, px, py, float64(scale), shape, corner, darkColor, grad, rect{X: 0, Y: 0, W: dim, H: dim})
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
		tileRect := rect{X: originX, Y: originY, W: float64(tileDim), H: float64(tileDim)}
		if variant.BackgroundGradient != nil {
			bgSpec, err := parseGradientSpec(variant.BackgroundGradient)
			if err != nil {
				return nil, err
			}
			writeGradientFillRect(&doc.buf, tileRect, bgSpec, tileRect)
		} else {
			tileBg := bg
			if variant.Light != nil {
				parsed, err := parseColor(*variant.Light)
				if err != nil {
					return nil, err
				}
				tileBg = parsed
			}
			writeRectFill(&doc.buf, originX, originY, float64(tileDim), float64(tileDim), tileBg)
		}

		darkColor, err := parseColor(variant.Dark)
		if err != nil {
			return nil, err
		}
		var grad *gradientSpec
		if variant.Gradient != nil {
			grad, err = parseGradientSpec(variant.Gradient)
			if err != nil {
				return nil, err
			}
		}
		corner := clampRadius(variant.Radius, float64(scale))

		if isIslandShape(variant.Shape) {
			if err := drawIslandsPS(&doc.buf, matrix, scale, border, variant.Radius, darkColor, grad, islandConnectivity(variant.Shape), tileRect, originX, originY); err != nil {
				return nil, err
			}
		} else {
			for y, rowVals := range matrix {
				for x, cell := range rowVals {
					if !cell {
						continue
					}
					px := originX + float64(x+border)*float64(scale)
					py := originY + float64(y+border)*float64(scale)
					drawModulePS(&doc.buf, px, py, float64(scale), variant.Shape, corner, darkColor, grad, tileRect)
				}
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

func drawIslandsPS(buf *bytes.Buffer, matrix [][]bool, scale int, border int, radius float64, color rgb, gradient *gradientSpec, connectivity islands.Connectivity, global rect, offsetX float64, offsetY float64) error {
	islandList := islands.FindIslandsWithConnectivity(matrix, connectivity)
	if len(islandList) == 0 {
		return nil
	}
	corner := clampRadius(radius, float64(scale))
	for _, island := range islandList {
		if len(island.Cells) == 0 {
			continue
		}
		islandRect := rect{
			X: offsetX + float64(island.MinX+border)*float64(scale),
			Y: offsetY + float64(island.MinY+border)*float64(scale),
			W: float64(island.MaxX-island.MinX+1) * float64(scale),
			H: float64(island.MaxY-island.MinY+1) * float64(scale),
		}
		baseRect := global
		if gradient != nil && gradient.Scope != "global" {
			baseRect = islandRect
		}
		for _, cell := range island.Cells {
			px := offsetX + float64(cell.X+border)*float64(scale)
			py := offsetY + float64(cell.Y+border)*float64(scale)
			mask := islands.CornerMaskAt(matrix, cell.X, cell.Y, connectivity)
			drawIslandModulePS(buf, px, py, float64(scale), corner, mask, color, gradient, baseRect)
		}
	}
	return nil
}

func drawIslandModulePS(buf *bytes.Buffer, x float64, y float64, size float64, corner float64, mask islands.CornerMask, color rgb, gradient *gradientSpec, base rect) {
	rTL := 0.0
	if mask.Has(islands.CornerTopLeft) {
		rTL = corner
	}
	rTR := 0.0
	if mask.Has(islands.CornerTopRight) {
		rTR = corner
	}
	rBR := 0.0
	if mask.Has(islands.CornerBottomRight) {
		rBR = corner
	}
	rBL := 0.0
	if mask.Has(islands.CornerBottomLeft) {
		rBL = corner
	}

	if gradient != nil {
		writeGradientFillExt(buf, rect{X: x, Y: y, W: size, H: size}, rTL, rTR, rBR, rBL, gradient, base)
		return
	}
	if rTL > 0 || rTR > 0 || rBR > 0 || rBL > 0 {
		writeRoundedRectFillExt(buf, x, y, size, size, rTL, rTR, rBR, rBL, color)
		return
	}
	writeRectFill(buf, x, y, size, size, color)
}

func isIslandShape(shape string) bool {
	switch strings.ToLower(strings.TrimSpace(shape)) {
	case "island", "island-4", "island-8":
		return true
	default:
		return false
	}
}

func islandConnectivity(shape string) islands.Connectivity {
	switch strings.ToLower(strings.TrimSpace(shape)) {
	case "island-8":
		return islands.Connectivity8
	default:
		return islands.Connectivity4
	}
}

func (d *PSDocument) finish() {
	fmt.Fprintf(&d.buf, "showpage\n%%%%EOF\n")
}

type rgb struct {
	R int
	G int
	B int
}

type gradientSpec struct {
	From     rgb
	To       rgb
	Angle    float64
	FromStop float64
	ToStop   float64
	Scope    string
}

type rect struct {
	X float64
	Y float64
	W float64
	H float64
}

func parseGradientSpec(gradient *config.Gradient) (*gradientSpec, error) {
	from, err := parseColor(gradient.From)
	if err != nil {
		return nil, err
	}
	to, err := parseColor(gradient.To)
	if err != nil {
		return nil, err
	}
	angle := 45.0
	if gradient.Angle != nil {
		angle = *gradient.Angle
	}
	fromStop := 0.0
	if gradient.FromStop != nil {
		fromStop = *gradient.FromStop
	}
	toStop := 1.0
	if gradient.ToStop != nil {
		toStop = *gradient.ToStop
	}
	fromStop = clampUnit(fromStop)
	toStop = clampUnit(toStop)
	if toStop < fromStop {
		fromStop, toStop = toStop, fromStop
	}
	scope := strings.TrimSpace(gradient.Scope)
	if scope == "" {
		scope = "module"
	}
	if scope != "global" {
		scope = "module"
	}
	return &gradientSpec{From: from, To: to, Angle: angle, FromStop: fromStop, ToStop: toStop, Scope: scope}, nil
}

func parseColor(value string) (rgb, error) {
	color, err := parseHexColor(value)
	if err != nil {
		return rgb{}, err
	}
	return rgb{R: color.R, G: color.G, B: color.B}, nil
}

func drawModulePS(buf *bytes.Buffer, x, y, size float64, shape string, corner float64, color rgb, gradient *gradientSpec, global rect) {
	if gradient != nil {
		area := rect{X: x, Y: y, W: size, H: size}
		base := area
		if gradient.Scope == "global" {
			base = global
		}
		writeGradientFill(buf, area, shape, corner, gradient, base)
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

func writeRoundedRectFillExt(buf *bytes.Buffer, x, y, w, h, rTL, rTR, rBR, rBL float64, color rgb) {
	if rTL <= 0 && rTR <= 0 && rBR <= 0 && rBL <= 0 {
		writeRectFill(buf, x, y, w, h, color)
		return
	}
	fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
	writeRoundedRectPathExt(buf, x, y, w, h, rTL, rTR, rBR, rBL)
	fmt.Fprintf(buf, "fill\n")
}

func writeRoundedRectPathExt(buf *bytes.Buffer, x, y, w, h, rTL, rTR, rBR, rBL float64) {
	x0 := x
	y0 := y
	x1 := x + w
	y1 := y + h
	fmt.Fprintf(buf, "newpath\n")
	fmt.Fprintf(buf, "%.4f %.4f moveto\n", x0+rTL, y0)
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x1-rTR, y0)
	if rTR > 0 {
		fmt.Fprintf(buf, "%.4f %.4f %.4f 270 360 arc\n", x1-rTR, y0+rTR, rTR)
	} else {
		fmt.Fprintf(buf, "%.4f %.4f lineto\n", x1, y0)
	}
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x1, y1-rBR)
	if rBR > 0 {
		fmt.Fprintf(buf, "%.4f %.4f %.4f 0 90 arc\n", x1-rBR, y1-rBR, rBR)
	} else {
		fmt.Fprintf(buf, "%.4f %.4f lineto\n", x1, y1)
	}
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x0+rBL, y1)
	if rBL > 0 {
		fmt.Fprintf(buf, "%.4f %.4f %.4f 90 180 arc\n", x0+rBL, y1-rBL, rBL)
	} else {
		fmt.Fprintf(buf, "%.4f %.4f lineto\n", x0, y1)
	}
	fmt.Fprintf(buf, "%.4f %.4f lineto\n", x0, y0+rTL)
	if rTL > 0 {
		fmt.Fprintf(buf, "%.4f %.4f %.4f 180 270 arc\n", x0+rTL, y0+rTL, rTL)
	} else {
		fmt.Fprintf(buf, "%.4f %.4f lineto\n", x0, y0)
	}
	fmt.Fprintf(buf, "closepath\n")
}

func writeCircleFill(buf *bytes.Buffer, cx, cy, r float64, color rgb) {
	fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
	fmt.Fprintf(buf, "newpath\n")
	fmt.Fprintf(buf, "%.4f %.4f %.4f 0 360 arc\n", cx, cy, r)
	fmt.Fprintf(buf, "closepath fill\n")
}

func writeGradientFill(buf *bytes.Buffer, area rect, shape string, corner float64, gradient *gradientSpec, base rect) {
	fmt.Fprintf(buf, "gsave\n")
	switch shape {
	case "square":
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f %.4f rectclip\n", area.X, area.Y, area.W, area.H)
	case "rounded":
		writeRoundedRectPath(buf, area.X, area.Y, area.W, area.H, corner)
		fmt.Fprintf(buf, "clip\nnewpath\n")
	case "dot":
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f 0 360 arc\nclip\nnewpath\n", area.X+area.W/2, area.Y+area.H/2, area.W*0.45)
	default:
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f %.4f rectclip\n", area.X, area.Y, area.W, area.H)
	}

	writeGradientFillRect(buf, area, gradient, base)

	fmt.Fprintf(buf, "grestore\n")
}

func writeGradientFillExt(buf *bytes.Buffer, area rect, rTL, rTR, rBR, rBL float64, gradient *gradientSpec, base rect) {
	fmt.Fprintf(buf, "gsave\n")
	if rTL > 0 || rTR > 0 || rBR > 0 || rBL > 0 {
		writeRoundedRectPathExt(buf, area.X, area.Y, area.W, area.H, rTL, rTR, rBR, rBL)
		fmt.Fprintf(buf, "clip\nnewpath\n")
	} else {
		fmt.Fprintf(buf, "newpath\n%.4f %.4f %.4f %.4f rectclip\n", area.X, area.Y, area.W, area.H)
	}
	writeGradientFillRect(buf, area, gradient, base)
	fmt.Fprintf(buf, "grestore\n")
}

func writeGradientFillRect(buf *bytes.Buffer, area rect, gradient *gradientSpec, base rect) {
	stepsX, stepsY := gradientSteps(area.W, area.H)
	stepX := area.W / float64(stepsX)
	stepY := area.H / float64(stepsY)
	for iy := 0; iy < stepsY; iy++ {
		for ix := 0; ix < stepsX; ix++ {
			px := area.X + float64(ix)*stepX
			py := area.Y + float64(iy)*stepY
			cx := px + stepX*0.5
			cy := py + stepY*0.5
			t := gradientT(cx, cy, base, gradient.Angle, gradient.FromStop, gradient.ToStop)
			color := mixColor(gradient.From, gradient.To, t)
			fmt.Fprintf(buf, "%.4f %.4f %.4f setrgbcolor\n", float64(color.R)/255.0, float64(color.G)/255.0, float64(color.B)/255.0)
			fmt.Fprintf(buf, "%.4f %.4f %.4f %.4f rectfill\n", px, py, stepX, stepY)
		}
	}
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

func minInt(a, b int) int {
	if a < b {
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

func gradientSteps(width float64, height float64) (int, int) {
	stepsX := int(math.Ceil(width / 8))
	stepsY := int(math.Ceil(height / 8))
	stepsX = maxInt(6, minInt(60, stepsX))
	stepsY = maxInt(6, minInt(60, stepsY))
	return stepsX, stepsY
}

func gradientT(x float64, y float64, base rect, angle float64, fromStop float64, toStop float64) float64 {
	x1, y1, x2, y2 := gradientLine(angle, base.W, base.H)
	x1 += base.X
	y1 += base.Y
	x2 += base.X
	y2 += base.Y
	dx := x2 - x1
	dy := y2 - y1
	denom := dx*dx + dy*dy
	if denom == 0 {
		return 0
	}
	t := ((x-x1)*dx + (y-y1)*dy) / denom
	t = clampUnit(t)
	if toStop == fromStop {
		if t < fromStop {
			return 0
		}
		return 1
	}
	return clampUnit((t - fromStop) / (toStop - fromStop))
}

func gradientLine(angleDeg float64, width float64, height float64) (float64, float64, float64, float64) {
	rad := angleDeg * math.Pi / 180
	dx := math.Cos(rad)
	dy := math.Sin(rad)
	corners := [][2]float64{
		{0, 0},
		{width, 0},
		{0, height},
		{width, height},
	}
	minProj := math.Inf(1)
	maxProj := math.Inf(-1)
	for _, corner := range corners {
		proj := corner[0]*dx + corner[1]*dy
		if proj < minProj {
			minProj = proj
		}
		if proj > maxProj {
			maxProj = proj
		}
	}
	return dx * minProj, dy * minProj, dx * maxProj, dy * maxProj
}

func clampUnit(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
