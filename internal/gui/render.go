package gui

import (
	"errors"
	"fmt"
	"image"
	"math"
	"sort"
	"strings"

	"qr_generator/internal/config"
	"qr_generator/internal/qr"
	"qr_generator/internal/render"
	"qr_generator/internal/renderpng"
)

const (
	defaultVariant  = "classic"
	defaultError    = "m"
	defaultPngScale = 3.0
)

type RenderRequest struct {
	Data         string   `json:"data"`
	Variant      string   `json:"variant"`
	ErrorLevel   string   `json:"errorLevel"`
	Scale        int      `json:"scale"`
	Border       int      `json:"border"`
	Dark         string   `json:"dark"`
	Light        string   `json:"light"`
	NoBackground bool     `json:"noBackground"`
	Radius       *float64 `json:"radius"`
	PngScale     float64  `json:"pngScale"`
}

type VariantInfo struct {
	Name        string  `json:"name"`
	Shape       string  `json:"shape"`
	Dark        string  `json:"dark"`
	Light       *string `json:"light"`
	Radius      float64 `json:"radius"`
	HasGradient bool    `json:"hasGradient"`
}

type resolvedStyle struct {
	variant  config.Variant
	dark     string
	light    *string
	radius   float64
	gradient *config.Gradient
}

func ListVariants(cfg *config.Config) ([]VariantInfo, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if len(cfg.Variants) == 0 {
		return nil, errors.New("variants are required")
	}
	nameList := make([]string, 0, len(cfg.Variants))
	for name := range cfg.Variants {
		nameList = append(nameList, name)
	}
	sort.Strings(nameList)
	out := make([]VariantInfo, 0, len(nameList))
	for _, name := range nameList {
		variant := cfg.Variants[name]
		out = append(out, VariantInfo{
			Name:        variant.Name,
			Shape:       variant.Shape,
			Dark:        variant.Dark,
			Light:       variant.Light,
			Radius:      variant.Radius,
			HasGradient: variant.Gradient != nil,
		})
	}
	return out, nil
}

func BuildSVG(cfg *config.Config, req RenderRequest) (string, error) {
	if cfg == nil {
		return "", errors.New("config is required")
	}
	resolvedReq := normalizeRequest(req)
	if err := validateRequest(resolvedReq); err != nil {
		return "", err
	}
	style, err := resolveStyle(cfg, resolvedReq)
	if err != nil {
		return "", err
	}
	matrix, err := qr.EncodeMatrix(resolvedReq.Data, resolvedReq.ErrorLevel)
	if err != nil {
		return "", err
	}
	return render.RenderSVG(
		matrix,
		resolvedReq.Scale,
		resolvedReq.Border,
		style.dark,
		style.light,
		style.variant.Shape,
		style.radius,
		style.gradient,
		nil,
		0,
		0,
		0,
		"after",
	)
}

func BuildPNG(cfg *config.Config, req RenderRequest) (*image.RGBA, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	resolvedReq := normalizeRequest(req)
	if err := validateRequest(resolvedReq); err != nil {
		return nil, err
	}
	style, err := resolveStyle(cfg, resolvedReq)
	if err != nil {
		return nil, err
	}
	matrix, err := qr.EncodeMatrix(resolvedReq.Data, resolvedReq.ErrorLevel)
	if err != nil {
		return nil, err
	}
	pngScale := resolvedReq.PngScale
	if pngScale <= 0 {
		pngScale = defaultPngScale
	}
	pixelScale := int(math.Round(float64(resolvedReq.Scale) * pngScale))
	if pixelScale <= 0 {
		return nil, errors.New("png scale must be greater than 0")
	}
	return renderpng.RenderPNG(
		matrix,
		pixelScale,
		resolvedReq.Border,
		style.dark,
		style.light,
		style.variant.Shape,
		style.radius,
		style.gradient,
	)
}

func normalizeRequest(req RenderRequest) RenderRequest {
	if strings.TrimSpace(req.Variant) == "" {
		req.Variant = defaultVariant
	}
	if strings.TrimSpace(req.ErrorLevel) == "" {
		req.ErrorLevel = defaultError
	}
	return req
}

func validateRequest(req RenderRequest) error {
	if strings.TrimSpace(req.Data) == "" {
		return errors.New("data is required")
	}
	if req.Scale <= 0 {
		return errors.New("scale must be greater than 0")
	}
	if req.Border < 0 {
		return errors.New("border must be 0 or greater")
	}
	return nil
}

func resolveStyle(cfg *config.Config, req RenderRequest) (resolvedStyle, error) {
	variant, ok := cfg.Variants[req.Variant]
	if !ok {
		return resolvedStyle{}, fmt.Errorf("unknown variant '%s'", req.Variant)
	}
	dark := variant.Dark
	if strings.TrimSpace(req.Dark) != "" {
		dark = req.Dark
	}
	var light *string
	if !req.NoBackground {
		if strings.TrimSpace(req.Light) != "" {
			light = &req.Light
		} else {
			light = variant.Light
		}
	}
	radius := variant.Radius
	if req.Radius != nil {
		radius = *req.Radius
	}
	gradient := variant.Gradient
	if strings.TrimSpace(req.Dark) != "" {
		gradient = nil
	}

	return resolvedStyle{
		variant:  variant,
		dark:     dark,
		light:    light,
		radius:   radius,
		gradient: gradient,
	}, nil
}
