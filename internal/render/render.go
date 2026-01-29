package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"qr_generator/internal/config"
)

type ColumnOffset struct {
	X float64
	Y float64
}

func ColumnOffsetsFromY(offsets []float64) []ColumnOffset {
	if offsets == nil {
		return nil
	}
	result := make([]ColumnOffset, len(offsets))
	for i, value := range offsets {
		result[i] = ColumnOffset{X: 0, Y: value}
	}
	return result
}

func RenderSVG(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
	backgroundGradient *config.Gradient,
	cutout bool,
	columnOffsets []ColumnOffset,
	extraPadY int,
	extraPadX int,
	rotateDeg float64,
	rotateMode string,
) (string, error) {
	size := len(matrix)
	totalModules := size + border*2
	dimension := totalModules * scale
	contentWidth := dimension + extraPadX*2
	contentHeight := dimension + extraPadY*2
	width := float64(contentWidth)
	height := float64(contentHeight)
	widthIsInt := rotateDeg == 0
	heightIsInt := rotateDeg == 0

	shapeRendering := "geometricPrecision"
	if shape == "square" {
		shapeRendering = "crispEdges"
	}

	if rotateDeg != 0 {
		angle := rotateDeg * math.Pi / 180
		cosA := math.Abs(math.Cos(angle))
		sinA := math.Abs(math.Sin(angle))
		width = float64(contentWidth)*cosA + float64(contentHeight)*sinA
		height = float64(contentWidth)*sinA + float64(contentHeight)*cosA
		widthIsInt = false
		heightIsInt = false
	}

	widthAttr := formatDim(width, widthIsInt)
	heightAttr := formatDim(height, heightIsInt)

	if cutout && light == nil && backgroundGradient == nil {
		return "", fmt.Errorf("cutout mode requires a background fill")
	}

	parts := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		fmt.Sprintf(
			"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%s\" height=\"%s\" viewBox=\"0 0 %s %s\" shape-rendering=\"%s\">",
			widthAttr,
			heightAttr,
			widthAttr,
			heightAttr,
			shapeRendering,
		),
	}

	defs := []string{}
	gradientID := ""
	fill := ""
	if !cutout {
		if gradient != nil {
			gradientID = gradient.ID
			if gradientID == "" {
				gradientID = "fg"
			}
		}
		fill = moduleFill(dark, gradientID)
		if def := svgGradientDef(gradient, gradientID, width, height, ""); def != "" {
			defs = append(defs, def)
		}
	}
	if backgroundGradient != nil {
		bgID := backgroundGradient.ID
		if bgID == "" {
			bgID = "bg"
		}
		if def := svgGradientDef(backgroundGradient, bgID, width, height, "global"); def != "" {
			defs = append(defs, def)
		}
	}
	if cutout {
		maskID := "qr-cutout"
		maskParts := []string{
			fmt.Sprintf(
				"<mask id=\"%s\" maskUnits=\"userSpaceOnUse\" maskContentUnits=\"userSpaceOnUse\" x=\"0\" y=\"0\" width=\"%s\" height=\"%s\">",
				maskID,
				widthAttr,
				heightAttr,
			),
			"<rect width=\"100%\" height=\"100%\" fill=\"white\"/>",
		}
		moduleParts, err := renderModuleGroup(
			matrix,
			scale,
			border,
			"#000000",
			shape,
			radius,
			columnOffsets,
			extraPadX,
			extraPadY,
			rotateDeg,
			rotateMode,
			(width-float64(contentWidth))/2,
			(height-float64(contentHeight))/2,
			contentWidth,
			contentHeight,
		)
		if err != nil {
			return "", err
		}
		maskParts = append(maskParts, moduleParts...)
		maskParts = append(maskParts, "</mask>")
		defs = append(defs, strings.Join(maskParts, ""))
	}
	if len(defs) > 0 {
		parts = append(parts, fmt.Sprintf("<defs>%s</defs>", strings.Join(defs, "")))
	}

	backgroundFill := ""
	if backgroundGradient != nil {
		bgID := backgroundGradient.ID
		if bgID == "" {
			bgID = "bg"
		}
		backgroundFill = fmt.Sprintf("url(#%s)", bgID)
	} else if light != nil {
		backgroundFill = *light
	}
	if backgroundFill != "" {
		if cutout {
			parts = append(parts, fmt.Sprintf("<rect width=\"100%%\" height=\"100%%\" fill=\"%s\" mask=\"url(#qr-cutout)\"/>", backgroundFill))
		} else {
			parts = append(parts, fmt.Sprintf("<rect width=\"100%%\" height=\"100%%\" fill=\"%s\"/>", backgroundFill))
		}
	}

	if !cutout {
		translateX := (width - float64(contentWidth)) / 2
		translateY := (height - float64(contentHeight)) / 2
		moduleParts, err := renderModuleGroup(
			matrix,
			scale,
			border,
			fill,
			shape,
			radius,
			columnOffsets,
			extraPadX,
			extraPadY,
			rotateDeg,
			rotateMode,
			translateX,
			translateY,
			contentWidth,
			contentHeight,
		)
		if err != nil {
			return "", err
		}
		parts = append(parts, moduleParts...)
	}
	parts = append(parts, "</svg>")
	return strings.Join(parts, "\n"), nil
}

func renderModuleGroup(
	matrix [][]bool,
	scale int,
	border int,
	fill string,
	shape string,
	radius float64,
	columnOffsets []ColumnOffset,
	extraPadX int,
	extraPadY int,
	rotateDeg float64,
	rotateMode string,
	translateX float64,
	translateY float64,
	contentWidth int,
	contentHeight int,
) ([]string, error) {
	parts := []string{fmt.Sprintf("<g transform=\"translate(%s %s)\">", formatFloat(translateX), formatFloat(translateY))}
	if rotateMode != "after" && rotateMode != "before" {
		return nil, fmt.Errorf("unknown rotate mode: %s", rotateMode)
	}
	centerX := float64(contentWidth) / 2
	centerY := float64(contentHeight) / 2
	baseOffsetX := float64(extraPadX)
	baseOffsetY := float64(extraPadY)
	size := len(matrix)

	if rotateMode == "before" && len(columnOffsets) > 0 {
		for columnIndex := 0; columnIndex < size; columnIndex++ {
			offset := columnOffsetValue(columnOffsets, columnIndex)
			if offset.X != 0 || offset.Y != 0 {
				parts = append(parts, fmt.Sprintf("<g transform=\"translate(%s %s)\">", formatFloat(offset.X), formatFloat(offset.Y)))
			}
			if rotateDeg != 0 {
				parts = append(parts, fmt.Sprintf("<g transform=\"rotate(%s %s %s)\">", formatFloat(rotateDeg), formatFloat(centerX), formatFloat(centerY)))
			}
			columnParts, err := renderModulesForColumn(matrix, scale, border, fill, shape, radius, columnIndex, baseOffsetX, baseOffsetY)
			if err != nil {
				return nil, err
			}
			parts = append(parts, columnParts...)
			if rotateDeg != 0 {
				parts = append(parts, "</g>")
			}
			if offset.X != 0 || offset.Y != 0 {
				parts = append(parts, "</g>")
			}
		}
	} else {
		if rotateDeg != 0 {
			parts = append(parts, fmt.Sprintf("<g transform=\"rotate(%s %s %s)\">", formatFloat(rotateDeg), formatFloat(centerX), formatFloat(centerY)))
		}
		moduleParts, err := renderModules(matrix, scale, border, fill, shape, radius, baseOffsetX, baseOffsetY, columnOffsets)
		if err != nil {
			return nil, err
		}
		parts = append(parts, moduleParts...)
		if rotateDeg != 0 {
			parts = append(parts, "</g>")
		}
	}

	parts = append(parts, "</g>")
	return parts, nil
}

func RenderCatalogSVG(
	matrix [][]bool,
	scale int,
	border int,
	variants []config.Variant,
	columns int,
	background string,
	labelSize int,
) (string, error) {
	if len(variants) == 0 {
		return "", fmt.Errorf("no variants available for catalog")
	}

	size := len(matrix)
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

	parts := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		fmt.Sprintf(
			"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\" shape-rendering=\"geometricPrecision\">",
			width,
			height,
			width,
			height,
		),
	}

	gradientDefs := []string{}
	for _, variant := range variants {
		if variant.Gradient != nil {
			gradientID := fmt.Sprintf("fg-%s", variant.Name)
			gradientDefs = append(gradientDefs, svgGradientDef(variant.Gradient, gradientID, 1, 1, "module"))
		}
	}
	if len(gradientDefs) > 0 {
		parts = append(parts, fmt.Sprintf("<defs>%s</defs>", strings.Join(gradientDefs, "")))
	}

	parts = append(parts, fmt.Sprintf("<rect width=\"100%%\" height=\"100%%\" fill=\"%s\"/>", background))

	for index, variant := range variants {
		col := index % columns
		row := index / columns
		originX := padding + col*(tileDim+padding)
		originY := padding + row*tileTotalHeight
		tileBg := background
		if variant.Light != nil {
			tileBg = *variant.Light
		}
		parts = append(parts, fmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>", originX, originY, tileDim, tileDim, tileBg))

		gradientID := ""
		if variant.Gradient != nil {
			gradientID = fmt.Sprintf("fg-%s", variant.Name)
		}
		fill := moduleFill(variant.Dark, gradientID)
		moduleParts, err := renderModules(matrix, scale, border, fill, variant.Shape, variant.Radius, float64(originX), float64(originY), nil)
		if err != nil {
			return "", err
		}
		parts = append(parts, moduleParts...)

		labelX := float64(originX) + float64(tileDim)/2
		labelY := float64(originY) + float64(tileDim) + float64(labelHeight)*0.75
		parts = append(parts, fmt.Sprintf(
			"<text x=\"%s\" y=\"%s\" font-size=\"%d\" font-family=\"Helvetica, Arial, sans-serif\" fill=\"#1a1a1a\" text-anchor=\"middle\">%s</text>",
			formatFloat(labelX),
			formatFloat(labelY),
			labelSize,
			escapeXML(variant.Name),
		))
	}

	parts = append(parts, "</svg>")
	return strings.Join(parts, "\n"), nil
}

func svgGradientDef(gradient *config.Gradient, gradientID string, width float64, height float64, scopeOverride string) string {
	if gradient == nil || gradientID == "" {
		return ""
	}
	colorFrom := gradient.From
	if colorFrom == "" {
		colorFrom = "#000000"
	}
	colorTo := gradient.To
	if colorTo == "" {
		colorTo = "#ffffff"
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
	scope := strings.TrimSpace(gradient.Scope)
	if scopeOverride != "" {
		scope = scopeOverride
	}
	if scope == "" {
		scope = "module"
	}
	if scope != "global" {
		scope = "module"
	}
	fromStop = clampUnit(fromStop)
	toStop = clampUnit(toStop)
	if toStop < fromStop {
		fromStop, toStop = toStop, fromStop
	}
	gradWidth := width
	gradHeight := height
	if scope != "global" {
		gradWidth = 1
		gradHeight = 1
	}
	x1, y1, x2, y2 := gradientLine(angle, gradWidth, gradHeight)
	gradientUnits := ""
	if scope == "global" {
		gradientUnits = " gradientUnits=\"userSpaceOnUse\""
	}
	return fmt.Sprintf(
		"<linearGradient id=\"%s\"%s x1=\"%s\" y1=\"%s\" x2=\"%s\" y2=\"%s\"><stop offset=\"%s\" stop-color=\"%s\"/><stop offset=\"%s\" stop-color=\"%s\"/></linearGradient>",
		gradientID,
		gradientUnits,
		formatFloat(x1),
		formatFloat(y1),
		formatFloat(x2),
		formatFloat(y2),
		formatPercent(fromStop),
		colorFrom,
		formatPercent(toStop),
		colorTo,
	)
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

func formatPercent(value float64) string {
	return fmt.Sprintf("%s%%", formatFloat(value*100))
}

func moduleFill(dark, gradientID string) string {
	if gradientID != "" {
		return fmt.Sprintf("url(#%s)", gradientID)
	}
	return dark
}

func columnOffsetValue(offsets []ColumnOffset, index int) ColumnOffset {
	if len(offsets) == 0 || index >= len(offsets) {
		return ColumnOffset{}
	}
	return offsets[index]
}

func moduleElement(shape string, px, py float64, scale int, radius float64, fill string) (string, error) {
	switch shape {
	case "square":
		return fmt.Sprintf(
			"<rect x=\"%s\" y=\"%s\" width=\"%d\" height=\"%d\" fill=\"%s\"/>",
			formatFloat(px),
			formatFloat(py),
			scale,
			scale,
			fill,
		), nil
	case "rounded":
		corner := math.Max(0, math.Min(radius, 0.5)) * float64(scale)
		cornerStr := formatFloat(corner)
		return fmt.Sprintf(
			"<rect x=\"%s\" y=\"%s\" width=\"%d\" height=\"%d\" rx=\"%s\" ry=\"%s\" fill=\"%s\"/>",
			formatFloat(px),
			formatFloat(py),
			scale,
			scale,
			cornerStr,
			cornerStr,
			fill,
		), nil
	case "dot":
		r := float64(scale) * 0.45
		offsetCenter := float64(scale) / 2
		cx := px + offsetCenter
		cy := py + offsetCenter
		return fmt.Sprintf(
			"<circle cx=\"%s\" cy=\"%s\" r=\"%s\" fill=\"%s\"/>",
			formatFloat(cx),
			formatFloat(cy),
			formatFloat(r),
			fill,
		), nil
	default:
		return "", fmt.Errorf("unknown shape: %s", shape)
	}
}

func renderModules(
	matrix [][]bool,
	scale int,
	border int,
	fill string,
	shape string,
	radius float64,
	offsetX float64,
	offsetY float64,
	columnOffsets []ColumnOffset,
) ([]string, error) {
	parts := []string{}
	for y, row := range matrix {
		for x, cell := range row {
			if cell {
				offset := columnOffsetValue(columnOffsets, x)
				px := float64(x+border)*float64(scale) + offsetX + offset.X
				py := float64(y+border)*float64(scale) + offsetY + offset.Y
				element, err := moduleElement(shape, px, py, scale, radius, fill)
				if err != nil {
					return nil, err
				}
				parts = append(parts, element)
			}
		}
	}
	return parts, nil
}

func renderModulesForColumn(
	matrix [][]bool,
	scale int,
	border int,
	fill string,
	shape string,
	radius float64,
	columnIndex int,
	offsetX float64,
	offsetY float64,
) ([]string, error) {
	parts := []string{}
	for y, row := range matrix {
		if row[columnIndex] {
			px := float64(columnIndex+border)*float64(scale) + offsetX
			py := float64(y+border)*float64(scale) + offsetY
			element, err := moduleElement(shape, px, py, scale, radius, fill)
			if err != nil {
				return nil, err
			}
			parts = append(parts, element)
		}
	}
	return parts, nil
}

func escapeXML(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
	)
	return replacer.Replace(text)
}

func formatFloat(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Sprintf("%v", value)
	}
	if value == math.Trunc(value) {
		return fmt.Sprintf("%.1f", value)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func formatDim(value float64, isInt bool) string {
	if isInt {
		return fmt.Sprintf("%d", int(value))
	}
	return formatFloat(value)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
