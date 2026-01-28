package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"qr_generator/internal/config"
)

type Args struct {
	Data              string
	Output            string
	OutputSet         bool
	Png               bool
	PngOutput         string
	PngScale          float64
	Gif               bool
	Animation         bool
	AnimationFormat   string
	AnimationVariant  string
	GifOutput         string
	GifVariant        string
	GifFps            int
	GifFrames         int
	GifHold           int
	WaveAmp           float64
	WavePeriod        float64
	FloatAngle        *float64
	ReadableGif       bool
	Pdf               bool
	PdfOutput         string
	Ps                bool
	PsOutput          string
	Variant           string
	Scale             int
	Border            int
	ErrorLevel        string
	Dark              string
	Light             string
	NoBackground      bool
	Radius            *float64
	ListVariants      bool
	Catalog           bool
	CatalogColumns    int
	CatalogBackground string
	CatalogLabelSize  int
}

type ErrUsage struct {
	Err   error
	Usage string
}

func (e ErrUsage) Error() string {
	return e.Err.Error()
}

func ParseArgs(args []string, cfg *config.Config, env Env, stdin io.Reader, stdinIsTTY bool) (*Args, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	var usage bytes.Buffer
	fs := flag.NewFlagSet("qr_generator", flag.ContinueOnError)
	fs.SetOutput(&usage)
	fs.Usage = func() {
		fmt.Fprintln(&usage, "Go CLI prototype for the QR generator.")
		fmt.Fprintln(&usage, "Usage:")
		fmt.Fprintln(&usage, "  qr_generator [options] [data]")
		fmt.Fprintln(&usage)
		fs.PrintDefaults()
	}

	output := OptionalString{}
	pngOutput := OptionalString{}
	pdfOutput := OptionalString{}
	psOutput := OptionalString{}
	gifOutput := OptionalString{}
	gifVariant := OptionalString{}
	gifFps := OptionalInt{}
	gifFrames := OptionalInt{}
	gifHold := OptionalInt{}
	waveAmp := OptionalFloat{}
	wavePeriod := OptionalFloat{}
	floatAngle := OptionalFloat{}
	radius := OptionalFloat{}
	dark := OptionalString{}
	light := OptionalString{}

	scale := OptionalInt{}
	border := OptionalInt{}
	variant := OptionalString{}
	errorLevel := OptionalString{}
	animationFormat := OptionalString{}
	animationVariant := OptionalString{}

	png := OptionalBool{}
	gif := OptionalBool{}
	animation := OptionalBool{}
	readableGif := OptionalBool{}
	pdf := OptionalBool{}
	ps := OptionalBool{}
	noBackground := OptionalBool{}
	catalog := OptionalBool{}
	listVariants := OptionalBool{}

	catalogColumns := OptionalInt{}
	catalogBackground := OptionalString{}
	catalogLabelSize := OptionalInt{}

	fs.Var(&output, "output", "Output SVG file (default: out/qr.svg or QR_OUTPUT).")
	fs.Var(&output, "o", "Output SVG file (default: out/qr.svg or QR_OUTPUT).")
	fs.Var(&png, "png", "Also write a PNG file alongside the SVG (requires cairosvg).")
	fs.Var(&pngOutput, "png-output", "PNG output file (default: derived from --output).")
	fs.Var(&gif, "gif", "Alias for --animation --animation-format gif.")
	fs.Var(&animation, "animation", "Also write an animated output (default format: gif).")
	fs.Var(&animationFormat, "animation-format", "Animation format (default: gif).")
	fs.Var(&animationVariant, "animation-variant", "Animation variant (default: wave).")
	fs.Var(&gifOutput, "gif-output", "GIF output file (default: derived from --output).")
	fs.Var(&gifVariant, "gif-variant", "GIF animation variant (default: wave).")
	fs.Var(&gifFps, "gif-fps", "Frames per second for GIF animation.")
	fs.Var(&gifFrames, "gif-frames", "Number of animation frames (loop segment).")
	fs.Var(&gifHold, "gif-hold", "Number of still frames before/after the motion.")
	fs.Var(&waveAmp, "wave-amp", "Wave amplitude in module units (default: expressive).")
	fs.Var(&wavePeriod, "wave-period", "Wave period in columns (modules).")
	fs.Var(&floatAngle, "float-angle", "Float motion direction in degrees (90 = vertical).")
	fs.Var(&readableGif, "readable-gif", "Use scan-safer defaults for GIF wave settings.")
	fs.Var(&pdf, "pdf", "Also write a PDF file alongside the SVG (requires cairosvg).")
	fs.Var(&pdfOutput, "pdf-output", "PDF output file (default: derived from --output).")
	fs.Var(&ps, "ps", "Also write a PostScript file alongside the SVG (requires cairosvg).")
	fs.Var(&psOutput, "ps-output", "PostScript output file (default: derived from --output).")
	fs.Var(&variant, "variant", "Visual style variant.")
	fs.Var(&variant, "v", "Visual style variant.")
	fs.Var(&scale, "scale", "Module size in pixels (default: 10).")
	fs.Var(&border, "border", "Quiet zone in modules (default: 4).")
	fs.Var(&errorLevel, "error", "Error correction level (default: m).")
	fs.Var(&dark, "dark", "Override foreground color (hex or CSS color).")
	fs.Var(&light, "light", "Override background color (hex or CSS color).")
	fs.Var(&noBackground, "no-background", "Make the background transparent.")
	fs.Var(&radius, "radius", "Corner radius for rounded modules (0-0.5).")
	fs.Var(&listVariants, "list-variants", "List available variants and exit.")
	fs.Var(&catalog, "catalog", "Generate a catalog grid containing all variants.")
	fs.Var(&catalogColumns, "catalog-columns", "Number of columns in the catalog grid (default: 3).")
	fs.Var(&catalogBackground, "catalog-background", "Background color for the catalog canvas (default: #ffffff).")
	fs.Var(&catalogLabelSize, "catalog-label-size", "Label font size for catalog (0 = auto).")

	pngScale := OptionalFloat{}
	fs.Var(&pngScale, "png-scale", "Scale factor for PNG export (default: 3.0).")

	if err := fs.Parse(args); err != nil {
		fs.Usage()
		return nil, ErrUsage{Err: err, Usage: usage.String()}
	}

	if len(fs.Args()) > 1 {
		fs.Usage()
		return nil, ErrUsage{Err: errors.New("too many positional arguments"), Usage: usage.String()}
	}

	resolved := &Args{}
	if listVariants.IsSet {
		resolved.ListVariants = listVariants.Value
	} else {
		resolved.ListVariants = false
	}
	resolved.Catalog = resolveBool(catalog, env, "QR_CATALOG", false)
	resolved.CatalogColumns = resolveInt(catalogColumns, env, "QR_CATALOG_COLUMNS", 3)
	resolved.CatalogBackground = resolveString(catalogBackground, env, "QR_CATALOG_BACKGROUND", "#ffffff")
	resolved.CatalogLabelSize = resolveInt(catalogLabelSize, env, "QR_CATALOG_LABEL_SIZE", 0)

	resolved.Png = resolveBool(png, env, "QR_PNG", false)
	resolved.PngScale = resolveFloat(pngScale, env, "QR_PNG_SCALE", 3.0)
	resolved.Pdf = resolveBool(pdf, env, "QR_PDF", false)
	resolved.Ps = resolveBool(ps, env, "QR_PS", false)
	resolved.Gif = resolveBool(gif, env, "QR_GIF", false)
	resolved.Animation = resolveBool(animation, env, "QR_ANIMATION", false)
	resolved.ReadableGif = resolveBool(readableGif, env, "QR_READABLE_GIF", false)
	resolved.NoBackground = resolveBool(noBackground, env, "QR_NO_BACKGROUND", false)

	resolved.Variant = resolveString(variant, env, "QR_VARIANT", "classic")
	resolved.Scale = resolveInt(scale, env, "QR_SCALE", 10)
	resolved.Border = resolveInt(border, env, "QR_BORDER", 4)
	resolved.ErrorLevel = resolveString(errorLevel, env, "QR_ERROR", "m")

	resolved.OutputSet = output.IsSet
	resolved.Output = resolveString(output, env, "QR_OUTPUT", "")

	resolved.PngOutput = resolveString(pngOutput, env, "QR_PNG_OUTPUT", "")
	resolved.PdfOutput = resolveString(pdfOutput, env, "QR_PDF_OUTPUT", "")
	resolved.PsOutput = resolveString(psOutput, env, "QR_PS_OUTPUT", "")
	resolved.GifOutput = resolveString(gifOutput, env, "QR_GIF_OUTPUT", "")

	resolved.GifVariant = resolveString(gifVariant, env, "QR_GIF_VARIANT", "")
	resolved.AnimationFormat = resolveString(animationFormat, env, "QR_ANIMATION_FORMAT", "")
	resolved.AnimationVariant = resolveString(animationVariant, env, "QR_ANIMATION_VARIANT", "")

	gifFpsValue, gifFpsSet := resolveOptionalInt(gifFps, env, "QR_GIF_FPS")
	gifFramesValue, gifFramesSet := resolveOptionalInt(gifFrames, env, "QR_GIF_FRAMES")
	gifHoldValue, gifHoldSet := resolveOptionalInt(gifHold, env, "QR_GIF_HOLD")
	waveAmpValue, waveAmpSet := resolveOptionalFloat(waveAmp, env, "QR_WAVE_AMP")
	wavePeriodValue, wavePeriodSet := resolveOptionalFloat(wavePeriod, env, "QR_WAVE_PERIOD")
	floatAngleValue, floatAngleSet := resolveOptionalFloat(floatAngle, env, "QR_FLOAT_ANGLE")

	radiusValue, radiusSet := resolveOptionalFloat(radius, env, "QR_RADIUS")
	if radiusSet {
		resolved.Radius = &radiusValue
	}

	if dark.IsSet {
		resolved.Dark = dark.Value
	} else if value, ok := env.Lookup("QR_DARK"); ok {
		resolved.Dark = value
	}

	if light.IsSet {
		resolved.Light = light.Value
	} else if value, ok := env.Lookup("QR_LIGHT"); ok {
		resolved.Light = value
	}

	if floatAngleSet {
		resolved.FloatAngle = &floatAngleValue
	}

	if resolved.Output == "" {
		resolved.Output = "out/qr.svg"
	}

	if resolved.Catalog && !resolved.OutputSet {
		if _, ok := env.Lookup("QR_OUTPUT"); !ok {
			resolved.Output = "out/catalog.svg"
		}
	}

	animationEnabled := resolved.Animation || resolved.Gif

	if resolved.AnimationFormat == "" {
		resolved.AnimationFormat = "gif"
	}

	if resolved.AnimationVariant == "" {
		if resolved.GifVariant != "" {
			resolved.AnimationVariant = resolved.GifVariant
		} else {
			resolved.AnimationVariant = "wave"
		}
	}

	defaults := defaultGifValues(cfg, resolved.ReadableGif)
	if !gifFpsSet {
		gifFpsValue = defaults.FPS
	}
	if !gifFramesSet {
		gifFramesValue = defaults.Frames
	}
	if !gifHoldSet {
		gifHoldValue = defaults.Hold
	}
	if !waveAmpSet {
		waveAmpValue = defaults.WaveAmp
	}
	if !wavePeriodSet {
		wavePeriodValue = defaults.WavePeriod
	}

	resolved.GifFps = gifFpsValue
	resolved.GifFrames = gifFramesValue
	resolved.GifHold = gifHoldValue
	resolved.WaveAmp = waveAmpValue
	resolved.WavePeriod = wavePeriodValue

	if resolved.ListVariants {
		return resolved, nil
	}

	if len(fs.Args()) == 1 {
		resolved.Data = fs.Args()[0]
	}

	if !resolved.ListVariants {
		data, err := readData(resolved.Data, env, stdin, stdinIsTTY)
		if err != nil {
			return nil, err
		}
		resolved.Data = data
	}

	if err := validate(resolved, cfg, animationEnabled); err != nil {
		return nil, err
	}

	return resolved, nil
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

func resolveString(opt OptionalString, env Env, name, fallback string) string {
	if opt.IsSet {
		return opt.Value
	}
	if value, ok := env.Lookup(name); ok {
		return value
	}
	return fallback
}

func resolveInt(opt OptionalInt, env Env, name string, fallback int) int {
	if opt.IsSet {
		return opt.Value
	}
	if value, ok := envInt(env, name); ok {
		return value
	}
	return fallback
}

func resolveFloat(opt OptionalFloat, env Env, name string, fallback float64) float64 {
	if opt.IsSet {
		return opt.Value
	}
	if value, ok := envFloat(env, name); ok {
		return value
	}
	return fallback
}

func resolveBool(opt OptionalBool, env Env, name string, fallback bool) bool {
	if opt.IsSet {
		return opt.Value
	}
	if value, ok := envBool(env, name); ok {
		return value
	}
	return fallback
}

func resolveOptionalInt(opt OptionalInt, env Env, name string) (int, bool) {
	if opt.IsSet {
		return opt.Value, true
	}
	if value, ok := envInt(env, name); ok {
		return value, true
	}
	return 0, false
}

func resolveOptionalFloat(opt OptionalFloat, env Env, name string) (float64, bool) {
	if opt.IsSet {
		return opt.Value, true
	}
	if value, ok := envFloat(env, name); ok {
		return value, true
	}
	return 0, false
}

func envInt(env Env, name string) (int, bool) {
	value, ok := env.Lookup(name)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func envFloat(env Env, name string) (float64, bool) {
	value, ok := env.Lookup(name)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func envBool(env Env, name string) (bool, bool) {
	value, ok := env.Lookup(name)
	if !ok {
		return false, false
	}
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "1", "true", "yes", "y", "on", "t":
		return true, true
	case "0", "false", "no", "n", "off", "f":
		return false, true
	default:
		return false, false
	}
}

func readData(current string, env Env, stdin io.Reader, stdinIsTTY bool) (string, error) {
	if current != "" {
		return current, nil
	}
	if !stdinIsTTY {
		raw, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		value := strings.TrimSpace(string(raw))
		if value != "" {
			return value, nil
		}
	}
	if value, ok := env.Lookup("QR_DATA"); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
	}
	return "", errors.New("data is required (pass text or pipe via stdin)")
}

func validate(args *Args, cfg *config.Config, animationEnabled bool) error {
	if args == nil {
		return errors.New("args are required")
	}
	if args.ListVariants {
		return nil
	}
	if args.Catalog && animationEnabled {
		return errors.New("animation output is not supported with --catalog")
	}
	if _, ok := cfg.Variants[args.Variant]; !ok {
		return fmt.Errorf("unknown variant '%s'", args.Variant)
	}
	if args.ErrorLevel != "l" && args.ErrorLevel != "m" && args.ErrorLevel != "q" && args.ErrorLevel != "h" {
		return fmt.Errorf("invalid error correction level '%s'", args.ErrorLevel)
	}
	if args.AnimationFormat != "gif" {
		return fmt.Errorf("animation format '%s' is not supported yet", args.AnimationFormat)
	}
	if args.Gif && args.AnimationFormat != "gif" {
		return errors.New("--gif can only be used with GIF output")
	}
	if args.AnimationVariant != "" && !contains(cfg.AnimationVariants, args.AnimationVariant) {
		return fmt.Errorf("animation variant '%s' is not supported", args.AnimationVariant)
	}
	if args.GifVariant != "" && !contains(cfg.AnimationVariants, args.GifVariant) {
		return fmt.Errorf("gif variant '%s' is not supported", args.GifVariant)
	}
	if args.Png && args.PngScale <= 0 {
		return errors.New("--png-scale must be greater than 0")
	}
	if animationEnabled {
		if args.GifFps <= 0 {
			return errors.New("--gif-fps must be greater than 0")
		}
		if args.GifFrames <= 0 {
			return errors.New("--gif-frames must be greater than 0")
		}
		if args.GifHold < 0 {
			return errors.New("--gif-hold must be 0 or greater")
		}
		if args.WaveAmp < 0 {
			return errors.New("--wave-amp must be 0 or greater")
		}
		if args.WavePeriod <= 0 {
			return errors.New("--wave-period must be greater than 0")
		}
	}
	return nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
