package renderpdf

import (
	"fmt"
	"math"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"qr_generator/internal/config"
)

type PDFDocument struct {
	pdf    *gofpdf.Fpdf
	width  float64
	height float64
}

func RenderPDF(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
	backgroundGradient *config.Gradient,
) (*PDFDocument, error) {
	size := len(matrix)
	if size == 0 {
		return nil, fmt.Errorf("matrix is empty")
	}
	dim := float64((size + border*2) * scale)
	doc := newPDF(dim, dim)
	if backgroundGradient != nil {
		spec, err := parseGradientSpec(backgroundGradient)
		if err != nil {
			return nil, err
		}
		x1, y1, x2, y2 := gradientCoords(spec.Angle, dim, dim, spec.FromStop, spec.ToStop)
		doc.pdf.LinearGradient(0, 0, dim, dim, spec.From.R, spec.From.G, spec.From.B, spec.To.R, spec.To.G, spec.To.B, x1, y1, x2, y2)
	} else if light != nil {
		color, err := parseColor(*light)
		if err != nil {
			return nil, err
		}
		doc.pdf.SetFillColor(color.R, color.G, color.B)
		doc.pdf.Rect(0, 0, dim, dim, "F")
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

	for y, row := range matrix {
		for x, cell := range row {
			if !cell {
				continue
			}
			px := float64(x+border) * float64(scale)
			py := float64(y+border) * float64(scale)
			drawModulePDF(doc.pdf, px, py, float64(scale), shape, corner, darkColor, grad, rect{X: 0, Y: 0, W: dim, H: dim})
		}
	}

	return doc, nil
}

func RenderCatalogPDF(
	matrix [][]bool,
	scale int,
	border int,
	variants []config.Variant,
	columns int,
	background string,
	labelSize int,
) (*PDFDocument, error) {
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

	doc := newPDF(width, height)
	doc.pdf.SetFillColor(bg.R, bg.G, bg.B)
	doc.pdf.Rect(0, 0, width, height, "F")

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
			x1, y1, x2, y2 := gradientCoords(bgSpec.Angle, tileRect.W, tileRect.H, bgSpec.FromStop, bgSpec.ToStop)
			doc.pdf.ClipRect(tileRect.X, tileRect.Y, tileRect.W, tileRect.H, false)
			doc.pdf.LinearGradient(tileRect.X, tileRect.Y, tileRect.W, tileRect.H, bgSpec.From.R, bgSpec.From.G, bgSpec.From.B, bgSpec.To.R, bgSpec.To.G, bgSpec.To.B, x1, y1, x2, y2)
			doc.pdf.ClipEnd()
		} else {
			tileBg := bg
			if variant.Light != nil {
				parsed, err := parseColor(*variant.Light)
				if err != nil {
					return nil, err
				}
				tileBg = parsed
			}
			doc.pdf.SetFillColor(tileBg.R, tileBg.G, tileBg.B)
			doc.pdf.Rect(originX, originY, float64(tileDim), float64(tileDim), "F")
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

		for y, rowVals := range matrix {
			for x, cell := range rowVals {
				if !cell {
					continue
				}
				px := originX + float64(x+border)*float64(scale)
				py := originY + float64(y+border)*float64(scale)
				drawModulePDF(doc.pdf, px, py, float64(scale), variant.Shape, corner, darkColor, grad, tileRect)
			}
		}

		labelX := originX + float64(tileDim)/2
		labelY := originY + float64(tileDim) + float64(labelHeight)*0.75
		doc.pdf.SetFont("Helvetica", "", float64(labelSize))
		labelWidth := doc.pdf.GetStringWidth(variant.Name)
		doc.pdf.Text(labelX-labelWidth/2, labelY, variant.Name)
	}

	return doc, nil
}

func (d *PDFDocument) Write(path string) error {
	return d.pdf.OutputFileAndClose(path)
}

func newPDF(width float64, height float64) *PDFDocument {
	init := gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: width, Ht: height},
	}
	pdf := gofpdf.NewCustom(&init)
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	return &PDFDocument{pdf: pdf, width: width, height: height}
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

func drawModulePDF(pdf *gofpdf.Fpdf, x, y, size float64, shape string, corner float64, color rgb, gradient *gradientSpec, global rect) {
	if gradient != nil {
		if gradient.Scope == "global" {
			x1, y1, x2, y2 := gradientCoords(gradient.Angle, global.W, global.H, gradient.FromStop, gradient.ToStop)
			switch shape {
			case "square":
				pdf.ClipRect(x, y, size, size, false)
				pdf.LinearGradient(global.X, global.Y, global.W, global.H, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, x1, y1, x2, y2)
				pdf.ClipEnd()
			case "rounded":
				pdf.ClipRoundedRect(x, y, size, size, corner, false)
				pdf.LinearGradient(global.X, global.Y, global.W, global.H, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, x1, y1, x2, y2)
				pdf.ClipEnd()
			case "dot":
				pdf.ClipCircle(x+size/2, y+size/2, size*0.45, false)
				pdf.LinearGradient(global.X, global.Y, global.W, global.H, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, x1, y1, x2, y2)
				pdf.ClipEnd()
			default:
				pdf.SetFillColor(color.R, color.G, color.B)
				pdf.Rect(x, y, size, size, "F")
			}
			return
		}
		x1, y1, x2, y2 := gradientCoords(gradient.Angle, size, size, gradient.FromStop, gradient.ToStop)
		switch shape {
		case "square":
			pdf.LinearGradient(x, y, size, size, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, x1, y1, x2, y2)
		case "rounded":
			pdf.ClipRoundedRect(x, y, size, size, corner, false)
			pdf.LinearGradient(x, y, size, size, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, x1, y1, x2, y2)
			pdf.ClipEnd()
		case "dot":
			pdf.ClipCircle(x+size/2, y+size/2, size*0.45, false)
			pdf.LinearGradient(x, y, size, size, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, x1, y1, x2, y2)
			pdf.ClipEnd()
		default:
			pdf.SetFillColor(color.R, color.G, color.B)
			pdf.Rect(x, y, size, size, "F")
		}
		return
	}

	pdf.SetFillColor(color.R, color.G, color.B)
	switch shape {
	case "square":
		pdf.Rect(x, y, size, size, "F")
	case "rounded":
		pdf.RoundedRect(x, y, size, size, corner, "F", "1234")
	case "dot":
		pdf.Circle(x+size/2, y+size/2, size*0.45, "F")
	default:
		pdf.Rect(x, y, size, size, "F")
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

func gradientCoords(angleDeg float64, width float64, height float64, fromStop float64, toStop float64) (float64, float64, float64, float64) {
	x1, y1, x2, y2 := gradientLine(angleDeg, width, height)
	fromStop = clampUnit(fromStop)
	toStop = clampUnit(toStop)
	if toStop < fromStop {
		fromStop, toStop = toStop, fromStop
	}
	dx := x2 - x1
	dy := y2 - y1
	sx := x1 + dx*fromStop
	sy := y1 + dy*fromStop
	ex := x1 + dx*toStop
	ey := y1 + dy*toStop
	if width != 0 {
		sx /= width
		ex /= width
	}
	if height != 0 {
		sy /= height
		ey /= height
	}
	return sx, sy, ex, ey
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
