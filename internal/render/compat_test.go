package render

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"qr_generator/internal/config"
)

func TestGoSvgMatchesPython(t *testing.T) {
	root := findRepoRoot(t)
	cfg, err := config.Load(filepath.Join(root, "variants.json"))
	if err != nil {
		t.Fatalf("load variants: %v", err)
	}

	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}
	if err := ensurePythonDep(pythonPath, "segno"); err != nil {
		t.Skipf("python dependency missing: %v", err)
	}

	matrix, err := pythonMatrix(root, pythonPath, "https://example.com", "m")
	if err != nil {
		t.Fatalf("python matrix: %v", err)
	}

	cases := []string{"classic", "sunset"}
	for _, name := range cases {
		variant, ok := cfg.Variants[name]
		if !ok {
			t.Fatalf("missing variant %s", name)
		}

		goSVG, err := RenderSVG(
			matrix,
			10,
			4,
			variant.Dark,
			variant.Light,
			variant.Shape,
			variant.Radius,
			variant.Gradient,
			nil,
			0,
			0,
			0,
			"after",
		)
		if err != nil {
			t.Fatalf("render svg %s: %v", name, err)
		}

		outPath := filepath.Join(t.TempDir(), name+".svg")
		cmd := exec.Command(
			pythonPath,
			"python/qr_generator.py",
			"https://example.com",
			"--variant",
			name,
			"--scale",
			"10",
			"--border",
			"4",
			"--error",
			"m",
			"-o",
			outPath,
		)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "PYTHONPATH=python")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("python render %s failed: %v\n%s", name, err, string(output))
		}

		pySVG, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("read python svg %s: %v", name, err)
		}

		if string(pySVG) != goSVG {
			t.Fatalf("go/python svg mismatch for %s:\n%s", name, firstDiff(string(pySVG), goSVG))
		}
	}
}

func firstDiff(expected, actual string) string {
	limit := len(expected)
	if len(actual) < limit {
		limit = len(actual)
	}
	idx := 0
	for idx < limit {
		if expected[idx] != actual[idx] {
			break
		}
		idx++
	}
	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + 40
	if end > len(expected) {
		end = len(expected)
	}
	if end > len(actual) {
		end = len(actual)
	}
	snippetExpected := strings.ReplaceAll(expected[start:end], "\n", "\\n")
	snippetActual := strings.ReplaceAll(actual[start:end], "\n", "\\n")
	return "expected: " + snippetExpected + "\nactual:   " + snippetActual
}

func ensurePythonDep(pythonPath, module string) error {
	cmd := exec.Command(pythonPath, "-c", "import "+module)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s (%s)", err.Error(), strings.TrimSpace(string(output)))
	}
	return nil
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

func reprPythonString(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "\\'") + "'"
}
