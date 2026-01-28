package animation

import "testing"

func TestComputeWaveOffsetsSize(t *testing.T) {
	offsets := computeWaveOffsets(5, 2.0, 10.0, 0.0)
	if len(offsets) != 5 {
		t.Fatalf("expected 5 offsets, got %d", len(offsets))
	}
}

func TestWaveRampMultiplier(t *testing.T) {
	value := waveRampMultiplier(0, 10, 4)
	if value != 0 {
		t.Fatalf("expected ramp=0 at start, got %v", value)
	}
	value = waveRampMultiplier(5, 10, 4)
	if value != 1 {
		t.Fatalf("expected ramp=1 mid-wave, got %v", value)
	}
}

func TestQuantizeOffset(t *testing.T) {
	value := quantizeOffset(2.3, 0.5)
	if value != 2.5 {
		t.Fatalf("unexpected quantized value: %v", value)
	}
}

func TestBuildFloatOffsetsCycles(t *testing.T) {
	matrix := make([][]bool, 3)
	for i := range matrix {
		matrix[i] = make([]bool, 3)
	}
	total, _, _, _ := buildFloatOffsets(matrix, 1, 0.4, 8, 90, 4, 2, 3, "still", 0, 0, "after")
	expected := 2 + (4 * 3) + 2
	if total != expected {
		t.Fatalf("expected %d frames, got %d", expected, total)
	}
}
