package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var tokenFlag string
	var guideFlag bool

	cmd := &cobra.Command{
		Use:   "login [provider]",
		Short: "Authenticate via the provider's official flow",
		Long: `Orbit is a portal — it hands off to the provider CLI for auth and deploy.

By default, orbit login opens the provider's browser OAuth (wrangler login, vercel login).
Use --guide for manual API token setup, or --token for scripting.`,
		Args: cobra.MaximumNArgs(1),
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
				return storeAndVerifyToken(cmd.Context(), id, p, token)
			}

			if guideFlag {
				if !isInteractive() {
					return fmt.Errorf("token guide requires an interactive terminal")
				}
				return runTokenLoginWizard(cmd.Context(), id)
			}

			if isInteractive() {
				return runOAuthLogin(cmd.Context(), p)
			}

			if _, ok := provider.AuthGuideFor(id); ok {
				guide, _ := provider.AuthGuideFor(id)
				return fmt.Errorf("not logged in — run interactively: orbit login %s\n  or visit: %s", id, guide.CreateURL)
			}
			return fmt.Errorf("not logged in — run: orbit login %s --token <token>", id)
		},
	}

	cmd.Flags().StringVar(&tokenFlag, "token", "", "API token (use - to read from stdin)")
	cmd.Flags().BoolVar(&guideFlag, "guide", false, "manual API token wizard (opens token page, paste to keychain)")
	return cmd
}

func storeAndVerifyToken(ctx context.Context, id string, p provider.Provider, token string) error {
	if err := credentials.Set(id, token); err != nil {
		return fmt.Errorf("store token: %w", err)
	}
	who, err := p.WhoAmI(ctx)
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

func runOAuthLogin(ctx context.Context, p provider.Provider) error {
	fmt.Printf("Logging in to %s…\n", p.DisplayName())
	res, err := p.Login(ctx)
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
