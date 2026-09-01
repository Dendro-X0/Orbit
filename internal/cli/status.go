package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
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

			printAuthStatus(cmd.Context(), stack)
			printPendingSetup(stack, st)

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
			} else if len(stack) > 0 {
				fmt.Println("\nConfigured:  (none)")
			}

			runDir, err := run.LatestRunDir(root)
			if err != nil {
				fmt.Println("\nLast run:    (none)")
				printNextSteps(cmd.Context(), root, stack, st, nil)
				printToolkitHints(root)
				return nil
			}

			relRun, _ := filepath.Rel(root, runDir)
			fmt.Printf("\nLast run:    %s\n", filepath.ToSlash(relRun))

			var summary *run.Summary
			if summary, err = run.LoadSummary(runDir); err == nil && summary.OK {
				fmt.Printf("  Status:    succeeded (%s)\n", summary.Duration)
				if summary.APIURL != "" {
					fmt.Printf("  API URL:   %s\n", summary.APIURL)
				}
				if summary.DocsURL != "" {
					fmt.Printf("  Docs URL:  %s\n", summary.DocsURL)
				}
				fmt.Printf("  Logs:      %s\n", summary.RunDir)
			} else if failure, err := run.LoadFailure(runDir); err == nil {
				fmt.Printf("  Status:    failed at %s\n", failure.FailedStep)
				fmt.Printf("  Error:     %s\n", failure.Message)
				if failure.Hint != nil && failure.Hint.Action != "" {
					fmt.Printf("  Action:    %s\n", failure.Hint.Action)
				}
				fmt.Printf("  Logs:      %s\n", failure.LogPaths.Combined)
				fmt.Printf("  Retry:     orbit retry\n")
			} else if manifest, err := run.LoadManifest(runDir); err == nil {
				if manifest.OK {
					fmt.Println("  Status:    succeeded")
				} else {
					fmt.Println("  Status:    incomplete")
				}
			}

			printNextSteps(cmd.Context(), root, stack, st, summary)
			if msg := cloudflareSecretsSummary(cmd.Context(), root, st); msg != "" {
				fmt.Printf("\nSecrets:     %s\n", msg)
			}
			printToolkitHints(root)
			return nil
		},
	}
}

func printAuthStatus(ctx context.Context, stack []string) {
	if len(stack) == 0 {
		return
	}
	fmt.Println("\nAuthentication:")
	for _, id := range stack {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		who, err := p.WhoAmI(ctx)
		if err != nil {
			fmt.Printf("  • %-12s error: %v\n", id, err)
			continue
		}
		if who.LoggedIn {
			line := "logged in"
			if who.Account != "" {
				line = who.Account
			}
			fmt.Printf("  • %-12s %s\n", id, line)
		} else {
			fmt.Printf("  • %-12s not logged in → orbit login %s\n", id, id)
		}
	}
}

func printPendingSetup(stack []string, st state.Project) {
	var pending []string
	for _, id := range stack {
		if !st.IsConfigured(id) {
			pending = append(pending, id)
		}
	}
	if len(pending) == 0 {
		return
	}
	fmt.Println("\nPending setup:")
	for _, id := range pending {
		fmt.Printf("  • %s → orbit configure --provider %s\n", id, id)
	}
}

func printNextSteps(ctx context.Context, root string, stack []string, st state.Project, summary *run.Summary) {
	if len(stack) == 0 {
		return
	}
	var steps []string
	for _, id := range stack {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		who, _ := p.WhoAmI(ctx)
		if !who.LoggedIn {
			steps = append(steps, fmt.Sprintf("orbit login %s", id))
			break
		}
	}
	for _, id := range stack {
		if !st.IsConfigured(id) {
			steps = append(steps, "orbit configure --all")
			break
		}
	}
	if summary == nil && len(st.ConfiguredProviders()) > 0 {
		steps = append(steps, "orbit deploy")
	}
	if summary != nil && summary.OK {
		if summary.APIURL != "" {
			steps = append(steps, "orbit open --target api")
		}
		if summary.DocsURL != "" {
			steps = append(steps, "orbit open --target docs")
		}
		if msg := cloudflareSecretsSummary(ctx, root, st); msg != "" {
			steps = append(steps, "orbit secrets")
		}
	}
	if len(steps) == 0 {
		return
	}
	fmt.Println("\nNext:")
	for _, s := range steps {
		fmt.Printf("  → %s\n", s)
	}
}
