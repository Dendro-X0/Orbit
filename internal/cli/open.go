package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open a deploy URL from the last run in your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			url, err := resolveOpenURL(root, target)
			if err != nil {
				return err
			}
			return openURL(url)
		},
	}

	cmd.Flags().StringVar(&target, "target", "api", "URL to open: api, docs, or any")
	return cmd
}

func resolveOpenURL(root, target string) (string, error) {
	runDir, err := run.LatestRunDir(root)
	if err != nil {
		return "", fmt.Errorf("no deploy runs yet — run: orbit deploy")
	}

	summary, _ := run.LoadSummary(runDir)
	logURL := deployURLsFromLog(runDir)

	api := ""
	docs := ""
	if summary != nil {
		if summary.APIURL != "" {
			api = summary.APIURL
		} else if summary.URL != "" && summary.DocsURL == "" {
			api = summary.URL
		}
		if summary.DocsURL != "" {
			docs = summary.DocsURL
		}
	}
	if api == "" {
		api = logURL.API
	}
	if docs == "" {
		docs = logURL.Docs
	}

	switch target {
	case "api":
		if api != "" {
			return api, nil
		}
	case "docs":
		if docs != "" {
			return docs, nil
		}
	case "any":
		if api != "" {
			return api, nil
		}
		if docs != "" {
			return docs, nil
		}
	default:
		return "", fmt.Errorf("unknown target %q — use api, docs, or any", target)
	}

	return "", fmt.Errorf("no %s URL in last run — try: orbit logs", target)
}

func deployURLsFromLog(runDir string) run.DeployURLs {
	b, err := os.ReadFile(filepath.Join(runDir, "combined.log"))
	if err != nil {
		return run.DeployURLs{}
	}
	return run.ExtractDeployURLs(string(b))
}
