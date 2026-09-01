package vercel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type projectLink struct {
	ProjectID string `json:"projectId"`
	OrgID     string `json:"orgId"`
}

func projectLinkPath(root, targetPath string) string {
	return filepath.Join(root, targetPath, ".vercel", "project.json")
}

func readProjectLink(root, targetPath string) (projectLink, bool, error) {
	path := projectLinkPath(root, targetPath)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return projectLink{}, false, nil
		}
		return projectLink{}, false, err
	}
	var link projectLink
	if err := json.Unmarshal(b, &link); err != nil {
		return projectLink{}, false, err
	}
	return link, link.ProjectID != "", nil
}

func findVercelTargets(root string) ([]target, error) {
	var targets []target

	if fileExists(filepath.Join(root, "vercel.json")) {
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
			if fileExists(filepath.Join(root, rel, "vercel.json")) {
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

type target struct {
	ID   string
	Path string
	Kind string
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
		if strings.HasPrefix(name, "VITE_") || strings.HasPrefix(name, "NEXT_PUBLIC_") {
			found[name] = true
		}
	}
	return found
}
