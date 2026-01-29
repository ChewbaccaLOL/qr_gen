package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"qr_generator/internal/config"
	guicore "qr_generator/internal/gui"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                context.Context
	cfg                *config.Config
	variantsPath       string
	customVariantsPath string
	baseVariants       map[string]config.Variant
	customVariants     map[string]config.Variant
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cwd, _ := os.Getwd()
	exePath, _ := os.Executable()
	variantsPath, err := guicore.FindVariantsPath(cwd, exePath)
	if err != nil {
		runtime.LogError(ctx, err.Error())
		return
	}
	cfg, err := config.Load(variantsPath)
	if err != nil {
		runtime.LogError(ctx, err.Error())
		return
	}
	a.baseVariants = copyVariants(cfg.Variants)
	a.customVariantsPath = guicore.CustomVariantsPath(variantsPath)
	customVariants, err := guicore.LoadCustomVariants(a.customVariantsPath)
	if err != nil {
		runtime.LogError(ctx, err.Error())
		customVariants = map[string]config.Variant{}
	}
	mergedVariants, err := guicore.MergeVariants(a.baseVariants, customVariants)
	if err != nil {
		runtime.LogError(ctx, err.Error())
		mergedVariants = a.baseVariants
		customVariants = map[string]config.Variant{}
	}
	cfg.Variants = mergedVariants
	a.cfg = cfg
	a.variantsPath = variantsPath
	a.customVariants = customVariants
}

func (a *App) GetVariantCatalog() ([]guicore.VariantInfo, error) {
	if a.cfg == nil {
		return nil, errors.New("variants config not loaded")
	}
	return guicore.ListVariants(a.baseVariants, a.customVariants)
}

func (a *App) IsWSL() bool {
	return guicore.IsWSL()
}

func (a *App) GenerateSVG(req guicore.RenderRequest) (string, error) {
	if a.cfg == nil {
		return "", errors.New("variants config not loaded")
	}
	svg, err := guicore.BuildSVG(a.cfg, req)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(svg)), nil
}

func (a *App) SaveSVG(req guicore.RenderRequest, outputPath string) (string, error) {
	if a.cfg == nil {
		return "", errors.New("variants config not loaded")
	}
	path := strings.TrimSpace(outputPath)
	if path == "" {
		return "", errors.New("output path is required")
	}
	svg, err := guicore.BuildSVG(a.cfg, req)
	if err != nil {
		return "", err
	}
	if err := ensureParentDir(path); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(svg), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) SavePNG(req guicore.RenderRequest, outputPath string) (string, error) {
	if a.cfg == nil {
		return "", errors.New("variants config not loaded")
	}
	path := strings.TrimSpace(outputPath)
	if path == "" {
		return "", errors.New("output path is required")
	}
	imageOut, err := guicore.BuildPNG(a.cfg, req)
	if err != nil {
		return "", err
	}
	if err := ensureParentDir(path); err != nil {
		return "", err
	}
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := png.Encode(file, imageOut); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) GeneratePNG(req guicore.RenderRequest) (string, error) {
	if a.cfg == nil {
		return "", errors.New("variants config not loaded")
	}
	imageOut, err := guicore.BuildPNG(a.cfg, req)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imageOut); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func (a *App) SuggestSavePath(format string) (string, error) {
	if a.ctx == nil {
		return "", errors.New("app not initialized")
	}
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		format = "svg"
	}
	filterPattern := fmt.Sprintf("*.%s", format)
	defaultName := fmt.Sprintf("qr.%s", format)
	defaultDir := ""
	if cwd, err := os.Getwd(); err == nil {
		candidate := guicore.DefaultExportDir(cwd)
		if candidate != "" {
			if err := os.MkdirAll(candidate, 0o755); err == nil {
				defaultDir = candidate
			}
		}
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Save QR",
		DefaultDirectory: defaultDir,
		DefaultFilename:  defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: strings.ToUpper(format), Pattern: filterPattern},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	return path, nil
}

func (a *App) GetVariantsPath() string {
	return a.variantsPath
}

func (a *App) SaveCustomVariant(req guicore.CustomVariantRequest) ([]guicore.VariantInfo, error) {
	if a.cfg == nil {
		return nil, errors.New("variants config not loaded")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("custom variant name is required")
	}
	if _, exists := a.baseVariants[name]; exists {
		return nil, fmt.Errorf("variant '%s' is built-in and cannot be replaced", name)
	}
	if _, exists := a.customVariants[name]; exists {
		return nil, fmt.Errorf("custom variant '%s' already exists", name)
	}
	baseName := strings.TrimSpace(req.BaseVariant)
	if baseName == "" {
		return nil, errors.New("base variant is required")
	}
	baseVariant, ok := a.cfg.Variants[baseName]
	if !ok {
		return nil, fmt.Errorf("unknown base variant '%s'", baseName)
	}
	customVariant := baseVariant
	customVariant.Name = name
	if strings.TrimSpace(req.Dark) != "" {
		customVariant.Dark = req.Dark
		customVariant.Gradient = nil
	}
	if req.NoBackground {
		customVariant.Light = nil
	} else if strings.TrimSpace(req.Light) != "" {
		light := req.Light
		customVariant.Light = &light
	} else {
		customVariant.Light = baseVariant.Light
	}
	if req.Radius != nil {
		customVariant.Radius = *req.Radius
	}
	if strings.TrimSpace(req.Shape) != "" {
		customVariant.Shape = req.Shape
	}
	if req.Gradient {
		customVariant.Gradient = buildCustomGradient(req, customVariant.Gradient, customVariant.Dark, a.cfg.Defaults, customVariant.Name)
	} else {
		customVariant.Gradient = nil
	}
	if req.BgGradient && !req.NoBackground {
		customVariant.BackgroundGradient = buildCustomBackgroundGradient(req, customVariant.BackgroundGradient, a.cfg.Defaults, customVariant.Name)
	} else {
		customVariant.BackgroundGradient = nil
	}
	candidate := copyVariants(a.customVariants)
	candidate[name] = customVariant
	if err := guicore.SaveCustomVariants(a.customVariantsPath, candidate); err != nil {
		return nil, err
	}
	merged, err := guicore.MergeVariants(a.baseVariants, candidate)
	if err != nil {
		return nil, err
	}
	a.customVariants = candidate
	a.cfg.Variants = merged
	return guicore.ListVariants(a.baseVariants, a.customVariants)
}

func (a *App) DeleteCustomVariant(name string) ([]guicore.VariantInfo, error) {
	if a.cfg == nil {
		return nil, errors.New("variants config not loaded")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("custom variant name is required")
	}
	if _, exists := a.baseVariants[name]; exists {
		return nil, fmt.Errorf("variant '%s' is built-in and cannot be deleted", name)
	}
	if _, exists := a.customVariants[name]; !exists {
		return nil, fmt.Errorf("custom variant '%s' not found", name)
	}
	candidate := copyVariants(a.customVariants)
	delete(candidate, name)
	if err := guicore.SaveCustomVariants(a.customVariantsPath, candidate); err != nil {
		return nil, err
	}
	merged, err := guicore.MergeVariants(a.baseVariants, candidate)
	if err != nil {
		return nil, err
	}
	a.customVariants = candidate
	a.cfg.Variants = merged
	return guicore.ListVariants(a.baseVariants, a.customVariants)
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func copyVariants(input map[string]config.Variant) map[string]config.Variant {
	if len(input) == 0 {
		return map[string]config.Variant{}
	}
	out := make(map[string]config.Variant, len(input))
	for name, variant := range input {
		out[name] = variant
	}
	return out
}

func chooseGradientColor(override string, base *config.Gradient, fallback string, useFrom bool) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	if base != nil {
		if useFrom {
			return base.From
		}
		return base.To
	}
	return fallback
}

func buildCustomGradient(req guicore.CustomVariantRequest, base *config.Gradient, fallback string, defaults config.Defaults, id string) *config.Gradient {
	gradient := &config.Gradient{ID: id}
	if base != nil {
		gradient.From = base.From
		gradient.To = base.To
		gradient.Angle = base.Angle
		gradient.FromStop = base.FromStop
		gradient.ToStop = base.ToStop
		gradient.Scope = base.Scope
	}
	gradient.From = chooseGradientColor(req.GradientFrom, base, fallback, true)
	gradient.To = chooseGradientColor(req.GradientTo, base, fallback, false)
	gradient.Angle = chooseGradientFloat(req.GradientAngle, gradient.Angle, defaults.GradientAngle)
	gradient.FromStop = chooseGradientFloat(req.GradientFromStop, gradient.FromStop, defaults.GradientFromStop)
	gradient.ToStop = chooseGradientFloat(req.GradientToStop, gradient.ToStop, defaults.GradientToStop)
	gradient.Scope = chooseGradientScope(req.GradientScope, gradient.Scope, defaults.GradientScope, "module")
	return gradient
}

func buildCustomBackgroundGradient(req guicore.CustomVariantRequest, base *config.Gradient, defaults config.Defaults, id string) *config.Gradient {
	gradient := &config.Gradient{ID: id}
	if base != nil {
		gradient.From = base.From
		gradient.To = base.To
		gradient.Angle = base.Angle
		gradient.FromStop = base.FromStop
		gradient.ToStop = base.ToStop
		gradient.Scope = base.Scope
	}
	if strings.TrimSpace(req.BgGradientFrom) != "" {
		gradient.From = req.BgGradientFrom
	}
	if strings.TrimSpace(req.BgGradientTo) != "" {
		gradient.To = req.BgGradientTo
	}
	gradient.Angle = chooseGradientFloat(req.BgGradientAngle, gradient.Angle, defaults.BgGradientAngle)
	gradient.FromStop = chooseGradientFloat(req.BgGradientFromStop, gradient.FromStop, defaults.BgGradientFromStop)
	gradient.ToStop = chooseGradientFloat(req.BgGradientToStop, gradient.ToStop, defaults.BgGradientToStop)
	gradient.Scope = "global"
	return gradient
}

func chooseGradientFloat(override *float64, base *float64, fallback float64) *float64 {
	if override != nil {
		value := *override
		return &value
	}
	if base != nil {
		value := *base
		return &value
	}
	return &fallback
}

func chooseGradientScope(override string, base string, fallback string, defaultScope string) string {
	scope := strings.TrimSpace(override)
	if scope == "" {
		scope = strings.TrimSpace(base)
	}
	if scope == "" {
		scope = strings.TrimSpace(fallback)
	}
	if scope == "" {
		scope = defaultScope
	}
	switch strings.ToLower(scope) {
	case "global":
		return "global"
	default:
		return "module"
	}
}
