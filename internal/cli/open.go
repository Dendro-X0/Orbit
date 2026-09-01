package cli

import (
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open",
		Short: "Open the last deploy URL in your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			runDir, err := run.LatestRunDir(root)
			if err != nil {
				return fmt.Errorf("no deploy runs yet — run: orbit deploy")
			}

			if summary, err := run.LoadSummary(runDir); err == nil && summary.URL != "" {
				return openURL(summary.URL)
			}

			return fmt.Errorf("no URL in last run — deploy may have failed or URL was not captured")
		},
	}
}
