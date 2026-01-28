package main

import (
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
	ctx          context.Context
	cfg          *config.Config
	variantsPath string
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
	a.cfg = cfg
	a.variantsPath = variantsPath
}

func (a *App) GetVariantCatalog() ([]guicore.VariantInfo, error) {
	if a.cfg == nil {
		return nil, errors.New("variants config not loaded")
	}
	return guicore.ListVariants(a.cfg)
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
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save QR",
		DefaultFilename: defaultName,
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

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
