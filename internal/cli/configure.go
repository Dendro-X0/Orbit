package cli

import (
	"fmt"
	"os"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newConfigureCmd() *cobra.Command {
	var dryRun bool
	var env string
	var yes bool

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Interactive project setup for your provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}
			if err := project.EnsureOrbitDir(root); err != nil {
				return err
			}

			st, _ := state.Load(statePath(root))
			provID := st.Provider
			targetID := st.TargetID
			environment := env

			useWizard := isInteractive() && !yes && providerFlag == "" && !dryRun
			if useWizard {
				wizard, err := runConfigureWizard(cmd.Context(), root, provID, targetID, environment)
				if err != nil {
					return err
				}
				provID = wizard.ProviderID
				targetID = wizard.TargetID
				environment = wizard.Environment
			} else {
				if provID == "" {
					provID = pickProvider(root)
				}
				if provID == "" {
					return fmt.Errorf("no supported provider detected — pass --provider")
				}
			}

			p, err := provider.Get(provID)
			if err != nil {
				return err
			}

			if !useWizard {
				fmt.Printf("Configuring %s at %s\n\n", p.DisplayName(), root)
			}

			res, err := p.Configure(cmd.Context(), root, provider.ConfigureOptions{
				Environment: environment,
				TargetID:    targetID,
				DryRun:      dryRun,
			})
			if err != nil {
				return err
			}

			printConfigureResult(res)

			if res.OK && !dryRun {
				st.Provider = provID
				st.TargetID = targetID
				st.Configured = true
				if environment != "" {
					st.Environment = environment
				}
				if err := state.Save(statePath(root), st); err != nil {
					return err
				}
			} else if !res.OK {
				return fmt.Errorf("configure failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned changes without applying")
	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	cmd.Flags().StringVar(&providerFlag, "provider", "", "provider id (cloudflare, …)")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip interactive prompts")
	return cmd
}

func printConfigureResult(res provider.ConfigureResult) {
	if res.Message != "" {
		if res.OK {
			fmt.Printf("✓ %s\n", res.Message)
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s\n", res.Message)
		}
	}
	for _, change := range res.Changed {
		fmt.Printf("  updated %s\n", change)
	}
	for _, h := range res.Hints {
		fmt.Printf("  hint [%s]: %s\n", h.Code, h.Message)
		if h.Action != "" {
			fmt.Printf("  action: %s\n", h.Action)
		}
	}
}
