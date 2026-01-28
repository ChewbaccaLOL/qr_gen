package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Gradient struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

type Variant struct {
	Name     string    `json:"name"`
	Shape    string    `json:"shape"`
	Dark     string    `json:"dark"`
	Light    *string   `json:"light"`
	Radius   float64   `json:"radius"`
	Gradient *Gradient `json:"gradient"`
}

type GifDefaults struct {
	FPS        int     `json:"gif_fps"`
	Frames     int     `json:"gif_frames"`
	Hold       int     `json:"gif_hold"`
	WaveAmp    float64 `json:"wave_amp"`
	WavePeriod float64 `json:"wave_period"`
}

type Defaults struct {
	GifFPS          int         `json:"gif_fps"`
	GifFrames       int         `json:"gif_frames"`
	GifHold         int         `json:"gif_hold"`
	WaveAmp         float64     `json:"wave_amp"`
	WavePeriod      float64     `json:"wave_period"`
	FloatJaggedSnap float64     `json:"float_jagged_snap"`
	FloatTilt       float64     `json:"float_tilt"`
	FloatAngle      float64     `json:"float_angle"`
	ReadableGif     GifDefaults `json:"readable_gif"`
}

type Config struct {
	Variants          map[string]Variant
	AnimationVariants []string
	Defaults          Defaults
}

type configFile struct {
	Variants          []Variant `json:"variants"`
	AnimationVariants []string  `json:"animation_variants"`
	Defaults          Defaults  `json:"defaults"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read variants config: %w", err)
	}

	var file configFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("parse variants config: %w", err)
	}
	if len(file.Variants) == 0 {
		return nil, errors.New("variants config must include at least one variant")
	}
	if len(file.AnimationVariants) == 0 {
		return nil, errors.New("variants config must include animation_variants")
	}

	variants := make(map[string]Variant, len(file.Variants))
	for _, variant := range file.Variants {
		if variant.Name == "" {
			return nil, errors.New("variant name is required")
		}
		if variant.Shape == "" {
			return nil, fmt.Errorf("variant '%s' is missing a shape", variant.Name)
		}
		if variant.Dark == "" {
			return nil, fmt.Errorf("variant '%s' is missing a dark color", variant.Name)
		}
		if _, exists := variants[variant.Name]; exists {
			return nil, fmt.Errorf("duplicate variant '%s'", variant.Name)
		}
		variants[variant.Name] = variant
	}

	return &Config{
		Variants:          variants,
		AnimationVariants: file.AnimationVariants,
		Defaults:          file.Defaults,
	}, nil
}
