package cli

import (
	"fmt"
	"os"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy via your provider",
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
				return fmt.Errorf("no provider — run: orbit configure")
			}

			p, err := provider.Get(provID)
			if err != nil {
				return err
			}

			pp, ok := p.(provider.PhaseProvider)
			if !ok {
				return fmt.Errorf("provider %s does not support phased deploy yet", provID)
			}

			if !st.Configured {
				fmt.Println("Project not configured — running configure first…")
				configure := newConfigureCmd()
				configure.SetContext(cmd.Context())
				if err := configure.RunE(cmd, nil); err != nil {
					return err
				}
			}

			steps := pp.Phases(root, provider.DeployOptions{Environment: env, TargetID: st.TargetID})
			fmt.Printf("orbit deploy — %s (%s)\n\n", p.DisplayName(), env)

			r := &run.Runner{}
			result, err := r.Execute(cmd.Context(), run.Options{
				Root:      root,
				Provider:  provID,
				Command:   "deploy",
				PrintLive: true,
			}, steps)

			if result != nil && result.Summary != nil {
				fmt.Printf("\n✓ Deployed in %s\n", result.Summary.Duration)
				fmt.Printf("  Logs: %s\n", result.Summary.RunDir)
			}
			if result != nil && result.Failure != nil {
				fmt.Fprintf(os.Stderr, "\n✗ Deploy failed at step %s\n", result.Failure.FailedStep)
				fmt.Fprintf(os.Stderr, "  %s\n", result.Failure.Message)
				if result.Failure.Hint != nil && result.Failure.Hint.Action != "" {
					fmt.Fprintf(os.Stderr, "  action: %s\n", result.Failure.Hint.Action)
				}
				fmt.Fprintf(os.Stderr, "  logs: %s\n", result.Failure.LogPaths.Combined)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	return cmd
}
