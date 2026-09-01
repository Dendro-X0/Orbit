package cli

import (
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy via one or more providers",
		Long:  "With no --provider flag, deploys all detected providers in stack order (e.g. Cloudflare API then Vercel docs).",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}
			if err := project.EnsureOrbitDir(root); err != nil {
				return err
			}

			st, _ := state.Load(statePath(root))
			ids, err := resolveDeployProviders(cmd.Context(), root, st, providerFlag)
			if err != nil {
				return err
			}

			deployEnv := env
			if deployEnv == "production" && st.Environment != "" {
				deployEnv = st.Environment
			}

			st, err = ensureConfigured(cmd.Context(), root, st, ids, deployEnv)
			if err != nil {
				return err
			}

			steps, err := buildDeploySteps(root, st, ids, deployEnv)
			if err != nil {
				return err
			}

			label := providerListLabel(ids)
			fmt.Printf("orbit deploy — %s (%s)\n\n", label, deployEnv)

			r := &run.Runner{}
			result, err := r.Execute(cmd.Context(), run.Options{
				Root:      root,
				Provider:  label,
				Command:   "deploy",
				PrintLive: true,
			}, steps)

			return printDeployResult(result, err)
		},
	}

	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	return cmd
}
