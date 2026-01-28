package cli

import (
	"strings"
	"testing"

	"qr_generator/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Variants: map[string]config.Variant{
			"classic": {
				Name:  "classic",
				Shape: "square",
				Dark:  "#000000",
			},
		},
		AnimationVariants: []string{"wave", "float"},
		Defaults: config.Defaults{
			GifFPS:          12,
			GifFrames:       40,
			GifHold:         24,
			WaveAmp:         0.45,
			WavePeriod:      10,
			FloatJaggedSnap: 0.25,
			FloatTilt:       18,
			FloatAngle:      90,
			ReadableGif: config.GifDefaults{
				FPS:        12,
				Frames:     32,
				Hold:       32,
				WaveAmp:    0.28,
				WavePeriod: 14,
			},
		},
	}
}

func TestParseListVariantsNoData(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	args, err := ParseArgs([]string{"--list-variants"}, cfg, env, strings.NewReader(""), true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !args.ListVariants {
		t.Fatalf("expected list variants")
	}
}

func TestParseMissingDataError(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	_, err := ParseArgs([]string{}, cfg, env, strings.NewReader(""), true)
	if err == nil {
		t.Fatalf("expected error for missing data")
	}
}

func TestParsePngScaleValidation(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	_, err := ParseArgs([]string{"--png", "--png-scale", "0", "hello"}, cfg, env, strings.NewReader(""), true)
	if err == nil {
		t.Fatalf("expected error for png scale")
	}
}

func TestParseGifFormatMismatch(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	_, err := ParseArgs([]string{"--gif", "--animation-format", "mp4", "hello"}, cfg, env, strings.NewReader(""), true)
	if err == nil {
		t.Fatalf("expected error for gif format mismatch")
	}
}

func TestParseCatalogAnimationConflict(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	_, err := ParseArgs([]string{"--catalog", "--animation", "hello"}, cfg, env, strings.NewReader(""), true)
	if err == nil {
		t.Fatalf("expected error for catalog + animation")
	}
}

func TestParseReadableGifDefaults(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	args, err := ParseArgs([]string{"--animation", "--readable-gif", "hello"}, cfg, env, strings.NewReader(""), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.GifFps != cfg.Defaults.ReadableGif.FPS {
		t.Fatalf("expected readable fps %d, got %d", cfg.Defaults.ReadableGif.FPS, args.GifFps)
	}
	if args.WavePeriod != cfg.Defaults.ReadableGif.WavePeriod {
		t.Fatalf("expected readable wave period %v, got %v", cfg.Defaults.ReadableGif.WavePeriod, args.WavePeriod)
	}
}

func TestParseUnknownVariant(t *testing.T) {
	cfg := testConfig()
	env := &MapEnv{Values: map[string]string{}}
	_, err := ParseArgs([]string{"--variant", "nope", "hello"}, cfg, env, strings.NewReader(""), true)
	if err == nil {
		t.Fatalf("expected error for unknown variant")
	}
}
