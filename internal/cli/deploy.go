package cli

import (
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy via one or more providers",
		Long:  "With no --provider flag, deploys all detected providers in stack order (e.g. Cloudflare API then Vercel docs).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(cmd, env, providerFlag)
		},
	}

	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	cmd.Flags().StringVar(&providerFlag, "provider", "", "deploy a single provider")
	return cmd
}
