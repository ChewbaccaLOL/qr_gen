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
