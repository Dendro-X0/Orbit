package cli

import (
	"path/filepath"

	"github.com/Dendro-X0/Orbit/internal/project"
)

func findRoot(start string) (string, error) {
	return project.FindRoot(start)
}

func orbitDir(root string) string {
	return project.OrbitDir(root)
}

func statePath(root string) string {
	return project.StatePath(root)
}

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}
