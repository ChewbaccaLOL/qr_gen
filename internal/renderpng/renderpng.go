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

type GradientSpec struct {
	From     color.RGBA
	To       color.RGBA
	Angle    float64
	FromStop float64
	ToStop   float64
	Scope    string
}

type GradientLUT struct {
	Mode   string
	Width  int
	Height int
	Data   []color.RGBA
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
	backgroundGradient *config.Gradient,
) (*image.RGBA, error) {
	return RenderPNGWithOffsets(
		matrix,
		scale,
		border,
		dark,
		light,
		shape,
		radius,
		gradient,
		backgroundGradient,
		nil,
		0,
		0,
		0,
		"after",
		false,
	)
}

func RenderPNGWithOffsets(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
	backgroundGradient *config.Gradient,
	columnOffsets []ColumnOffset,
	extraPadX int,
	extraPadY int,
	rotateDeg float64,
	rotateMode string,
	rotateTiles bool,
) (*image.RGBA, error) {
	if scale <= 0 {
		return nil, errors.New("scale must be greater than 0")
	}
	size := len(matrix)
	if size == 0 {
		return nil, errors.New("matrix is empty")
	}
	if rotateMode != "after" && rotateMode != "before" {
		return nil, fmt.Errorf("unknown rotate mode: %s", rotateMode)
	}
	dim := (size + border*2) * scale
	contentWidth := dim + extraPadX*2
	contentHeight := dim + extraPadY*2
	width := float64(contentWidth)
	height := float64(contentHeight)
	widthInt := int(math.Ceil(width))
	heightInt := int(math.Ceil(height))
	if widthInt <= 0 || heightInt <= 0 {
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

	img := image.NewRGBA(image.Rect(0, 0, widthInt, heightInt))
	if lightColor != nil && backgroundGradient == nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{C: *lightColor}, image.Point{}, draw.Src)
	}

	islandShape := isIslandShape(shape)
	connectivity := islandConnectivity(shape)
	var gradientSpec *GradientSpec
	var gradientLUT *GradientLUT
	var globalLUT *GradientLUT
	if gradient != nil {
		spec, err := buildGradientSpec(gradient)
		if err != nil {
			return nil, err
		}
		gradientSpec = spec
		if spec.Scope == "global" {
			globalLUT = buildGlobalGradientLUT(widthInt, heightInt, spec)
			if !islandShape {
				gradientLUT = globalLUT
			}
		} else if !islandShape {
			gradientLUT = buildModuleGradientLUT(scale, spec)
		}
	}

	var bgLUT *GradientLUT
	if backgroundGradient != nil {
		spec, err := buildGradientSpec(backgroundGradient)
		if err != nil {
			return nil, err
		}
		spec.Scope = "global"
		bgLUT = buildGlobalGradientLUT(widthInt, heightInt, spec)
		if bgLUT != nil {
			applyGradientBackground(img, bgLUT)
		}
	}

	centerX := float64(contentWidth) / 2
	centerY := float64(contentHeight) / 2
	translateX := 0.0
	translateY := 0.0
	baseOffsetX := float64(extraPadX)
	baseOffsetY := float64(extraPadY)
	angleRad := rotateDeg * math.Pi / 180
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	if islandShape {
		if rotateDeg != 0 && rotateTiles {
			baseImg := image.NewRGBA(image.Rect(0, 0, contentWidth, contentHeight))
			if err := drawIslands(baseImg, matrix, scale, border, radius, darkColor, gradientSpec, globalLUT, connectivity, columnOffsets, baseOffsetX, baseOffsetY, 0, rotateMode, centerX, centerY); err != nil {
				return nil, err
			}
			rotateInto(img, baseImg, translateX, translateY, centerX, centerY, cosA, sinA)
			return img, nil
		}
		if err := drawIslands(img, matrix, scale, border, radius, darkColor, gradientSpec, globalLUT, connectivity, columnOffsets, baseOffsetX, baseOffsetY, rotateDeg, rotateMode, centerX, centerY); err != nil {
			return nil, err
		}
		return img, nil
	}

	if rotateDeg != 0 && rotateTiles {
		baseImg := image.NewRGBA(image.Rect(0, 0, contentWidth, contentHeight))
		for y, row := range matrix {
			for x, cell := range row {
				if !cell {
					continue
				}
				offset := columnOffsetValue(columnOffsets, x)
				px := float64(x+border)*float64(scale) + baseOffsetX
				py := float64(y+border)*float64(scale) + baseOffsetY
				if rotateMode == "before" {
					px += offset.X
					py += offset.Y
				} else if rotateMode == "after" {
					px += offset.X
					py += offset.Y
				}
				if err := drawModule(baseImg, int(math.Round(px)), int(math.Round(py)), scale, shape, radius, darkColor, gradientLUT); err != nil {
					return nil, err
				}
			}
		}
		rotateInto(img, baseImg, translateX, translateY, centerX, centerY, cosA, sinA)
		return img, nil
	}

	for y, row := range matrix {
		for x, cell := range row {
			if !cell {
				continue
			}
			offset := columnOffsetValue(columnOffsets, x)
			baseX := float64(x+border)*float64(scale) + baseOffsetX
			baseY := float64(y+border)*float64(scale) + baseOffsetY
			px := baseX
			py := baseY
			if rotateMode == "before" {
				px += offset.X
				py += offset.Y
			}
			if rotateDeg != 0 {
				dx := px - centerX
				dy := py - centerY
				rx := dx*cosA - dy*sinA
				ry := dx*sinA + dy*cosA
				px = rx + centerX
				py = ry + centerY
			}
			if rotateMode == "after" {
				px += offset.X
				py += offset.Y
			}
			px += translateX
			py += translateY
			if err := drawModule(img, int(math.Round(px)), int(math.Round(py)), scale, shape, radius, darkColor, gradientLUT); err != nil {
				return nil, err
			}
		}
	}

	return img, nil
}

type ColumnOffset struct {
	X float64
	Y float64
}

func columnOffsetValue(offsets []ColumnOffset, index int) ColumnOffset {
	if len(offsets) == 0 || index >= len(offsets) {
		return ColumnOffset{}
	}
	return offsets[index]
}

func rotateInto(dst *image.RGBA, src *image.RGBA, translateX float64, translateY float64, centerX float64, centerY float64, cosA float64, sinA float64) {
	bounds := dst.Bounds()
	srcBounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			fx := float64(x) - translateX
			fy := float64(y) - translateY
			dx := fx - centerX
			dy := fy - centerY
			sx := dx*cosA + dy*sinA + centerX
			sy := -dx*sinA + dy*cosA + centerY
			ix := int(math.Round(sx))
			iy := int(math.Round(sy))
			if ix < srcBounds.Min.X || iy < srcBounds.Min.Y || ix >= srcBounds.Max.X || iy >= srcBounds.Max.Y {
				continue
			}
			pixel := src.RGBAAt(ix, iy)
			if pixel.A == 0 {
				continue
			}
			dst.SetRGBA(x, y, pixel)
		}
	}
}

func drawModule(
	img *image.RGBA,
	originX int,
	originY int,
	scale int,
	shape string,
	radius float64,
	dark color.RGBA,
	gradientLUT *GradientLUT,
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

func fillRect(img *image.RGBA, x, y, w, h int, dark color.RGBA, gradientLUT *GradientLUT) error {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			img.SetRGBA(x+px, y+py, pickColor(dark, gradientLUT, w, px, py, x+px, y+py))
		}
	}
	return nil
}

func fillRoundedRect(img *image.RGBA, x, y, w, h int, radius float64, dark color.RGBA, gradientLUT *GradientLUT) error {
	r := math.Max(0, math.Min(radius, 0.5)) * float64(w)
	r2 := r * r
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			fx := float64(px) + 0.5
			fy := float64(py) + 0.5
			if insideRoundedRect(fx, fy, float64(w), float64(h), r, r2) {
				img.SetRGBA(x+px, y+py, pickColor(dark, gradientLUT, w, px, py, x+px, y+py))
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

func fillCircle(img *image.RGBA, x, y, scale int, dark color.RGBA, gradientLUT *GradientLUT) error {
	r := float64(scale) * 0.45
	r2 := r * r
	cx := float64(scale) / 2
	cy := float64(scale) / 2
	for py := 0; py < scale; py++ {
		for px := 0; px < scale; px++ {
			dx := float64(px) + 0.5 - cx
			dy := float64(py) + 0.5 - cy
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(x+px, y+py, pickColor(dark, gradientLUT, scale, px, py, x+px, y+py))
			}
		}
	}
	return nil
}

func pickColor(base color.RGBA, lut *GradientLUT, size int, px int, py int, absX int, absY int) color.RGBA {
	if lut == nil {
		return base
	}
	if lut.Mode == "global" {
		if absX < 0 || absY < 0 || absX >= lut.Width || absY >= lut.Height {
			return base
		}
		return lut.Data[absY*lut.Width+absX]
	}
	if size <= 0 {
		return base
	}
	return lut.Data[py*size+px]
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

func buildGradientSpec(gradient *config.Gradient) (*GradientSpec, error) {
	if gradient == nil {
		return nil, nil
	}
	fromValue := gradient.From
	if strings.TrimSpace(fromValue) == "" {
		fromValue = "#000000"
	}
	toValue := gradient.To
	if strings.TrimSpace(toValue) == "" {
		toValue = "#ffffff"
	}
	from, err := parseColor(fromValue)
	if err != nil {
		return nil, fmt.Errorf("invalid gradient from color: %w", err)
	}
	to, err := parseColor(toValue)
	if err != nil {
		return nil, fmt.Errorf("invalid gradient to color: %w", err)
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
	scope := strings.ToLower(strings.TrimSpace(gradient.Scope))
	if scope == "" {
		scope = "module"
	}
	if scope != "global" {
		scope = "module"
	}
	return &GradientSpec{
		From:     from,
		To:       to,
		Angle:    angle,
		FromStop: fromStop,
		ToStop:   toStop,
		Scope:    scope,
	}, nil
}

func buildModuleGradientLUT(size int, spec *GradientSpec) *GradientLUT {
	if spec == nil || size <= 0 {
		return nil
	}
	data := buildGradientLUT(float64(size), float64(size), size, size, spec)
	return &GradientLUT{
		Mode:   "module",
		Width:  size,
		Height: size,
		Data:   data,
	}
}

func buildGlobalGradientLUT(width int, height int, spec *GradientSpec) *GradientLUT {
	if spec == nil || width <= 0 || height <= 0 {
		return nil
	}
	data := buildGradientLUT(float64(width), float64(height), width, height, spec)
	return &GradientLUT{
		Mode:   "global",
		Width:  width,
		Height: height,
		Data:   data,
	}
}

func buildGradientLUT(width float64, height float64, widthInt int, heightInt int, spec *GradientSpec) []color.RGBA {
	if spec == nil || widthInt <= 0 || heightInt <= 0 {
		return nil
	}
	axis := gradientAxisForRect(width, height, spec.Angle)
	out := make([]color.RGBA, widthInt*heightInt)
	for y := 0; y < heightInt; y++ {
		for x := 0; x < widthInt; x++ {
			px := float64(x) + 0.5
			py := float64(y) + 0.5
			t := gradientAt(px, py, axis)
			t = applyStops(t, spec.FromStop, spec.ToStop)
			out[y*widthInt+x] = lerpColor(spec.From, spec.To, t)
		}
	}
	return out
}

type gradientAxis struct {
	dx  float64
	dy  float64
	min float64
	max float64
}

func gradientAxisForRect(width float64, height float64, angleDeg float64) gradientAxis {
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
	return gradientAxis{dx: dx, dy: dy, min: minProj, max: maxProj}
}

func gradientAt(x float64, y float64, axis gradientAxis) float64 {
	denom := axis.max - axis.min
	if denom == 0 {
		return 0
	}
	proj := x*axis.dx + y*axis.dy
	return (proj - axis.min) / denom
}

func applyStops(t float64, fromStop float64, toStop float64) float64 {
	denom := toStop - fromStop
	if denom == 0 {
		return 0
	}
	return clampUnit((t - fromStop) / denom)
}

func applyGradientBackground(img *image.RGBA, lut *GradientLUT) {
	if img == nil || lut == nil || lut.Mode != "global" {
		return
	}
	for y := 0; y < lut.Height; y++ {
		for x := 0; x < lut.Width; x++ {
			img.SetRGBA(x, y, lut.Data[y*lut.Width+x])
		}
	}
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
