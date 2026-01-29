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
	Data               string   `json:"data"`
	Variant            string   `json:"variant"`
	Shape              string   `json:"shape"`
	ErrorLevel         string   `json:"errorLevel"`
	Scale              int      `json:"scale"`
	Border             int      `json:"border"`
	Dark               string   `json:"dark"`
	Light              string   `json:"light"`
	NoBackground       bool     `json:"noBackground"`
	Radius             *float64 `json:"radius"`
	Gradient           bool     `json:"gradientEnabled"`
	GradientFrom       string   `json:"gradientFrom"`
	GradientTo         string   `json:"gradientTo"`
	GradientAngle      *float64 `json:"gradientAngle"`
	GradientFromStop   *float64 `json:"gradientFromStop"`
	GradientToStop     *float64 `json:"gradientToStop"`
	GradientScope      string   `json:"gradientScope"`
	BgGradient         bool     `json:"bgGradientEnabled"`
	BgGradientFrom     string   `json:"bgGradientFrom"`
	BgGradientTo       string   `json:"bgGradientTo"`
	BgGradientAngle    *float64 `json:"bgGradientAngle"`
	BgGradientFromStop *float64 `json:"bgGradientFromStop"`
	BgGradientToStop   *float64 `json:"bgGradientToStop"`
	PngScale           float64  `json:"pngScale"`
}

type VariantInfo struct {
	Name               string   `json:"name"`
	Shape              string   `json:"shape"`
	Dark               string   `json:"dark"`
	Light              *string  `json:"light"`
	Radius             float64  `json:"radius"`
	HasGradient        bool     `json:"hasGradient"`
	GradientFrom       *string  `json:"gradientFrom"`
	GradientTo         *string  `json:"gradientTo"`
	GradientAngle      *float64 `json:"gradientAngle"`
	GradientFromStop   *float64 `json:"gradientFromStop"`
	GradientToStop     *float64 `json:"gradientToStop"`
	GradientScope      string   `json:"gradientScope"`
	HasBgGradient      bool     `json:"hasBgGradient"`
	BgGradientFrom     *string  `json:"bgGradientFrom"`
	BgGradientTo       *string  `json:"bgGradientTo"`
	BgGradientAngle    *float64 `json:"bgGradientAngle"`
	BgGradientFromStop *float64 `json:"bgGradientFromStop"`
	BgGradientToStop   *float64 `json:"bgGradientToStop"`
	IsCustom           bool     `json:"isCustom"`
}

type resolvedStyle struct {
	variant    config.Variant
	dark       string
	light      *string
	radius     float64
	shape      string
	gradient   *config.Gradient
	bgGradient *config.Gradient
}

func ListVariants(baseVariants map[string]config.Variant, customVariants map[string]config.Variant) ([]VariantInfo, error) {
	if len(baseVariants) == 0 {
		return nil, errors.New("base variants are required")
	}
	if customVariants == nil {
		customVariants = map[string]config.Variant{}
	}
	nameList := config.EnabledVariantNames(baseVariants)
	out := make([]VariantInfo, 0, len(nameList)+len(customVariants))
	for _, name := range nameList {
		variant := baseVariants[name]
		var gradientFrom *string
		var gradientTo *string
		var gradientAngle *float64
		var gradientFromStop *float64
		var gradientToStop *float64
		gradientScope := ""
		if variant.Gradient != nil {
			from := variant.Gradient.From
			to := variant.Gradient.To
			gradientFrom = &from
			gradientTo = &to
			gradientAngle = variant.Gradient.Angle
			gradientFromStop = variant.Gradient.FromStop
			gradientToStop = variant.Gradient.ToStop
			gradientScope = variant.Gradient.Scope
		}
		var bgGradientFrom *string
		var bgGradientTo *string
		var bgGradientAngle *float64
		var bgGradientFromStop *float64
		var bgGradientToStop *float64
		if variant.BackgroundGradient != nil {
			from := variant.BackgroundGradient.From
			to := variant.BackgroundGradient.To
			bgGradientFrom = &from
			bgGradientTo = &to
			bgGradientAngle = variant.BackgroundGradient.Angle
			bgGradientFromStop = variant.BackgroundGradient.FromStop
			bgGradientToStop = variant.BackgroundGradient.ToStop
		}
		out = append(out, VariantInfo{
			Name:               variant.Name,
			Shape:              variant.Shape,
			Dark:               variant.Dark,
			Light:              variant.Light,
			Radius:             variant.Radius,
			HasGradient:        variant.Gradient != nil,
			GradientFrom:       gradientFrom,
			GradientTo:         gradientTo,
			GradientAngle:      gradientAngle,
			GradientFromStop:   gradientFromStop,
			GradientToStop:     gradientToStop,
			GradientScope:      gradientScope,
			HasBgGradient:      variant.BackgroundGradient != nil,
			BgGradientFrom:     bgGradientFrom,
			BgGradientTo:       bgGradientTo,
			BgGradientAngle:    bgGradientAngle,
			BgGradientFromStop: bgGradientFromStop,
			BgGradientToStop:   bgGradientToStop,
			IsCustom:           false,
		})
	}
	if len(customVariants) == 0 {
		return out, nil
	}
	customNames := make([]string, 0, len(customVariants))
	for name, variant := range customVariants {
		if variant.Disabled {
			continue
		}
		if _, exists := baseVariants[name]; exists {
			return nil, fmt.Errorf("custom variant '%s' conflicts with base", name)
		}
		customNames = append(customNames, name)
	}
	sort.Strings(customNames)
	for _, name := range customNames {
		variant := customVariants[name]
		var gradientFrom *string
		var gradientTo *string
		var gradientAngle *float64
		var gradientFromStop *float64
		var gradientToStop *float64
		gradientScope := ""
		if variant.Gradient != nil {
			from := variant.Gradient.From
			to := variant.Gradient.To
			gradientFrom = &from
			gradientTo = &to
			gradientAngle = variant.Gradient.Angle
			gradientFromStop = variant.Gradient.FromStop
			gradientToStop = variant.Gradient.ToStop
			gradientScope = variant.Gradient.Scope
		}
		var bgGradientFrom *string
		var bgGradientTo *string
		var bgGradientAngle *float64
		var bgGradientFromStop *float64
		var bgGradientToStop *float64
		if variant.BackgroundGradient != nil {
			from := variant.BackgroundGradient.From
			to := variant.BackgroundGradient.To
			bgGradientFrom = &from
			bgGradientTo = &to
			bgGradientAngle = variant.BackgroundGradient.Angle
			bgGradientFromStop = variant.BackgroundGradient.FromStop
			bgGradientToStop = variant.BackgroundGradient.ToStop
		}
		out = append(out, VariantInfo{
			Name:               variant.Name,
			Shape:              variant.Shape,
			Dark:               variant.Dark,
			Light:              variant.Light,
			Radius:             variant.Radius,
			HasGradient:        variant.Gradient != nil,
			GradientFrom:       gradientFrom,
			GradientTo:         gradientTo,
			GradientAngle:      gradientAngle,
			GradientFromStop:   gradientFromStop,
			GradientToStop:     gradientToStop,
			GradientScope:      gradientScope,
			HasBgGradient:      variant.BackgroundGradient != nil,
			BgGradientFrom:     bgGradientFrom,
			BgGradientTo:       bgGradientTo,
			BgGradientAngle:    bgGradientAngle,
			BgGradientFromStop: bgGradientFromStop,
			BgGradientToStop:   bgGradientToStop,
			IsCustom:           true,
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
		style.shape,
		style.radius,
		style.gradient,
		style.bgGradient,
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
		style.shape,
		style.radius,
		style.gradient,
		style.bgGradient,
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
	shape := variant.Shape
	if strings.TrimSpace(req.Shape) != "" {
		shape = req.Shape
	}
	gradient := variant.Gradient
	if req.Gradient {
		gradient = buildGradientOverride(variant.Gradient, dark, req.GradientFrom, req.GradientTo, req.GradientAngle, req.GradientFromStop, req.GradientToStop, req.GradientScope, cfg.Defaults)
	} else {
		gradient = nil
	}
	bgGradient := variant.BackgroundGradient
	if req.BgGradient && !req.NoBackground {
		bgGradient = buildBackgroundGradientOverride(
			variant.BackgroundGradient,
			req.BgGradientFrom,
			req.BgGradientTo,
			req.BgGradientAngle,
			req.BgGradientFromStop,
			req.BgGradientToStop,
			cfg.Defaults,
		)
	} else {
		bgGradient = nil
	}

	return resolvedStyle{
		variant:    variant,
		dark:       dark,
		light:      light,
		radius:     radius,
		shape:      shape,
		gradient:   gradient,
		bgGradient: bgGradient,
	}, nil
}

func buildGradientOverride(
	base *config.Gradient,
	dark string,
	from string,
	to string,
	angle *float64,
	fromStop *float64,
	toStop *float64,
	scope string,
	defaults config.Defaults,
) *config.Gradient {
	gradient := &config.Gradient{}
	if base != nil {
		gradient.ID = base.ID
		gradient.From = base.From
		gradient.To = base.To
		gradient.Angle = base.Angle
		gradient.FromStop = base.FromStop
		gradient.ToStop = base.ToStop
		gradient.Scope = base.Scope
	} else {
		gradient.From = dark
		gradient.To = dark
	}
	if strings.TrimSpace(from) != "" {
		gradient.From = from
	}
	if strings.TrimSpace(to) != "" {
		gradient.To = to
	}
	if angle != nil {
		gradient.Angle = angle
	} else if gradient.Angle == nil {
		gradient.Angle = floatPtr(defaults.GradientAngle)
	}
	if fromStop != nil {
		gradient.FromStop = fromStop
	} else if gradient.FromStop == nil {
		gradient.FromStop = floatPtr(defaults.GradientFromStop)
	}
	if toStop != nil {
		gradient.ToStop = toStop
	} else if gradient.ToStop == nil {
		gradient.ToStop = floatPtr(defaults.GradientToStop)
	}
	if strings.TrimSpace(scope) != "" {
		gradient.Scope = scope
	} else if strings.TrimSpace(gradient.Scope) == "" {
		gradient.Scope = defaults.GradientScope
	}
	return gradient
}

func buildBackgroundGradientOverride(
	base *config.Gradient,
	from string,
	to string,
	angle *float64,
	fromStop *float64,
	toStop *float64,
	defaults config.Defaults,
) *config.Gradient {
	gradient := &config.Gradient{}
	if base != nil {
		gradient.ID = base.ID
		gradient.From = base.From
		gradient.To = base.To
		gradient.Angle = base.Angle
		gradient.FromStop = base.FromStop
		gradient.ToStop = base.ToStop
		gradient.Scope = base.Scope
	}
	if strings.TrimSpace(from) != "" {
		gradient.From = from
	}
	if strings.TrimSpace(to) != "" {
		gradient.To = to
	}
	if angle != nil {
		gradient.Angle = angle
	} else if gradient.Angle == nil {
		gradient.Angle = floatPtr(defaults.BgGradientAngle)
	}
	if fromStop != nil {
		gradient.FromStop = fromStop
	} else if gradient.FromStop == nil {
		gradient.FromStop = floatPtr(defaults.BgGradientFromStop)
	}
	if toStop != nil {
		gradient.ToStop = toStop
	} else if gradient.ToStop == nil {
		gradient.ToStop = floatPtr(defaults.BgGradientToStop)
	}
	if strings.TrimSpace(gradient.Scope) == "" {
		gradient.Scope = "global"
	}
	return gradient
}

func floatPtr(value float64) *float64 {
	return &value
}
