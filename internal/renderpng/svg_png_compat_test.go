package renderpng

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"qr_generator/internal/config"
	"qr_generator/internal/render"
)

func TestSvgToPngTolerance(t *testing.T) {
	root := findRepoRoot(t)
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}
	if err := ensurePythonDep(pythonPath, "segno"); err != nil {
		t.Skipf("python dependency missing: %v", err)
	}
	if err := ensurePythonDep(pythonPath, "cairosvg"); err != nil {
		t.Skipf("cairosvg missing: %v", err)
	}

	cfg, err := config.Load(filepath.Join(root, "variants.json"))
	if err != nil {
		t.Fatalf("load variants: %v", err)
	}

	matrix, err := pythonMatrix(root, pythonPath, "https://example.com", "m")
	if err != nil {
		t.Fatalf("python matrix: %v", err)
	}

	cases := []string{"classic", "rounded", "sunset", "neon"}
	for _, name := range cases {
		variant := cfg.Variants[name]
		svg, err := render.RenderSVG(
			matrix,
			10,
			4,
			variant.Dark,
			variant.Light,
			variant.Shape,
			variant.Radius,
			variant.Gradient,
			variant.BackgroundGradient,
			nil,
			0,
			0,
			0,
			"after",
		)
		if err != nil {
			t.Fatalf("render svg: %v", err)
		}

		tmpDir := t.TempDir()
		svgPath := filepath.Join(tmpDir, name+".svg")
		pngPath := filepath.Join(tmpDir, name+".png")
		if err := os.WriteFile(svgPath, []byte(svg), 0o644); err != nil {
			t.Fatalf("write svg: %v", err)
		}

		if err := runCairoSVG(pythonPath, svgPath, pngPath); err != nil {
			t.Fatalf("cairosvg: %v", err)
		}

		file, err := os.Open(pngPath)
		if err != nil {
			t.Fatalf("open png: %v", err)
		}
		rasterized, err := png.Decode(file)
		file.Close()
		if err != nil {
			t.Fatalf("decode png: %v", err)
		}

		native, err := RenderPNG(
			matrix,
			10,
			4,
			variant.Dark,
			variant.Light,
			variant.Shape,
			variant.Radius,
			variant.Gradient,
			variant.BackgroundGradient,
		)
		if err != nil {
			t.Fatalf("render native png: %v", err)
		}

		if rasterized.Bounds().Dx() != native.Bounds().Dx() || rasterized.Bounds().Dy() != native.Bounds().Dy() {
			t.Fatalf("png size mismatch: svg=%v native=%v", rasterized.Bounds(), native.Bounds())
		}

		threshold := similarityThresholdForVariant(name)
		ratio := pixelSimilarityThresholded(rasterized, native, 8, 0x80)
		if ratio < threshold {
			t.Fatalf("png similarity too low for %s: %.4f (threshold %.4f)", name, ratio, threshold)
		}
	}
}

func runCairoSVG(pythonPath, svgPath, pngPath string) error {
	script := fmt.Sprintf("import cairosvg; cairosvg.svg2png(url=%s, write_to=%s)", reprPythonString(svgPath), reprPythonString(pngPath))
	cmd := exec.Command(pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func pixelSimilarity(a, b image.Image, maxDelta uint8) float64 {
	bounds := a.Bounds()
	total := (bounds.Dx()) * (bounds.Dy())
	if total == 0 {
		return 1
	}
	match := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, aa := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()
			if channelDiff(ar, br) <= maxDelta && channelDiff(ag, bg) <= maxDelta && channelDiff(ab, bb) <= maxDelta && channelDiff(aa, ba) <= maxDelta {
				match++
			}
		}
	}
	return float64(match) / float64(total)
}

func pixelSimilarityThresholded(a, b image.Image, maxDelta uint8, alphaThreshold uint8) float64 {
	bounds := a.Bounds()
	total := bounds.Dx() * bounds.Dy()
	if total == 0 {
		return 1
	}
	match := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, aa := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()
			aa8 := uint8(aa >> 8)
			ba8 := uint8(ba >> 8)
			if aa8 < alphaThreshold && ba8 < alphaThreshold {
				match++
				continue
			}
			if channelDiff(ar, br) <= maxDelta && channelDiff(ag, bg) <= maxDelta && channelDiff(ab, bb) <= maxDelta && channelDiff(aa, ba) <= maxDelta {
				match++
			}
		}
	}
	return float64(match) / float64(total)
}

func channelDiff(a, b uint32) uint8 {
	if a > b {
		return uint8((a - b) >> 8)
	}
	return uint8((b - a) >> 8)
}

func pythonMatrix(root, pythonPath, data, level string) ([][]bool, error) {
	script := strings.Join([]string{
		"import json, segno",
		"qr = segno.make(" + reprPythonString(data) + ", error=" + reprPythonString(level) + ")",
		"matrix = [list(row) for row in qr.matrix]",
		"print(json.dumps(matrix))",
	}, "; ")
	cmd := exec.Command(pythonPath, "-c", script)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v (%s)", err, strings.TrimSpace(string(output)))
	}
	var raw [][]int
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("decode python matrix: %v", err)
	}
	matrix := make([][]bool, len(raw))
	for y, row := range raw {
		matrix[y] = make([]bool, len(row))
		for x, cell := range row {
			matrix[y][x] = cell != 0
		}
	}
	return matrix, nil
}

func ensurePythonDep(pythonPath, module string) error {
	cmd := exec.Command(pythonPath, "-c", "import "+module)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s (%s)", err.Error(), strings.TrimSpace(string(output)))
	}
	return nil
}

func reprPythonString(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "\\'") + "'"
}

func similarityThreshold() float64 {
	value := strings.TrimSpace(os.Getenv("PNG_SIM_THRESHOLD"))
	if value == "" {
		return 0.94
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.94
	}
	if parsed <= 0 || parsed > 1 {
		return 0.94
	}
	return parsed
}

func similarityThresholdForVariant(variant string) float64 {
	defaultThreshold := similarityThreshold()
	switch variant {
	case "sunset", "neon":
		if defaultThreshold > 0.90 {
			return 0.90
		}
		return defaultThreshold
	case "rounded", "dot", "clear-rounded", "clear-dot":
		if defaultThreshold > 0.94 {
			return 0.94
		}
		return defaultThreshold
	default:
		return defaultThreshold
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := cwd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "variants.json")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate variants.json from %s", cwd)
	return ""
}
