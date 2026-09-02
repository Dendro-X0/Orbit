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

var corsOriginsRe = regexp.MustCompile(`(?m)^\s*CORS_ORIGINS\s*=\s*"([^"]+)"`)

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
