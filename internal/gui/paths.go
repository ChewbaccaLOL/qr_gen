package gui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func FindVariantsPath(cwd string, exePath string) (string, error) {
	candidates := make([]string, 0, 6)
	seen := make(map[string]struct{})
	addCandidate := func(dir string) {
		if dir == "" {
			return
		}
		if _, ok := seen[dir]; ok {
			return
		}
		seen[dir] = struct{}{}
		candidates = append(candidates, dir)
	}

	addCandidate(cwd)
	addCandidate(filepath.Dir(cwd))
	if exePath != "" {
		exeDir := filepath.Dir(exePath)
		addCandidate(exeDir)
		parent := filepath.Dir(exeDir)
		addCandidate(parent)
		addCandidate(filepath.Dir(parent))
		addCandidate(filepath.Dir(filepath.Dir(parent)))
	}

	if len(candidates) == 0 {
		return "", errors.New("no search paths provided")
	}

	for _, dir := range candidates {
		candidate := filepath.Join(dir, "variants.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("unable to locate variants.json (checked %d locations)", len(candidates))
}

func DefaultExportDir(cwd string) string {
	if cwd == "" {
		return ""
	}
	return filepath.Join(cwd, "out")
}
