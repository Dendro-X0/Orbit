package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/run"
)

const (
	signetConfig    = "signet.toml"
	legacySignetCfg = "selfsign.toml"
)

func hasSignetProject(root string) bool {
	return run.FileExists(filepath.Join(root, signetConfig)) ||
		run.FileExists(filepath.Join(root, legacySignetCfg))
}

func printToolkitHints(root string) {
	var hints []string
	if hasSignetProject(root) {
		hints = append(hints, "signet.toml detected — desktop signing: signet doctor, signet build")
	}
	if len(hints) == 0 {
		return
	}
	fmt.Println("\nToolkit:")
	for _, h := range hints {
		fmt.Printf("  • %s\n", h)
	}
	fmt.Println("  Docs: Orbit + Signet compose workflows (signet build → orbit deploy → signet release)")
}

func parseProviderIDs(label string) []string {
	if label == "" {
		return nil
	}
	parts := strings.Split(label, "+")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sliceStepsFrom(steps []run.Step, stepID string) ([]run.Step, error) {
	for i, s := range steps {
		if s.ID == stepID {
			return steps[i:], nil
		}
	}
	return nil, fmt.Errorf("step %q not found in deploy plan", stepID)
}

func openURL(url string) error {
	if url == "" {
		return fmt.Errorf("no URL available")
	}
	if err := openBrowser(url); err != nil {
		fmt.Printf("Open in browser: %s\n", url)
		return err
	}
	fmt.Printf("Opened %s\n", url)
	return nil
}
