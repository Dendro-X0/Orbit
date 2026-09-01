package cli

import (
	"fmt"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var tokenFlag string

	cmd := &cobra.Command{
		Use:   "login [provider]",
		Short: "Log in to a cloud provider",
		Long:  "Interactive OAuth via the provider CLI, or pass --token to store an API token in the OS keychain.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := ""
			if len(args) > 0 {
				id = args[0]
			}
			if id == "" && isInteractive() && tokenFlag == "" {
				var err error
				id, err = runLoginWizard()
				if err != nil {
					return err
				}
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

			token, err := readLoginToken(tokenFlag)
			if err != nil {
				return err
			}
			if token != "" {
				if err := credentials.Set(id, token); err != nil {
					return fmt.Errorf("store token: %w", err)
				}
				who, err := p.WhoAmI(cmd.Context())
				if err != nil {
					return err
				}
				if !who.LoggedIn {
					_ = credentials.Delete(id)
					return fmt.Errorf("token rejected: %s", who.Message)
				}
				fmt.Printf("✓ Stored %s token in keychain", p.DisplayName())
				if who.Account != "" {
					fmt.Printf(" (%s)", who.Account)
				}
				fmt.Println()
				return nil
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

	cmd.Flags().StringVar(&tokenFlag, "token", "", "API token (use - to read from stdin)")
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
					if credentials.Has(p.ID()) {
						status += " (keychain)"
					}
				}
				fmt.Printf("%-12s %s\n", p.ID()+":", status)
			}
			return nil
		},
	}
}
