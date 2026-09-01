package cli

import (
	"context"
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newConfigureCmd() *cobra.Command {
	var dryRun bool
	var env string

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
			if provID == "" {
				provID = pickProvider(root)
			}
			if provID == "" {
				return fmt.Errorf("no supported provider detected — pass --provider")
			}

			p, err := provider.Get(provID)
			if err != nil {
				return err
			}

			fmt.Printf("Configuring %s at %s\n\n", p.DisplayName(), root)
			res, err := p.Configure(cmd.Context(), root, provider.ConfigureOptions{
				Environment: env,
				DryRun:      dryRun,
			})
			if err != nil {
				return err
			}

			if res.Message != "" {
				if res.OK {
					fmt.Printf("✓ %s\n", res.Message)
				} else {
					fmt.Printf("✗ %s\n", res.Message)
				}
			}
			for _, h := range res.Hints {
				fmt.Printf("  hint [%s]: %s\n", h.Code, h.Message)
				if h.Action != "" {
					fmt.Printf("  action: %s\n", h.Action)
				}
			}

			if res.OK && !dryRun {
				st.Provider = provID
				st.Configured = true
				if env != "" {
					st.Environment = env
				}
				_ = state.Save(statePath(root), st)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned changes without applying")
	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	cmd.Flags().StringVar(&providerFlag, "provider", "", "provider id (cloudflare, …)")
	return cmd
}

func pickProvider(root string) string {
	if providerFlag != "" {
		return providerFlag
	}
	for _, p := range provider.All() {
		det, err := p.Detect(context.Background(), root)
		if err == nil && det.Supported {
			return p.ID()
		}
	}
	return ""
}
