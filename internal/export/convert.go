package export

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type CairoFormat string

const (
	FormatPDF CairoFormat = "pdf"
	FormatPS  CairoFormat = "ps"
	FormatPNG CairoFormat = "png"
)

func EnsureCairoSVG(pythonPath string) error {
	if pythonPath == "" {
		return errors.New("python path is required")
	}
	cmd := exec.Command(pythonPath, "-c", "import cairosvg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("cairosvg not available: %w", err)
		}
		return fmt.Errorf("cairosvg not available: %s", message)
	}
	return nil
}

func CairoSVGScript(format CairoFormat, svgPath string, outputPath string) (string, error) {
	if svgPath == "" || outputPath == "" {
		return "", errors.New("svg and output paths are required")
	}
	switch format {
	case FormatPDF, FormatPS, FormatPNG:
		return fmt.Sprintf(
			"import cairosvg; cairosvg.svg2%[1]s(url=%[2]q, write_to=%[3]q)",
			string(format),
			svgPath,
			outputPath,
		), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func CairoSVGCommand(pythonPath string, format CairoFormat, svgPath string, outputPath string) (*exec.Cmd, error) {
	if pythonPath == "" {
		return nil, errors.New("python path is required")
	}
	script, err := CairoSVGScript(format, svgPath, outputPath)
	if err != nil {
		return nil, err
	}
	return exec.Command(pythonPath, "-c", script), nil
}
