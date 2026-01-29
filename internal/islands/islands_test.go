package islands

import (
	"sort"
	"testing"
)

func TestFindIslands(t *testing.T) {
	matrix := [][]bool{
		{true, false, true},
		{true, false, true},
		{false, false, true},
	}
	islands4 := FindIslandsWithConnectivity(matrix, Connectivity4)
	if len(islands4) != 2 {
		t.Fatalf("expected 2 islands, got %d", len(islands4))
	}
	sort.Slice(islands4, func(i, j int) bool {
		if islands4[i].MinY == islands4[j].MinY {
			return islands4[i].MinX < islands4[j].MinX
		}
		return islands4[i].MinY < islands4[j].MinY
	})
	if islands4[0].MinX != 0 || islands4[0].MinY != 0 || islands4[0].MaxX != 0 || islands4[0].MaxY != 1 {
		t.Fatalf("unexpected bounds for island 0: %+v", islands4[0])
	}
	if islands4[1].MinX != 2 || islands4[1].MinY != 0 || islands4[1].MaxX != 2 || islands4[1].MaxY != 2 {
		t.Fatalf("unexpected bounds for island 1: %+v", islands4[1])
	}
}

func TestFindIslandsConnectivity8(t *testing.T) {
	matrix := [][]bool{
		{true, false},
		{false, true},
	}
	islands4 := FindIslandsWithConnectivity(matrix, Connectivity4)
	if len(islands4) != 2 {
		t.Fatalf("expected 2 islands for 4-connected, got %d", len(islands4))
	}
	islands8 := FindIslandsWithConnectivity(matrix, Connectivity8)
	if len(islands8) != 1 {
		t.Fatalf("expected 1 island for 8-connected, got %d", len(islands8))
	}
}

func TestCornerMaskAt(t *testing.T) {
	matrix := [][]bool{{true}}
	mask := CornerMaskAt(matrix, 0, 0, Connectivity4)
	if !mask.Has(CornerTopLeft) || !mask.Has(CornerTopRight) || !mask.Has(CornerBottomRight) || !mask.Has(CornerBottomLeft) {
		t.Fatalf("expected all corners set, got %v", mask)
	}
}
