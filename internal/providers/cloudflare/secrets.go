package cloudflare

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/run"
)

var secretTokenRe = regexp.MustCompile(`\b[A-Z][A-Z0-9_]{2,}\b`)

// ParseSecretNames reads recommended secret names from wrangler.toml comment blocks.
func ParseSecretNames(wranglerPath string) ([]string, error) {
	content, err := os.ReadFile(wranglerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	seen := map[string]bool{}
	inBlock := false

	for _, line := range strings.Split(string(content), "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "#") {
			if inBlock {
				break
			}
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(trim, "#"))
		lower := strings.ToLower(body)
		if isSecretsCommentHeader(lower) {
			inBlock = true
			if idx := strings.Index(body, ":"); idx >= 0 {
				addSecretTokens(body[idx+1:], seen, &names)
			}
			continue
		}
		if inBlock {
			if body == "" {
				break
			}
			addSecretTokens(body, seen, &names)
		}
	}
	return names, nil
}

func addSecretTokens(text string, seen map[string]bool, names *[]string) {
	for _, tok := range secretTokenRe.FindAllString(text, -1) {
		if seen[tok] {
			continue
		}
		seen[tok] = true
		*names = append(*names, tok)
	}
}

func isSecretsCommentHeader(lower string) bool {
	return strings.HasPrefix(lower, "secrets") ||
		strings.Contains(lower, "secrets (") ||
		strings.Contains(lower, "secrets:")
}

// ListRemoteSecrets returns secret names configured on the remote worker.
func ListRemoteSecrets(ctx context.Context, workDir string, extraEnv []string) ([]string, error) {
	out, err := run.Capture(ctx, "wrangler", []string{"secret", "list"}, workDir, extraEnv...)
	if err != nil {
		return nil, err
	}
	var names []string
	seen := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// JSON array or plain name per line
		for _, tok := range secretTokenRe.FindAllString(line, -1) {
			if seen[tok] {
				continue
			}
			seen[tok] = true
			names = append(names, tok)
		}
	}
	return names, nil
}

// SecretStatus compares documented secrets against remote worker configuration.
func SecretStatus(ctx context.Context, root, targetPath string, extraEnv []string) (required, set, missing []string, err error) {
	path := wranglerPath(root, targetPath)
	required, err = ParseSecretNames(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(required) == 0 {
		return nil, nil, nil, nil
	}
	workDir := filepath.Join(root, targetPath)
	remote, err := ListRemoteSecrets(ctx, workDir, extraEnv)
	if err != nil {
		return required, nil, required, nil
	}
	remoteSet := map[string]bool{}
	for _, name := range remote {
		remoteSet[name] = true
	}
	for _, name := range required {
		if remoteSet[name] {
			set = append(set, name)
		} else {
			missing = append(missing, name)
		}
	}
	return required, set, missing, nil
}

func PutSecret(ctx context.Context, workDir, name string, extraEnv []string) error {
	return run.RunInteractive(ctx, "wrangler", []string{"secret", "put", name}, workDir, extraEnv...)
}
