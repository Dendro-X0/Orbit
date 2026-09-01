package project

import (
	"fmt"
	"os"
	"path/filepath"
)

const OrbitDirName = ".orbit"

// FindRoot walks up from cwd looking for .git or go.mod or package.json.
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if markerPresent(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no project root found from %s", start)
		}
		dir = parent
	}
}

func markerPresent(dir string) bool {
	for _, name := range []string{".git", "go.mod", "package.json", "pnpm-workspace.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func OrbitDir(root string) string {
	return filepath.Join(root, OrbitDirName)
}

func RunsDir(root string) string {
	return filepath.Join(OrbitDir(root), "runs")
}

func StatePath(root string) string {
	return filepath.Join(OrbitDir(root), "state.json")
}

func EnsureOrbitDir(root string) error {
	return os.MkdirAll(RunsDir(root), 0o755)
}
