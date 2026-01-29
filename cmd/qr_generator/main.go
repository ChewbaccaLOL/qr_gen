package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"qr_generator/internal/animation"
	"qr_generator/internal/cli"
	"qr_generator/internal/config"
	"qr_generator/internal/qr"
	"qr_generator/internal/render"
	"qr_generator/internal/renderpdf"
	"qr_generator/internal/renderpng"
	"qr_generator/internal/renderps"
)

const (
	exitUsage = 2
)

func main() {
	env := cli.OSEnv{}
	cli.LoadDotenv(".env", env)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: unable to resolve working directory")
		os.Exit(exitUsage)
	}
	configPath := filepath.Join(cwd, "variants.json")

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}

	args, err := cli.ParseArgs(os.Args[1:], cfg, env, os.Stdin, isTerminal(os.Stdin))
	if err != nil {
		if usageErr, ok := err.(cli.ErrUsage); ok {
			fmt.Fprintln(os.Stderr, usageErr.Usage)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}

	if args.ListVariants {
		names := config.EnabledVariantNames(cfg.Variants)
		for _, name := range names {
			fmt.Println(name)
		}
		fmt.Println()
		fmt.Println("Animations:")
		for _, name := range cfg.AnimationVariants {
			fmt.Println(name)
		}
		return
	}

	animationEnabled := args.Animation || args.Gif

	matrix, err := qr.EncodeMatrix(args.Data, args.ErrorLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}

	var svg string
	var variant config.Variant
	var dark string
	var light *string
	var radius float64
	var gradient *config.Gradient
	var backgroundGradient *config.Gradient
	if args.Catalog {
		variants := config.EnabledVariants(cfg.Variants)
		svg, err = render.RenderCatalogSVG(
			matrix,
			args.Scale,
			args.Border,
			variants,
			args.CatalogColumns,
			args.CatalogBackground,
			args.CatalogLabelSize,
		)
	} else {
		variant = cfg.Variants[args.Variant]
		dark = variant.Dark
		if args.Dark != "" {
			dark = args.Dark
		}
		if !args.NoBackground {
			if args.Light != "" {
				light = &args.Light
			} else {
				light = variant.Light
			}
		}
		radius = variant.Radius
		if args.Radius != nil {
			radius = *args.Radius
		}
		gradient = resolveForegroundGradient(args, variant, dark, cfg.Defaults)
		if args.NoBackground {
			backgroundGradient = nil
		} else {
			backgroundGradient = resolveBackgroundGradient(args, variant, light, cfg.Defaults)
		}
		svg, err = render.RenderSVG(
			matrix,
			args.Scale,
			args.Border,
			dark,
			light,
			variant.Shape,
			radius,
			gradient,
			backgroundGradient,
			nil,
			0,
			0,
			0,
			"after",
		)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}

	if err := ensureParentDir(args.Output); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}
	if err := os.WriteFile(args.Output, []byte(svg), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}
	fmt.Printf("Saved %s\n", args.Output)

	if args.Png {
		scale := int(math.Round(float64(args.Scale) * args.PngScale))
		if scale <= 0 {
			fmt.Fprintln(os.Stderr, "error: --png-scale must be greater than 0")
			os.Exit(exitUsage)
		}
		pngPath := args.PngOutput
		if pngPath == "" {
			pngPath = deriveOutputPath(args.Output, ".png")
		}
		var pngImage *image.RGBA
		if args.Catalog {
			labelSize := args.CatalogLabelSize
			if labelSize > 0 {
				labelSize = int(math.Round(float64(labelSize) * args.PngScale))
			}
			names := make([]string, 0, len(cfg.Variants))
			for name := range cfg.Variants {
				names = append(names, name)
			}
			sort.Strings(names)
			var variants []config.Variant
			for _, name := range names {
				variants = append(variants, cfg.Variants[name])
			}
			pngImage, err = renderpng.RenderCatalogPNG(
				matrix,
				scale,
				args.Border,
				variants,
				args.CatalogColumns,
				args.CatalogBackground,
				labelSize,
			)
		} else {
			pngImage, err = renderpng.RenderPNG(
				matrix,
				scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
			)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		if err := ensureParentDir(pngPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		file, err := os.Create(pngPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		defer file.Close()
		if err := png.Encode(file, pngImage); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		fmt.Printf("Saved %s\n", pngPath)
	}

	if args.Pdf {
		pdfPath := args.PdfOutput
		if pdfPath == "" {
			pdfPath = deriveOutputPath(args.Output, ".pdf")
		}
		if err := ensureParentDir(pdfPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		var doc *renderpdf.PDFDocument
		if args.Catalog {
			names := make([]string, 0, len(cfg.Variants))
			for name := range cfg.Variants {
				names = append(names, name)
			}
			sort.Strings(names)
			var variants []config.Variant
			for _, name := range names {
				variants = append(variants, cfg.Variants[name])
			}
			doc, err = renderpdf.RenderCatalogPDF(
				matrix,
				args.Scale,
				args.Border,
				variants,
				args.CatalogColumns,
				args.CatalogBackground,
				args.CatalogLabelSize,
			)
		} else {
			doc, err = renderpdf.RenderPDF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
			)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		if err := doc.Write(pdfPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		fmt.Printf("Saved %s\n", pdfPath)
	}

	if args.Ps {
		psPath := args.PsOutput
		if psPath == "" {
			psPath = deriveOutputPath(args.Output, ".ps")
		}
		if err := ensureParentDir(psPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		var doc *renderps.PSDocument
		if args.Catalog {
			names := make([]string, 0, len(cfg.Variants))
			for name := range cfg.Variants {
				names = append(names, name)
			}
			sort.Strings(names)
			var variants []config.Variant
			for _, name := range names {
				variants = append(variants, cfg.Variants[name])
			}
			doc, err = renderps.RenderCatalogPS(
				matrix,
				args.Scale,
				args.Border,
				variants,
				args.CatalogColumns,
				args.CatalogBackground,
				args.CatalogLabelSize,
			)
		} else {
			doc, err = renderps.RenderPS(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
			)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		if err := doc.Write(psPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		fmt.Printf("Saved %s\n", psPath)
	}

	if animationEnabled {
		if args.AnimationFormat != "gif" {
			fmt.Fprintf(os.Stderr, "error: animation format '%s' is not supported yet\n", args.AnimationFormat)
			os.Exit(exitUsage)
		}
		gifPath := args.GifOutput
		if gifPath == "" {
			gifPath = deriveOutputPath(args.Output, ".gif")
		}
		if err := ensureParentDir(gifPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}

		variant := cfg.Variants[args.Variant]
		dark := variant.Dark
		if args.Dark != "" {
			dark = args.Dark
		}
		var light *string
		if !args.NoBackground {
			if args.Light != "" {
				light = &args.Light
			} else {
				light = variant.Light
			}
		}
		radius := variant.Radius
		if args.Radius != nil {
			radius = *args.Radius
		}
		gradient := resolveForegroundGradient(args, variant, dark, cfg.Defaults)
		var backgroundGradient *config.Gradient
		if !args.NoBackground {
			backgroundGradient = resolveBackgroundGradient(args, variant, light, cfg.Defaults)
		}

		var gifOut *gif.GIF
		switch args.AnimationVariant {
		case "wave":
			gifOut, err = animation.BuildWaveGIF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
				args.WaveAmp,
				args.WavePeriod,
				args.GifFrames,
				args.GifHold,
				"still",
				args.GifFps,
				true,
			)
		case "wave-loop":
			gifOut, err = animation.BuildWaveGIF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
				args.WaveAmp,
				args.WavePeriod,
				args.GifFrames,
				0,
				"loop",
				args.GifFps,
				true,
			)
		case "float":
			floatAngle := args.FloatAngle
			if floatAngle == nil {
				angle := cfg.Defaults.FloatAngle + cfg.Defaults.FloatTilt
				floatAngle = &angle
			}
			gifOut, err = animation.BuildFloatGIF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
				args.FloatCycles,
				"still",
				0,
				cfg.Defaults.FloatTilt,
				"after",
				args.GifFps,
				true,
			)
		case "float-tilt-first":
			floatAngle := args.FloatAngle
			if floatAngle == nil {
				angle := cfg.Defaults.FloatAngle
				floatAngle = &angle
			}
			gifOut, err = animation.BuildFloatGIF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
				args.FloatCycles,
				"still",
				0,
				cfg.Defaults.FloatTilt,
				"before",
				args.GifFps,
				true,
			)
		case "float-jagged":
			floatAngle := args.FloatAngle
			if floatAngle == nil {
				angle := cfg.Defaults.FloatAngle + cfg.Defaults.FloatTilt
				floatAngle = &angle
			}
			gifOut, err = animation.BuildFloatGIF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
				args.FloatCycles,
				"still",
				cfg.Defaults.FloatJaggedSnap,
				cfg.Defaults.FloatTilt,
				"after",
				args.GifFps,
				true,
			)
		case "float-tilt-still":
			floatAngle := args.FloatAngle
			if floatAngle == nil {
				angle := cfg.Defaults.FloatAngle + cfg.Defaults.FloatTilt
				floatAngle = &angle
			}
			gifOut, err = animation.BuildFloatGIF(
				matrix,
				args.Scale,
				args.Border,
				dark,
				light,
				variant.Shape,
				radius,
				gradient,
				backgroundGradient,
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
				args.FloatCycles,
				"still",
				0,
				cfg.Defaults.FloatTilt,
				"after",
				args.GifFps,
				false,
			)
		default:
			fmt.Fprintf(os.Stderr, "error: animation variant '%s' is not supported yet\n", args.AnimationVariant)
			os.Exit(exitUsage)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		file, err := os.Create(gifPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		defer file.Close()
		if err := gif.EncodeAll(file, gifOut); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
		}
		fmt.Printf("Saved %s\n", gifPath)
	}
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return true
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func deriveOutputPath(path string, extension string) string {
	ext := extension
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".svg") {
		return path[:len(path)-4] + ext
	}
	return path + ext
}

func resolveForegroundGradient(args *cli.Args, variant config.Variant, dark string, defaults config.Defaults) *config.Gradient {
	if args.GradientDisabled {
		return nil
	}
	enable := args.GradientEnabled || args.GradientConfigSet
	if !enable {
		if variant.Gradient == nil {
			return nil
		}
		if args.Dark != "" {
			return nil
		}
	}
	base := variant.Gradient
	gradient := &config.Gradient{}
	if base != nil {
		*gradient = *base
	}
	if gradient.ID == "" {
		gradient.ID = "fg"
	}
	if strings.TrimSpace(args.GradientFrom) != "" {
		gradient.From = args.GradientFrom
	} else if strings.TrimSpace(gradient.From) == "" {
		gradient.From = dark
	}
	if strings.TrimSpace(args.GradientTo) != "" {
		gradient.To = args.GradientTo
	} else if strings.TrimSpace(gradient.To) == "" {
		gradient.To = dark
	}
	gradient.Angle = chooseFloat(args.GradientAngle, gradient.Angle, defaults.GradientAngle)
	gradient.FromStop = chooseFloat(args.GradientFromStop, gradient.FromStop, defaults.GradientFromStop)
	gradient.ToStop = chooseFloat(args.GradientToStop, gradient.ToStop, defaults.GradientToStop)
	gradient.Scope = chooseScope(args.GradientScope, gradient.Scope, defaults.GradientScope, "module")
	return gradient
}

func resolveBackgroundGradient(args *cli.Args, variant config.Variant, light *string, defaults config.Defaults) *config.Gradient {
	if args.BgGradientDisabled {
		return nil
	}
	enable := args.BgGradientEnabled || args.BgGradientConfigSet || variant.BackgroundGradient != nil
	if !enable {
		return nil
	}
	base := variant.BackgroundGradient
	gradient := &config.Gradient{}
	if base != nil {
		*gradient = *base
	}
	if gradient.ID == "" {
		gradient.ID = "bg"
	}
	if strings.TrimSpace(args.BgGradientFrom) != "" {
		gradient.From = args.BgGradientFrom
	} else if strings.TrimSpace(gradient.From) == "" {
		gradient.From = fallbackLight(light, "#ffffff")
	}
	if strings.TrimSpace(args.BgGradientTo) != "" {
		gradient.To = args.BgGradientTo
	} else if strings.TrimSpace(gradient.To) == "" {
		gradient.To = fallbackLight(light, "#ffffff")
	}
	gradient.Angle = chooseFloat(args.BgGradientAngle, gradient.Angle, defaults.BgGradientAngle)
	gradient.FromStop = chooseFloat(args.BgGradientFromStop, gradient.FromStop, defaults.BgGradientFromStop)
	gradient.ToStop = chooseFloat(args.BgGradientToStop, gradient.ToStop, defaults.BgGradientToStop)
	gradient.Scope = "global"
	return gradient
}

func chooseFloat(override *float64, base *float64, fallback float64) *float64 {
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

func chooseScope(override string, base string, fallback string, defaultScope string) string {
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
	if scope == "global" {
		return "global"
	}
	return "module"
}

func fallbackLight(light *string, fallback string) string {
	if light != nil && strings.TrimSpace(*light) != "" {
		return *light
	}
	return fallback
}
