package renderpdf

import (
	"fmt"
	"math"

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
) (*PDFDocument, error) {
	size := len(matrix)
	if size == 0 {
		return nil, fmt.Errorf("matrix is empty")
	}
	dim := float64((size + border*2) * scale)
	doc := newPDF(dim, dim)
	if light != nil {
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
			drawModulePDF(doc.pdf, px, py, float64(scale), shape, corner, darkColor, grad)
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
				drawModulePDF(doc.pdf, px, py, float64(scale), variant.Shape, corner, darkColor, grad)
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

func drawModulePDF(pdf *gofpdf.Fpdf, x, y, size float64, shape string, corner float64, color rgb, gradient *gradientColors) {
	if gradient != nil {
		switch shape {
		case "square":
			pdf.LinearGradient(x, y, size, size, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, 0, 0, 1, 1)
		case "rounded":
			pdf.ClipRoundedRect(x, y, size, size, corner, false)
			pdf.LinearGradient(x, y, size, size, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, 0, 0, 1, 1)
			pdf.ClipEnd()
		case "dot":
			pdf.ClipCircle(x+size/2, y+size/2, size*0.45, false)
			pdf.LinearGradient(x, y, size, size, gradient.From.R, gradient.From.G, gradient.From.B, gradient.To.R, gradient.To.G, gradient.To.B, 0, 0, 1, 1)
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
