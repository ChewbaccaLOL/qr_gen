package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"qr_generator/internal/animation"
	"qr_generator/internal/cli"
	"qr_generator/internal/config"
	"qr_generator/internal/export"
	"qr_generator/internal/qr"
	"qr_generator/internal/render"
	"qr_generator/internal/renderpng"
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
		names := make([]string, 0, len(cfg.Variants))
		for name := range cfg.Variants {
			names = append(names, name)
		}
		sort.Strings(names)
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
		gradient = variant.Gradient
		if args.Dark != "" {
			gradient = nil
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

	if args.Pdf || args.Ps {
		pythonPath, err := exec.LookPath("python3")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: python3 is required for PDF/PS export")
			os.Exit(exitUsage)
		}
		if err := export.EnsureCairoSVG(pythonPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitUsage)
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
			cmd, err := export.CairoSVGCommand(pythonPath, export.FormatPDF, args.Output, pdfPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(exitUsage)
			}
			cmd.Dir = cwd
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n%s\n", err, string(output))
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
			cmd, err := export.CairoSVGCommand(pythonPath, export.FormatPS, args.Output, psPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(exitUsage)
			}
			cmd.Dir = cwd
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n%s\n", err, string(output))
				os.Exit(exitUsage)
			}
			fmt.Printf("Saved %s\n", psPath)
		}
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
		gradient := variant.Gradient
		if args.Dark != "" {
			gradient = nil
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
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
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
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
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
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
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
				args.WaveAmp,
				args.WavePeriod,
				*floatAngle,
				args.GifFrames,
				args.GifHold,
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
