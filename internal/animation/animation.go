package animation

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"math"

	"qr_generator/internal/config"
	"qr_generator/internal/render"
	"qr_generator/internal/renderpng"
)

func BuildWaveGIF(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
	backgroundGradient *config.Gradient,
	waveAmp float64,
	wavePeriod float64,
	frames int,
	hold int,
	mode string,
	fps int,
	rotateTiles bool,
) (*gif.GIF, error) {
	if frames <= 0 {
		return nil, errors.New("frames must be greater than 0")
	}
	if fps <= 0 {
		return nil, errors.New("fps must be greater than 0")
	}
	if wavePeriod <= 0 {
		return nil, errors.New("wave period must be greater than 0")
	}
	if mode != "still" && mode != "loop" {
		return nil, fmt.Errorf("unknown wave mode: %s", mode)
	}

	totalFrames, offsets := buildWaveOffsets(matrix, scale, waveAmp, wavePeriod, frames, hold, mode)
	extraPadY := 0
	amplitudePx := waveAmp * float64(scale)
	if amplitudePx != 0 {
		extraPadY = int(math.Ceil(math.Abs(amplitudePx))) + 1
	}

	framesOut := make([]*image.Paletted, 0, totalFrames)
	delays := make([]int, 0, totalFrames)
	delay := gifDelay(fps)

	for _, columnOffsets := range offsets {
		img, err := renderpng.RenderPNGWithOffsets(
			matrix,
			scale,
			border,
			dark,
			light,
			shape,
			radius,
			gradient,
			backgroundGradient,
			toPngOffsets(columnOffsets),
			0,
			extraPadY,
			0,
			"after",
			rotateTiles,
		)
		if err != nil {
			return nil, err
		}
		framesOut = append(framesOut, toPaletted(img, light == nil))
		delays = append(delays, delay)
	}

	gifOut := &gif.GIF{Image: framesOut, Delay: delays, Disposal: make([]byte, len(framesOut))}
	for i := range gifOut.Disposal {
		gifOut.Disposal[i] = gif.DisposalBackground
	}
	return gifOut, nil
}

func BuildFloatGIF(
	matrix [][]bool,
	scale int,
	border int,
	dark string,
	light *string,
	shape string,
	radius float64,
	gradient *config.Gradient,
	backgroundGradient *config.Gradient,
	floatAmp float64,
	floatPeriod float64,
	floatAngleDeg float64,
	frames int,
	hold int,
	cycles int,
	mode string,
	snap float64,
	rotateDeg float64,
	rotateMode string,
	fps int,
	rotateTiles bool,
) (*gif.GIF, error) {
	if frames <= 0 {
		return nil, errors.New("frames must be greater than 0")
	}
	if fps <= 0 {
		return nil, errors.New("fps must be greater than 0")
	}
	if floatPeriod <= 0 {
		return nil, errors.New("float period must be greater than 0")
	}
	if mode != "still" && mode != "loop" {
		return nil, fmt.Errorf("unknown float mode: %s", mode)
	}
	if rotateMode != "after" && rotateMode != "before" {
		return nil, fmt.Errorf("unknown rotate mode: %s", rotateMode)
	}

	totalFrames, offsets, extraPadX, extraPadY := buildFloatOffsets(
		matrix,
		scale,
		floatAmp,
		floatPeriod,
		floatAngleDeg,
		frames,
		hold,
		cycles,
		mode,
		snap,
		rotateDeg,
		rotateMode,
	)

	framesOut := make([]*image.Paletted, 0, totalFrames)
	delays := make([]int, 0, totalFrames)
	delay := gifDelay(fps)

	for _, columnOffsets := range offsets {
		img, err := renderpng.RenderPNGWithOffsets(
			matrix,
			scale,
			border,
			dark,
			light,
			shape,
			radius,
			gradient,
			backgroundGradient,
			toPngOffsets(columnOffsets),
			extraPadX,
			extraPadY,
			rotateDeg,
			rotateMode,
			rotateTiles,
		)
		if err != nil {
			return nil, err
		}
		framesOut = append(framesOut, toPaletted(img, light == nil))
		delays = append(delays, delay)
	}

	gifOut := &gif.GIF{Image: framesOut, Delay: delays, Disposal: make([]byte, len(framesOut))}
	for i := range gifOut.Disposal {
		gifOut.Disposal[i] = gif.DisposalBackground
	}
	return gifOut, nil
}

func buildWaveOffsets(matrix [][]bool, scale int, waveAmp float64, wavePeriod float64, frames int, hold int, mode string) (int, [][]render.ColumnOffset) {
	size := len(matrix)
	amplitudePx := waveAmp * float64(scale)
	looped := mode == "loop"
	phaseDenominator := frames
	if looped {
		phaseDenominator = maxInt(frames-1, 1)
	}
	phaseStep := (2 * math.Pi) / float64(phaseDenominator)
	stillOffsets := make([]float64, size)
	totalFrames := frames
	if !looped {
		totalFrames = hold + frames + hold
	}

	rampFrames := 0
	if !looped {
		rampFrames = maxInt(2, int(math.Round(float64(frames)*0.2)))
	}

	framesOffsets := make([][]render.ColumnOffset, 0, totalFrames)
	for frameIndex := 0; frameIndex < totalFrames; frameIndex++ {
		var offsets []float64
		if looped {
			phase := phaseStep * float64(frameIndex)
			offsets = computeWaveOffsets(size, amplitudePx, wavePeriod, phase)
		} else {
			if frameIndex < hold || frameIndex >= hold+frames {
				offsets = stillOffsets
			} else {
				waveIndex := frameIndex - hold
				phase := phaseStep * float64(waveIndex)
				ramp := waveRampMultiplier(waveIndex, frames, rampFrames)
				offsets = computeWaveOffsets(size, amplitudePx*ramp, wavePeriod, phase)
			}
		}
		framesOffsets = append(framesOffsets, render.ColumnOffsetsFromY(offsets))
	}
	return totalFrames, framesOffsets
}

func buildFloatOffsets(
	matrix [][]bool,
	scale int,
	floatAmp float64,
	floatPeriod float64,
	floatAngleDeg float64,
	frames int,
	hold int,
	cycles int,
	mode string,
	snap float64,
	rotateDeg float64,
	rotateMode string,
) (int, [][]render.ColumnOffset, int, int) {
	size := len(matrix)
	amplitudePx := floatAmp * float64(scale)
	looped := mode == "loop"
	if cycles <= 0 {
		cycles = 1
	}
	if looped {
		cycles = 1
	}
	phaseDenominator := frames
	if looped {
		phaseDenominator = maxInt(frames-1, 1)
	}
	phaseStep := (2 * math.Pi) / float64(phaseDenominator)
	stillOffsets := make([]float64, size)
	snapPx := snap * float64(scale)
	clothAmp := amplitudePx * 0.4

	angleRad := floatAngleDeg * math.Pi / 180
	if rotateMode == "after" && rotateDeg != 0 {
		angleRad = (floatAngleDeg - rotateDeg) * math.Pi / 180
	}
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)
	maxOffset := math.Abs(amplitudePx) + math.Abs(clothAmp)
	maxOffsetX := math.Abs(maxOffset * cosA)
	maxOffsetY := math.Abs(maxOffset * sinA)
	extraPadX := 0
	extraPadY := 0
	if maxOffsetX != 0 {
		extraPadX = int(math.Ceil(maxOffsetX)) + 1
	}
	if maxOffsetY != 0 {
		extraPadY = int(math.Ceil(maxOffsetY)) + 1
	}

	activeFrames := frames
	if !looped {
		activeFrames = frames * cycles
	}
	totalFrames := frames
	if !looped {
		totalFrames = hold + activeFrames + hold
	}
	rampFrames := 0
	if !looped {
		rampFrames = maxInt(2, int(math.Round(float64(frames)*0.2)))
	}

	framesOffsets := make([][]render.ColumnOffset, 0, totalFrames)
	for frameIndex := 0; frameIndex < totalFrames; frameIndex++ {
		var offsets []float64
		if looped {
			phase := phaseStep * float64(frameIndex)
			baseOffset := amplitudePx * math.Sin(phase)
			phaseMod := phase * 0.7
			offsets = make([]float64, size)
			for x := 0; x < size; x++ {
				offsets[x] = baseOffset + clothAmp*math.Sin((2*math.Pi*(float64(x)/floatPeriod))+phaseMod)
			}
		} else {
			if frameIndex < hold || frameIndex >= hold+activeFrames {
				offsets = stillOffsets
			} else {
				floatIndex := frameIndex - hold
				phase := phaseStep * float64(floatIndex)
				ramp := waveRampMultiplier(floatIndex, activeFrames, rampFrames)
				baseOffset := amplitudePx * math.Sin(phase) * ramp
				phaseMod := phase * 0.7
				offsets = make([]float64, size)
				for x := 0; x < size; x++ {
					offsets[x] = baseOffset + clothAmp*ramp*math.Sin((2*math.Pi*(float64(x)/floatPeriod))+phaseMod)
				}
			}
		}
		columnOffsets := make([]render.ColumnOffset, size)
		for i, offset := range offsets {
			if snapPx > 0 {
				offset = quantizeOffset(offset, snapPx)
			}
			columnOffsets[i] = render.ColumnOffset{X: offset * cosA, Y: offset * sinA}
		}
		framesOffsets = append(framesOffsets, columnOffsets)
	}

	return totalFrames, framesOffsets, extraPadX, extraPadY
}

func computeWaveOffsets(size int, amplitudePx float64, period float64, phase float64) []float64 {
	offsets := make([]float64, size)
	if amplitudePx == 0 {
		return offsets
	}
	for x := 0; x < size; x++ {
		offsets[x] = amplitudePx * math.Sin((2*math.Pi*(float64(x)/period))+phase)
	}
	return offsets
}

func smoothstep(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t * t * (3 - 2*t)
}

func waveRampMultiplier(frameIndex int, frames int, rampFrames int) float64 {
	if rampFrames <= 1 || frames <= 1 {
		return 1
	}
	if rampFrames > frames/2 {
		rampFrames = frames / 2
	}
	if rampFrames <= 1 {
		return 1
	}
	if frameIndex < rampFrames {
		t := float64(frameIndex) / float64(rampFrames-1)
		return smoothstep(t)
	}
	if frameIndex >= frames-rampFrames {
		t := float64(frames-1-frameIndex) / float64(rampFrames-1)
		return smoothstep(t)
	}
	return 1
}

func quantizeOffset(offset float64, snapPx float64) float64 {
	if snapPx <= 0 {
		return offset
	}
	return math.Round(offset/snapPx) * snapPx
}

func toPaletted(img image.Image, transparent bool) *image.Paletted {
	needsTransparency := transparent || hasTransparency(img)
	if !needsTransparency {
		paletted := image.NewPaletted(img.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(paletted, paletted.Rect, img, image.Point{})
		return paletted
	}

	pal := make(color.Palette, 0, 256)
	pal = append(pal, color.Transparent)
	if len(palette.Plan9) > 255 {
		pal = append(pal, palette.Plan9[:255]...)
	} else {
		pal = append(pal, palette.Plan9...)
	}
	paletted := image.NewPaletted(img.Bounds(), pal)
	draw.FloydSteinberg.Draw(paletted, paletted.Rect, img, image.Point{})

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha == 0 {
				paletted.SetColorIndex(x, y, 0)
			}
		}
	}

	return paletted
}

func hasTransparency(img image.Image) bool {
	if img == nil {
		return false
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha == 0 {
				return true
			}
		}
	}
	return false
}

func gifDelay(fps int) int {
	if fps <= 0 {
		return 1
	}
	return int(math.Round(100.0 / float64(fps)))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func toPngOffsets(offsets []render.ColumnOffset) []renderpng.ColumnOffset {
	if offsets == nil {
		return nil
	}
	out := make([]renderpng.ColumnOffset, len(offsets))
	for i, offset := range offsets {
		out[i] = renderpng.ColumnOffset{X: offset.X, Y: offset.Y}
	}
	return out
}
