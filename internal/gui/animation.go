package gui

import (
	"errors"
	"fmt"
	"image/gif"
	"strings"

	"qr_generator/internal/animation"
	"qr_generator/internal/config"
	"qr_generator/internal/qr"
)

type AnimationRequest struct {
	RenderRequest
	AnimationVariant string   `json:"animationVariant"`
	GifFps           *int     `json:"gifFps"`
	GifFrames        *int     `json:"gifFrames"`
	GifHold          *int     `json:"gifHold"`
	WaveAmp          *float64 `json:"waveAmp"`
	WavePeriod       *float64 `json:"wavePeriod"`
	FloatAngle       *float64 `json:"floatAngle"`
	FloatCycles      *int     `json:"floatCycles"`
	ReadableGif      bool     `json:"readableGif"`
}

type AnimationConfig struct {
	Variants []string          `json:"variants"`
	Defaults AnimationDefaults `json:"defaults"`
}

type AnimationDefaults struct {
	GifFps      int         `json:"gifFps"`
	GifFrames   int         `json:"gifFrames"`
	GifHold     int         `json:"gifHold"`
	WaveAmp     float64     `json:"waveAmp"`
	WavePeriod  float64     `json:"wavePeriod"`
	FloatAngle  float64     `json:"floatAngle"`
	FloatTilt   float64     `json:"floatTilt"`
	FloatHold   int         `json:"floatHold"`
	FloatCycles int         `json:"floatCycles"`
	Readable    GifDefaults `json:"readable"`
}

type GifDefaults struct {
	GifFps     int     `json:"gifFps"`
	GifFrames  int     `json:"gifFrames"`
	GifHold    int     `json:"gifHold"`
	WaveAmp    float64 `json:"waveAmp"`
	WavePeriod float64 `json:"wavePeriod"`
}

func BuildAnimationConfig(cfg *config.Config) (AnimationConfig, error) {
	if cfg == nil {
		return AnimationConfig{}, errors.New("config is required")
	}
	defaults := AnimationDefaults{
		GifFps:      cfg.Defaults.GifFPS,
		GifFrames:   cfg.Defaults.GifFrames,
		GifHold:     cfg.Defaults.GifHold,
		WaveAmp:     cfg.Defaults.WaveAmp,
		WavePeriod:  cfg.Defaults.WavePeriod,
		FloatAngle:  cfg.Defaults.FloatAngle,
		FloatTilt:   cfg.Defaults.FloatTilt,
		FloatHold:   cfg.Defaults.FloatHold,
		FloatCycles: cfg.Defaults.FloatCycles,
		Readable: GifDefaults{
			GifFps:     cfg.Defaults.ReadableGif.FPS,
			GifFrames:  cfg.Defaults.ReadableGif.Frames,
			GifHold:    cfg.Defaults.ReadableGif.Hold,
			WaveAmp:    cfg.Defaults.ReadableGif.WaveAmp,
			WavePeriod: cfg.Defaults.ReadableGif.WavePeriod,
		},
	}
	return AnimationConfig{
		Variants: append([]string{}, cfg.AnimationVariants...),
		Defaults: defaults,
	}, nil
}

func BuildGIF(cfg *config.Config, req AnimationRequest) (*gif.GIF, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	req.RenderRequest = normalizeRequest(req.RenderRequest)
	if err := validateRequest(req.RenderRequest); err != nil {
		return nil, err
	}
	animationVariant := strings.TrimSpace(req.AnimationVariant)
	if animationVariant == "" {
		animationVariant = "wave"
	}
	if !animationVariantEnabled(cfg, animationVariant) {
		return nil, fmt.Errorf("unknown animation variant '%s'", animationVariant)
	}

	defaults := defaultGifValues(cfg, req.ReadableGif)
	gifFps := chooseInt(req.GifFps, defaults.FPS)
	gifFrames := chooseInt(req.GifFrames, defaults.Frames)
	gifHold := chooseInt(req.GifHold, defaults.Hold)
	waveAmp := chooseFloat(req.WaveAmp, defaults.WaveAmp)
	wavePeriod := chooseFloat(req.WavePeriod, defaults.WavePeriod)

	isFloat := strings.HasPrefix(animationVariant, "float")
	if isFloat && req.GifHold == nil && cfg.Defaults.FloatHold > 0 {
		gifHold = cfg.Defaults.FloatHold
	}

	floatCycles := chooseInt(req.FloatCycles, cfg.Defaults.FloatCycles)
	if floatCycles <= 0 {
		floatCycles = 1
	}

	floatAngle := 0.0
	if isFloat {
		if req.FloatAngle != nil {
			floatAngle = *req.FloatAngle
		} else {
			floatAngle = cfg.Defaults.FloatAngle
			if animationVariant != "float-tilt-first" {
				floatAngle += cfg.Defaults.FloatTilt
			}
		}
	}

	if animationVariant == "wave-loop" {
		gifHold = 0
	}

	style, err := resolveStyle(cfg, req.RenderRequest)
	if err != nil {
		return nil, err
	}
	matrix, err := qr.EncodeMatrix(req.RenderRequest.Data, req.RenderRequest.ErrorLevel)
	if err != nil {
		return nil, err
	}

	switch animationVariant {
	case "wave":
		return animation.BuildWaveGIF(
			matrix,
			req.RenderRequest.Scale,
			req.RenderRequest.Border,
			style.dark,
			style.light,
			style.shape,
			style.radius,
			style.gradient,
			style.bgGradient,
			waveAmp,
			wavePeriod,
			gifFrames,
			gifHold,
			"still",
			gifFps,
			true,
		)
	case "wave-loop":
		return animation.BuildWaveGIF(
			matrix,
			req.RenderRequest.Scale,
			req.RenderRequest.Border,
			style.dark,
			style.light,
			style.shape,
			style.radius,
			style.gradient,
			style.bgGradient,
			waveAmp,
			wavePeriod,
			gifFrames,
			0,
			"loop",
			gifFps,
			true,
		)
	case "float":
		return animation.BuildFloatGIF(
			matrix,
			req.RenderRequest.Scale,
			req.RenderRequest.Border,
			style.dark,
			style.light,
			style.shape,
			style.radius,
			style.gradient,
			style.bgGradient,
			waveAmp,
			wavePeriod,
			floatAngle,
			gifFrames,
			gifHold,
			floatCycles,
			"still",
			0,
			cfg.Defaults.FloatTilt,
			"after",
			gifFps,
			true,
		)
	case "float-tilt-first":
		return animation.BuildFloatGIF(
			matrix,
			req.RenderRequest.Scale,
			req.RenderRequest.Border,
			style.dark,
			style.light,
			style.shape,
			style.radius,
			style.gradient,
			style.bgGradient,
			waveAmp,
			wavePeriod,
			floatAngle,
			gifFrames,
			gifHold,
			floatCycles,
			"still",
			0,
			cfg.Defaults.FloatTilt,
			"before",
			gifFps,
			true,
		)
	case "float-jagged":
		return animation.BuildFloatGIF(
			matrix,
			req.RenderRequest.Scale,
			req.RenderRequest.Border,
			style.dark,
			style.light,
			style.shape,
			style.radius,
			style.gradient,
			style.bgGradient,
			waveAmp,
			wavePeriod,
			floatAngle,
			gifFrames,
			gifHold,
			floatCycles,
			"still",
			cfg.Defaults.FloatJaggedSnap,
			cfg.Defaults.FloatTilt,
			"after",
			gifFps,
			true,
		)
	case "float-tilt-still":
		return animation.BuildFloatGIF(
			matrix,
			req.RenderRequest.Scale,
			req.RenderRequest.Border,
			style.dark,
			style.light,
			style.shape,
			style.radius,
			style.gradient,
			style.bgGradient,
			waveAmp,
			wavePeriod,
			floatAngle,
			gifFrames,
			gifHold,
			floatCycles,
			"still",
			0,
			cfg.Defaults.FloatTilt,
			"after",
			gifFps,
			false,
		)
	default:
		return nil, fmt.Errorf("animation variant '%s' is not supported yet", animationVariant)
	}
}

type gifDefaults struct {
	FPS        int
	Frames     int
	Hold       int
	WaveAmp    float64
	WavePeriod float64
}

func defaultGifValues(cfg *config.Config, readable bool) gifDefaults {
	if readable {
		return gifDefaults{
			FPS:        cfg.Defaults.ReadableGif.FPS,
			Frames:     cfg.Defaults.ReadableGif.Frames,
			Hold:       cfg.Defaults.ReadableGif.Hold,
			WaveAmp:    cfg.Defaults.ReadableGif.WaveAmp,
			WavePeriod: cfg.Defaults.ReadableGif.WavePeriod,
		}
	}
	return gifDefaults{
		FPS:        cfg.Defaults.GifFPS,
		Frames:     cfg.Defaults.GifFrames,
		Hold:       cfg.Defaults.GifHold,
		WaveAmp:    cfg.Defaults.WaveAmp,
		WavePeriod: cfg.Defaults.WavePeriod,
	}
}

func animationVariantEnabled(cfg *config.Config, name string) bool {
	for _, variant := range cfg.AnimationVariants {
		if variant == name {
			return true
		}
	}
	return false
}

func chooseInt(value *int, fallback int) int {
	if value != nil {
		return *value
	}
	return fallback
}

func chooseFloat(value *float64, fallback float64) float64 {
	if value != nil {
		return *value
	}
	return fallback
}
