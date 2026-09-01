package fly

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var appNameRe = regexp.MustCompile(`(?m)^\s*app\s*=\s*['"]([^'"]+)['"]`)

type target struct {
	ID   string
	Path string
	Kind string
}

func findFlyTargets(root string) ([]target, error) {
	var targets []target

	if fileExists(filepath.Join(root, "fly.toml")) {
		targets = append(targets, target{ID: "root", Path: ".", Kind: detectKind(root, ".")})
	}

	appsDir := filepath.Join(root, "apps")
	entries, err := os.ReadDir(appsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			rel := filepath.Join("apps", entry.Name())
			if fileExists(filepath.Join(root, rel, "fly.toml")) {
				targets = append(targets, target{
					ID:   entry.Name(),
					Path: rel,
					Kind: detectKind(root, rel),
				})
			}
		}
	}

	return targets, nil
}

func detectKind(root, rel string) string {
	dir := filepath.Join(root, rel)
	if fileExists(filepath.Join(dir, "Dockerfile")) {
		return "container"
	}
	return "fly"
}

func parseAppName(flyTomlPath string) (string, error) {
	b, err := os.ReadFile(flyTomlPath)
	if err != nil {
		return "", err
	}
	m := appNameRe.FindStringSubmatch(string(b))
	if len(m) < 2 {
		return "", nil
	}
	return strings.TrimSpace(m[1]), nil
}

func flyTomlPath(root, targetPath string) string {
	return filepath.Join(root, targetPath, "fly.toml")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
