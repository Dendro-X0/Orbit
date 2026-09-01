package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show project configuration and last deploy run",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			st, _ := state.Load(statePath(root))
			stack := detectStack(cmd.Context(), root)

			fmt.Println("orbit status")
			fmt.Println()
			fmt.Printf("Project:     %s\n", root)

			if len(stack) > 0 {
				fmt.Printf("Stack:       %s\n", strings.Join(stack, ", "))
			} else {
				fmt.Println("Stack:       (none detected)")
			}

			env := st.Environment
			if env == "" {
				env = "production"
			}
			fmt.Printf("Environment: %s\n", env)

			if len(st.ConfiguredProviders()) > 0 {
				fmt.Println("\nConfigured:")
				for _, id := range stack {
					if !st.IsConfigured(id) {
						continue
					}
					target := st.TargetFor(id)
					if target != "" {
						fmt.Printf("  • %s (%s)\n", id, target)
					} else {
						fmt.Printf("  • %s\n", id)
					}
				}
			} else {
				fmt.Println("\nConfigured:  (none — run: orbit configure)")
			}

			runDir, err := run.LatestRunDir(root)
			if err != nil {
				fmt.Println("\nLast run:    (none)")
				printToolkitHints(root)
				return nil
			}

			relRun, _ := filepath.Rel(root, runDir)
			fmt.Printf("\nLast run:    %s\n", filepath.ToSlash(relRun))

			if summary, err := run.LoadSummary(runDir); err == nil && summary.OK {
				fmt.Printf("  Status:    succeeded (%s)\n", summary.Duration)
				if summary.URL != "" {
					fmt.Printf("  API URL:   %s\n", summary.URL)
				}
				fmt.Printf("  Logs:      %s\n", summary.RunDir)
			} else if failure, err := run.LoadFailure(runDir); err == nil {
				fmt.Printf("  Status:    failed at %s\n", failure.FailedStep)
				fmt.Printf("  Error:     %s\n", failure.Message)
				fmt.Printf("  Logs:      %s\n", failure.LogPaths.Combined)
				fmt.Printf("  Retry:     orbit retry\n")
			} else if manifest, err := run.LoadManifest(runDir); err == nil {
				if manifest.OK {
					fmt.Println("  Status:    succeeded")
				} else {
					fmt.Println("  Status:    incomplete")
				}
			}

			printToolkitHints(root)
			return nil
		},
	}
}
