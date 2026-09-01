package cli

import (
	"fmt"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login [provider]",
		Short: "Log in to a cloud provider",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := ""
			if len(args) > 0 {
				id = args[0]
			}
			if id == "" {
				ids := provider.IDs()
				if len(ids) == 0 {
					return fmt.Errorf("no providers registered")
				}
				if len(ids) == 1 {
					id = ids[0]
				} else {
					return fmt.Errorf("choose a provider: %s", strings.Join(ids, ", "))
				}
			}

			p, err := provider.Get(id)
			if err != nil {
				return err
			}

			fmt.Printf("Logging in to %s…\n", p.DisplayName())
			res, err := p.Login(cmd.Context())
			if err != nil {
				return err
			}
			if !res.OK {
				return fmt.Errorf("login failed: %s", res.Message)
			}
			if res.Account != "" {
				fmt.Printf("✓ Logged in as %s\n", res.Account)
			} else {
				fmt.Printf("✓ %s\n", res.Message)
			}
			return nil
		},
	}
	return cmd
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show connected provider accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, p := range provider.All() {
				res, err := p.WhoAmI(cmd.Context())
				if err != nil {
					return err
				}
				status := "not logged in"
				if res.LoggedIn {
					status = "logged in"
					if res.Account != "" {
						status = res.Account
					}
				}
				fmt.Printf("%-12s %s\n", p.ID()+":", status)
			}
			return nil
		},
	}
}
