package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"qr_generator/internal/config"
)

const customVariantsFilename = "variants.custom.json"

type CustomVariantRequest struct {
	Name               string   `json:"name"`
	BaseVariant        string   `json:"baseVariant"`
	Dark               string   `json:"dark"`
	Light              string   `json:"light"`
	NoBackground       bool     `json:"noBackground"`
	Radius             *float64 `json:"radius"`
	Shape              string   `json:"shape"`
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
}

type customVariantsFile struct {
	Variants []config.Variant `json:"variants"`
}

func CustomVariantsPath(basePath string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(basePath), customVariantsFilename)
}

func LoadCustomVariants(path string) (map[string]config.Variant, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("custom variants path is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]config.Variant{}, nil
		}
		return nil, fmt.Errorf("read custom variants: %w", err)
	}
	if len(raw) == 0 {
		return map[string]config.Variant{}, nil
	}
	var file customVariantsFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("parse custom variants: %w", err)
	}
	variants := make(map[string]config.Variant, len(file.Variants))
	for _, variant := range file.Variants {
		if err := validateVariant(variant); err != nil {
			return nil, err
		}
		if _, exists := variants[variant.Name]; exists {
			return nil, fmt.Errorf("duplicate custom variant '%s'", variant.Name)
		}
		variants[variant.Name] = variant
	}
	return variants, nil
}

func SaveCustomVariants(path string, variants map[string]config.Variant) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("custom variants path is required")
	}
	if len(variants) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove custom variants: %w", err)
		}
		return nil
	}
	names := make([]string, 0, len(variants))
	for name := range variants {
		names = append(names, name)
	}
	sort.Strings(names)
	list := make([]config.Variant, 0, len(names))
	for _, name := range names {
		variant := variants[name]
		variant.Name = name
		if err := validateVariant(variant); err != nil {
			return err
		}
		list = append(list, variant)
	}
	payload := customVariantsFile{Variants: list}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal custom variants: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create custom variants dir: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write custom variants: %w", err)
	}
	return nil
}

func MergeVariants(baseVariants map[string]config.Variant, customVariants map[string]config.Variant) (map[string]config.Variant, error) {
	if len(baseVariants) == 0 {
		return nil, errors.New("base variants are required")
	}
	merged := make(map[string]config.Variant, len(baseVariants)+len(customVariants))
	for name, variant := range baseVariants {
		merged[name] = variant
	}
	for name, variant := range customVariants {
		if _, exists := merged[name]; exists {
			return nil, fmt.Errorf("custom variant '%s' conflicts with base", name)
		}
		merged[name] = variant
	}
	return merged, nil
}

func validateVariant(variant config.Variant) error {
	if strings.TrimSpace(variant.Name) == "" {
		return errors.New("variant name is required")
	}
	if strings.TrimSpace(variant.Shape) == "" {
		return fmt.Errorf("variant '%s' is missing a shape", variant.Name)
	}
	if strings.TrimSpace(variant.Dark) == "" {
		return fmt.Errorf("variant '%s' is missing a dark color", variant.Name)
	}
	return nil
}
