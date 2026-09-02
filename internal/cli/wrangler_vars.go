package cli

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/state"
)

var (
	corsOriginsRe = regexp.MustCompile(`(?m)^\s*CORS_ORIGINS\s*=\s*"([^"]+)"`)
	healthRouteRe = regexp.MustCompile(`\.(?:get|route|all)\s*\(\s*["']([^"']*health[^"']*)["']`)
)

const defaultAPIHealthPath = "/health"

func readWranglerCORSOrigins(wranglerPath string) []string {
	content, err := os.ReadFile(wranglerPath)
	if err != nil {
		return nil
	}
	match := corsOriginsRe.FindStringSubmatch(string(content))
	if len(match) < 2 {
		return nil
	}
	var origins []string
	for _, part := range strings.Split(match[1], ",") {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func wranglerPathForCloudflare(ctx context.Context, root string, st state.Project) string {
	p, err := provider.Get("cloudflare")
	if err != nil {
		return ""
	}
	det, err := p.Detect(ctx, root)
	if err != nil || !det.Supported {
		return ""
	}
	targetPath := targetPathFor(det, st.TargetFor("cloudflare"))
	if targetPath == "" {
		return ""
	}
	return filepath.Join(root, targetPath, "wrangler.toml")
}

func corsMissingOrigin(origins []string, url string) bool {
	if url == "" {
		return false
	}
	url = strings.TrimSuffix(strings.TrimSpace(url), "/")
	for _, origin := range origins {
		origin = strings.TrimSuffix(strings.TrimSpace(origin), "/")
		if origin == url {
			return false
		}
	}
	return true
}

func detectAPIHealthPath(root, targetPath string) string {
	if targetPath == "" {
		return defaultAPIHealthPath
	}
	srcDir := filepath.Join(root, targetPath, "src")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return defaultAPIHealthPath
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".js") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(srcDir, name))
		if err != nil {
			continue
		}
		if match := healthRouteRe.FindStringSubmatch(string(content)); len(match) > 1 {
			return match[1]
		}
	}
	return defaultAPIHealthPath
}

func apiHealthURL(ctx context.Context, root string, st state.Project, apiURL string) string {
	apiURL = strings.TrimSuffix(strings.TrimSpace(apiURL), "/")
	if apiURL == "" {
		return ""
	}
	p, err := provider.Get("cloudflare")
	if err != nil {
		return apiURL + defaultAPIHealthPath
	}
	det, err := p.Detect(ctx, root)
	if err != nil || !det.Supported {
		return apiURL + defaultAPIHealthPath
	}
	targetPath := targetPathFor(det, st.TargetFor("cloudflare"))
	return apiURL + detectAPIHealthPath(root, targetPath)
}

func corsOnlyLocalDev(origins []string) bool {
	if len(origins) == 0 {
		return false
	}
	for _, origin := range origins {
		lower := strings.ToLower(origin)
		if strings.HasPrefix(lower, "https://") && !strings.Contains(lower, "localhost") {
			return false
		}
	}
	return true
}
