package netlify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type siteLink struct {
	SiteID string `json:"siteId"`
}

type target struct {
	ID   string
	Path string
	Kind string
}

func findNetlifyTargets(root string) ([]target, error) {
	var targets []target

	if fileExists(filepath.Join(root, "netlify.toml")) {
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
			if fileExists(filepath.Join(root, rel, "netlify.toml")) {
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
	switch {
	case fileExists(filepath.Join(dir, "next.config.js")) ||
		fileExists(filepath.Join(dir, "next.config.ts")) ||
		fileExists(filepath.Join(dir, "next.config.mjs")):
		return "nextjs"
	case fileExists(filepath.Join(dir, "vite.config.ts")) ||
		fileExists(filepath.Join(dir, "vite.config.js")):
		return "vite"
	default:
		return "static"
	}
}

func statePath(root, targetPath string) string {
	return filepath.Join(root, targetPath, ".netlify", "state.json")
}

func readSiteLink(root, targetPath string) (siteLink, bool, error) {
	path := statePath(root, targetPath)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return siteLink{}, false, nil
		}
		return siteLink{}, false, err
	}
	var link siteLink
	if err := json.Unmarshal(b, &link); err != nil {
		return siteLink{}, false, err
	}
	return link, link.SiteID != "", nil
}

func requiredViteEnvVars(root, targetPath string) []string {
	examplePath := filepath.Join(root, targetPath, ".env.example")
	b, err := os.ReadFile(examplePath)
	if err != nil {
		return nil
	}
	var vars []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if strings.HasPrefix(key, "VITE_") {
			vars = append(vars, key)
		}
	}
	return vars
}

func parseEnvList(output string) map[string]bool {
	found := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		if strings.HasPrefix(name, "VITE_") {
			found[name] = true
		}
	}
	return found
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
