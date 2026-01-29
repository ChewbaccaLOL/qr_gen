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
	Data                string
	Output              string
	OutputSet           bool
	Png                 bool
	PngOutput           string
	PngScale            float64
	Gif                 bool
	Animation           bool
	AnimationFormat     string
	AnimationVariant    string
	GifOutput           string
	GifVariant          string
	GifFps              int
	GifFrames           int
	GifHold             int
	WaveAmp             float64
	WavePeriod          float64
	FloatCycles         int
	FloatAngle          *float64
	ReadableGif         bool
	Pdf                 bool
	PdfOutput           string
	Ps                  bool
	PsOutput            string
	Variant             string
	Scale               int
	Border              int
	ErrorLevel          string
	Dark                string
	Light               string
	NoBackground        bool
	Cutout              bool
	Radius              *float64
	GradientEnabled     bool
	GradientDisabled    bool
	GradientFrom        string
	GradientTo          string
	GradientAngle       *float64
	GradientFromStop    *float64
	GradientToStop      *float64
	GradientScope       string
	GradientConfigSet   bool
	BgGradientEnabled   bool
	BgGradientDisabled  bool
	BgGradientFrom      string
	BgGradientTo        string
	BgGradientAngle     *float64
	BgGradientFromStop  *float64
	BgGradientToStop    *float64
	BgGradientConfigSet bool
	ListVariants        bool
	Catalog             bool
	CatalogColumns      int
	CatalogBackground   string
	CatalogLabelSize    int
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
	gradientFrom := OptionalString{}
	gradientTo := OptionalString{}
	gradientScope := OptionalString{}
	gradientAngle := OptionalFloat{}
	gradientFromStop := OptionalFloat{}
	gradientToStop := OptionalFloat{}
	bgGradientFrom := OptionalString{}
	bgGradientTo := OptionalString{}
	bgGradientAngle := OptionalFloat{}
	bgGradientFromStop := OptionalFloat{}
	bgGradientToStop := OptionalFloat{}

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
	cutout := OptionalBool{}
	gradientEnabled := OptionalBool{}
	gradientDisabled := OptionalBool{}
	bgGradientEnabled := OptionalBool{}
	bgGradientDisabled := OptionalBool{}
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
	fs.Var(&gradientEnabled, "gradient", "Enable gradient fill for modules.")
	fs.Var(&gradientDisabled, "no-gradient", "Disable gradient fill even if the variant includes one.")
	fs.Var(&gradientFrom, "gradient-from", "Gradient start color (hex or CSS color).")
	fs.Var(&gradientTo, "gradient-to", "Gradient end color (hex or CSS color).")
	fs.Var(&gradientAngle, "gradient-angle", "Gradient direction in degrees (0 = left to right).")
	fs.Var(&gradientFromStop, "gradient-from-stop", "Gradient start stop position (0-1).")
	fs.Var(&gradientToStop, "gradient-to-stop", "Gradient end stop position (0-1).")
	fs.Var(&gradientScope, "gradient-scope", "Gradient scope (module or global).")
	fs.Var(&bgGradientEnabled, "bg-gradient", "Enable background gradient.")
	fs.Var(&bgGradientDisabled, "no-bg-gradient", "Disable background gradient.")
	fs.Var(&bgGradientFrom, "bg-gradient-from", "Background gradient start color (hex or CSS color).")
	fs.Var(&bgGradientTo, "bg-gradient-to", "Background gradient end color (hex or CSS color).")
	fs.Var(&bgGradientAngle, "bg-gradient-angle", "Background gradient direction in degrees.")
	fs.Var(&bgGradientFromStop, "bg-gradient-from-stop", "Background gradient start stop position (0-1).")
	fs.Var(&bgGradientToStop, "bg-gradient-to-stop", "Background gradient end stop position (0-1).")
	fs.Var(&noBackground, "no-background", "Make the background transparent.")
	fs.Var(&cutout, "cutout", "Render modules as cutouts in the background.")
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
	resolved.GradientEnabled = resolveBool(gradientEnabled, env, "QR_GRADIENT", false)
	resolved.GradientDisabled = resolveBool(gradientDisabled, env, "QR_NO_GRADIENT", false)
	resolved.BgGradientEnabled = resolveBool(bgGradientEnabled, env, "QR_BG_GRADIENT", false)
	resolved.BgGradientDisabled = resolveBool(bgGradientDisabled, env, "QR_NO_BG_GRADIENT", false)

	resolved.Variant = resolveString(variant, env, "QR_VARIANT", "classic")
	resolved.Scale = resolveInt(scale, env, "QR_SCALE", 10)
	resolved.Border = resolveInt(border, env, "QR_BORDER", 4)
	resolved.ErrorLevel = resolveString(errorLevel, env, "QR_ERROR", "m")

	cutoutValue := false
	cutoutSet := false
	if cutout.IsSet {
		cutoutValue = cutout.Value
		cutoutSet = true
	} else if value, ok := envBool(env, "QR_CUTOUT"); ok {
		cutoutValue = value
		cutoutSet = true
	}
	if !cutoutSet {
		if variantCfg, ok := cfg.Variants[resolved.Variant]; ok {
			cutoutValue = variantCfg.Cutout
		}
	}
	resolved.Cutout = cutoutValue

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

	gradientFromValue, gradientFromSet := resolveOptionalString(gradientFrom, env, "QR_GRADIENT_FROM")
	gradientToValue, gradientToSet := resolveOptionalString(gradientTo, env, "QR_GRADIENT_TO")
	gradientScopeValue, gradientScopeSet := resolveOptionalString(gradientScope, env, "QR_GRADIENT_SCOPE")
	gradientAngleValue, gradientAngleSet := resolveOptionalFloat(gradientAngle, env, "QR_GRADIENT_ANGLE")
	gradientFromStopValue, gradientFromStopSet := resolveOptionalFloat(gradientFromStop, env, "QR_GRADIENT_FROM_STOP")
	gradientToStopValue, gradientToStopSet := resolveOptionalFloat(gradientToStop, env, "QR_GRADIENT_TO_STOP")

	if gradientFromSet {
		resolved.GradientFrom = gradientFromValue
	}
	if gradientToSet {
		resolved.GradientTo = gradientToValue
	}
	if gradientScopeSet {
		resolved.GradientScope = gradientScopeValue
	}
	if gradientAngleSet {
		resolved.GradientAngle = &gradientAngleValue
	}
	if gradientFromStopSet {
		resolved.GradientFromStop = &gradientFromStopValue
	}
	if gradientToStopSet {
		resolved.GradientToStop = &gradientToStopValue
	}
	resolved.GradientConfigSet = gradientFromSet || gradientToSet || gradientScopeSet || gradientAngleSet || gradientFromStopSet || gradientToStopSet

	bgGradientFromValue, bgGradientFromSet := resolveOptionalString(bgGradientFrom, env, "QR_BG_GRADIENT_FROM")
	bgGradientToValue, bgGradientToSet := resolveOptionalString(bgGradientTo, env, "QR_BG_GRADIENT_TO")
	bgGradientAngleValue, bgGradientAngleSet := resolveOptionalFloat(bgGradientAngle, env, "QR_BG_GRADIENT_ANGLE")
	bgGradientFromStopValue, bgGradientFromStopSet := resolveOptionalFloat(bgGradientFromStop, env, "QR_BG_GRADIENT_FROM_STOP")
	bgGradientToStopValue, bgGradientToStopSet := resolveOptionalFloat(bgGradientToStop, env, "QR_BG_GRADIENT_TO_STOP")

	if bgGradientFromSet {
		resolved.BgGradientFrom = bgGradientFromValue
	}
	if bgGradientToSet {
		resolved.BgGradientTo = bgGradientToValue
	}
	if bgGradientAngleSet {
		resolved.BgGradientAngle = &bgGradientAngleValue
	}
	if bgGradientFromStopSet {
		resolved.BgGradientFromStop = &bgGradientFromStopValue
	}
	if bgGradientToStopSet {
		resolved.BgGradientToStop = &bgGradientToStopValue
	}
	resolved.BgGradientConfigSet = bgGradientFromSet || bgGradientToSet || bgGradientAngleSet || bgGradientFromStopSet || bgGradientToStopSet

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

	floatCycles := cfg.Defaults.FloatCycles
	if floatCycles <= 0 {
		floatCycles = 1
	}
	resolved.FloatCycles = floatCycles
	_, gifHoldEnvSet := env.Lookup("QR_GIF_HOLD")
	if strings.HasPrefix(resolved.AnimationVariant, "float") && !gifHoldSet && !gifHoldEnvSet {
		if cfg.Defaults.FloatHold > 0 {
			resolved.GifHold = cfg.Defaults.FloatHold
		}
	}

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

func resolveOptionalString(opt OptionalString, env Env, name string) (string, bool) {
	if opt.IsSet {
		return opt.Value, true
	}
	if value, ok := env.Lookup(name); ok {
		return value, true
	}
	return "", false
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
	if args.Cutout && args.NoBackground {
		return errors.New("--cutout cannot be used with --no-background")
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
	if args.GradientEnabled && args.GradientDisabled {
		return errors.New("--gradient and --no-gradient cannot be used together")
	}
	if args.BgGradientEnabled && args.BgGradientDisabled {
		return errors.New("--bg-gradient and --no-bg-gradient cannot be used together")
	}
	if args.GradientFromStop != nil {
		if *args.GradientFromStop < 0 || *args.GradientFromStop > 1 {
			return errors.New("--gradient-from-stop must be between 0 and 1")
		}
	}
	if args.GradientToStop != nil {
		if *args.GradientToStop < 0 || *args.GradientToStop > 1 {
			return errors.New("--gradient-to-stop must be between 0 and 1")
		}
	}
	if args.GradientScope != "" && args.GradientScope != "module" && args.GradientScope != "global" {
		return errors.New("--gradient-scope must be 'module' or 'global'")
	}
	if args.BgGradientFromStop != nil {
		if *args.BgGradientFromStop < 0 || *args.BgGradientFromStop > 1 {
			return errors.New("--bg-gradient-from-stop must be between 0 and 1")
		}
	}
	if args.BgGradientToStop != nil {
		if *args.BgGradientToStop < 0 || *args.BgGradientToStop > 1 {
			return errors.New("--bg-gradient-to-stop must be between 0 and 1")
		}
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
