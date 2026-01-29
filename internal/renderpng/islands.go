package renderpng

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"qr_generator/internal/islands"
)

type islandGradient struct {
	LUT     *GradientLUT
	OffsetX int
	OffsetY int
}

type islandCell struct {
	cell islands.Cell
	x    int
	y    int
}

func drawIslands(
	img *image.RGBA,
	matrix [][]bool,
	scale int,
	border int,
	radius float64,
	dark color.RGBA,
	gradient *GradientSpec,
	globalLUT *GradientLUT,
	connectivity islands.Connectivity,
	columnOffsets []ColumnOffset,
	baseOffsetX float64,
	baseOffsetY float64,
	rotateDeg float64,
	rotateMode string,
	centerX float64,
	centerY float64,
) error {
	islandList := islands.FindIslandsWithConnectivity(matrix, connectivity)
	if len(islandList) == 0 {
		return nil
	}
	corner := math.Max(0, math.Min(radius, 0.5)) * float64(scale)
	angleRad := rotateDeg * math.Pi / 180
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	for _, island := range islandList {
		if len(island.Cells) == 0 {
			continue
		}
		positions := make([]islandCell, 0, len(island.Cells))
		minX := math.MaxInt
		minY := math.MaxInt
		maxX := math.MinInt
		maxY := math.MinInt
		for _, cell := range island.Cells {
			offset := columnOffsetValue(columnOffsets, cell.X)
			px := float64(cell.X+border)*float64(scale) + baseOffsetX
			py := float64(cell.Y+border)*float64(scale) + baseOffsetY
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
			originX := int(math.Round(px))
			originY := int(math.Round(py))
			positions = append(positions, islandCell{cell: cell, x: originX, y: originY})
			if originX < minX {
				minX = originX
			}
			if originY < minY {
				minY = originY
			}
			if originX+scale > maxX {
				maxX = originX + scale
			}
			if originY+scale > maxY {
				maxY = originY + scale
			}
		}

		var grad *islandGradient
		if gradient != nil {
			if gradient.Scope == "global" {
				if globalLUT != nil {
					grad = &islandGradient{LUT: globalLUT, OffsetX: 0, OffsetY: 0}
				}
			} else {
				width := maxX - minX
				height := maxY - minY
				if width > 0 && height > 0 {
					data := buildGradientLUT(float64(width), float64(height), width, height, gradient)
					grad = &islandGradient{
						LUT:     &GradientLUT{Mode: "island", Width: width, Height: height, Data: data},
						OffsetX: minX,
						OffsetY: minY,
					}
				}
			}
		}

		for _, pos := range positions {
			mask := islands.CornerMaskAt(matrix, pos.cell.X, pos.cell.Y, connectivity)
			if err := fillIslandModule(img, pos.x, pos.y, scale, corner, mask, dark, grad); err != nil {
				return err
			}
		}
	}
	return nil
}

func fillIslandModule(
	img *image.RGBA,
	originX int,
	originY int,
	scale int,
	corner float64,
	mask islands.CornerMask,
	dark color.RGBA,
	gradient *islandGradient,
) error {
	if scale <= 0 {
		return fmt.Errorf("invalid module size")
	}
	r := corner
	r2 := r * r
	for py := 0; py < scale; py++ {
		for px := 0; px < scale; px++ {
			fx := float64(px) + 0.5
			fy := float64(py) + 0.5
			if insideSelectiveRoundedRect(fx, fy, float64(scale), float64(scale), r, r2, mask) {
				color := pickIslandColor(dark, gradient, originX+px, originY+py)
				img.SetRGBA(originX+px, originY+py, color)
			}
		}
	}
	return nil
}

func insideSelectiveRoundedRect(x float64, y float64, w float64, h float64, r float64, r2 float64, mask islands.CornerMask) bool {
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
		if mask.Has(islands.CornerTopLeft) {
			dx := x - r
			dy := y - r
			return dx*dx+dy*dy <= r2
		}
		return true
	}
	if x > w-r && y < r {
		if mask.Has(islands.CornerTopRight) {
			dx := x - (w - r)
			dy := y - r
			return dx*dx+dy*dy <= r2
		}
		return true
	}
	if x > w-r && y > h-r {
		if mask.Has(islands.CornerBottomRight) {
			dx := x - (w - r)
			dy := y - (h - r)
			return dx*dx+dy*dy <= r2
		}
		return true
	}
	if x < r && y > h-r {
		if mask.Has(islands.CornerBottomLeft) {
			dx := x - r
			dy := y - (h - r)
			return dx*dx+dy*dy <= r2
		}
		return true
	}
	return true
}

func pickIslandColor(base color.RGBA, gradient *islandGradient, absX int, absY int) color.RGBA {
	if gradient == nil || gradient.LUT == nil {
		return base
	}
	lut := gradient.LUT
	switch lut.Mode {
	case "global":
		if absX < 0 || absY < 0 || absX >= lut.Width || absY >= lut.Height {
			return base
		}
		return lut.Data[absY*lut.Width+absX]
	case "island":
		relX := absX - gradient.OffsetX
		relY := absY - gradient.OffsetY
		if relX < 0 || relY < 0 || relX >= lut.Width || relY >= lut.Height {
			return base
		}
		return lut.Data[relY*lut.Width+relX]
	default:
		return base
	}
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
