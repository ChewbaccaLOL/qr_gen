package gui

import (
	"testing"

	"qr_generator/internal/config"
)

func testAnimationConfig() *config.Config {
	light := "#ffffff"
	return &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#000000",
				Light: &light,
			},
		},
		AnimationVariants: []string{"wave", "wave-loop", "float"},
		Defaults: config.Defaults{
			GifFPS:      12,
			GifFrames:   6,
			GifHold:     2,
			WaveAmp:     0.45,
			WavePeriod:  10,
			FloatAngle:  90,
			FloatTilt:   10,
			FloatHold:   3,
			FloatCycles: 2,
			ReadableGif: config.GifDefaults{
				FPS:        10,
				Frames:     4,
				Hold:       4,
				WaveAmp:    0.3,
				WavePeriod: 12,
			},
		},
	}
}

func TestBuildGIFWaveFrames(t *testing.T) {
	cfg := testAnimationConfig()
	frames := 4
	hold := 2
	fps := 8
	waveAmp := 0.4
	wavePeriod := 8.0
	req := AnimationRequest{
		RenderRequest: RenderRequest{
			Data:       "hello",
			Variant:    "classic",
			ErrorLevel: "m",
			Scale:      4,
			Border:     2,
		},
		AnimationVariant: "wave",
		GifFrames:        &frames,
		GifHold:          &hold,
		GifFps:           &fps,
		WaveAmp:          &waveAmp,
		WavePeriod:       &wavePeriod,
	}
	gifOut, err := BuildGIF(cfg, req)
	if err != nil {
		t.Fatalf("build gif: %v", err)
	}
	if len(gifOut.Image) != frames+(hold*2) {
		t.Fatalf("expected %d frames, got %d", frames+(hold*2), len(gifOut.Image))
	}
}

func TestBuildGIFWaveLoopFrames(t *testing.T) {
	cfg := testAnimationConfig()
	frames := 5
	hold := 3
	fps := 8
	waveAmp := 0.4
	wavePeriod := 8.0
	req := AnimationRequest{
		RenderRequest: RenderRequest{
			Data:       "hello",
			Variant:    "classic",
			ErrorLevel: "m",
			Scale:      4,
			Border:     2,
		},
		AnimationVariant: "wave-loop",
		GifFrames:        &frames,
		GifHold:          &hold,
		GifFps:           &fps,
		WaveAmp:          &waveAmp,
		WavePeriod:       &wavePeriod,
	}
	gifOut, err := BuildGIF(cfg, req)
	if err != nil {
		t.Fatalf("build gif: %v", err)
	}
	if len(gifOut.Image) != frames {
		t.Fatalf("expected %d frames, got %d", frames, len(gifOut.Image))
	}
}

func TestBuildGIFUnknownVariant(t *testing.T) {
	cfg := testAnimationConfig()
	req := AnimationRequest{
		RenderRequest: RenderRequest{
			Data:       "hello",
			Variant:    "classic",
			ErrorLevel: "m",
			Scale:      4,
			Border:     2,
		},
		AnimationVariant: "unknown",
	}
	if _, err := BuildGIF(cfg, req); err == nil {
		t.Fatalf("expected error for unknown animation variant")
	}
}

func TestBuildAnimationConfigReadable(t *testing.T) {
	cfg := testAnimationConfig()
	configOut, err := BuildAnimationConfig(cfg)
	if err != nil {
		t.Fatalf("build animation config: %v", err)
	}
	if configOut.Defaults.Readable.GifFrames != cfg.Defaults.ReadableGif.Frames {
		t.Fatalf("expected readable frames %d, got %d", cfg.Defaults.ReadableGif.Frames, configOut.Defaults.Readable.GifFrames)
	}
}
