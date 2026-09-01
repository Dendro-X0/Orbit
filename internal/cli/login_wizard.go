package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/charmbracelet/huh"
)

func runOAuthLoginWizard(ctx context.Context, id string, p provider.Provider) error {
	guide, hasGuide := provider.AuthGuideFor(id)

	fmt.Printf("\n%s — browser login\n", p.DisplayName())
	fmt.Println()
	fmt.Println("Orbit opens the provider's official OAuth flow in your browser.")
	fmt.Println("Complete authorization there, then return to this terminal.")
	fmt.Println()

	if hasGuide && len(guide.OAuthSteps) > 0 {
		for i, step := range guide.OAuthSteps {
			fmt.Printf("  %d. %s\n", i+1, step)
		}
		fmt.Println()
	} else {
		fmt.Println("  1. Your browser will open")
		fmt.Println("  2. Sign in and authorize the provider CLI")
		fmt.Println("  3. Return here when finished")
		fmt.Println()
	}

	if hasGuide && guide.DocsURL != "" {
		fmt.Printf("Docs: %s\n", guide.DocsURL)
	}

	start := true
	if isInteractive() {
		if err := huh.NewConfirm().
			Title("Open browser login?").
			Description(fmt.Sprintf("Starts %s via the official CLI.", p.DisplayName())).
			Value(&start).
			Run(); err != nil {
			return err
		}
	}
	if !start {
		return fmt.Errorf("login cancelled")
	}

	fmt.Println()
	if err := runOAuthLogin(ctx, p); err != nil {
		return err
	}

	who, err := p.WhoAmI(ctx)
	if err != nil {
		return err
	}
	if !who.LoggedIn {
		return fmt.Errorf("login incomplete: %s", who.Message)
	}
	return nil
}

func runTokenLoginWizard(ctx context.Context, id string) error {
	guide, ok := provider.AuthGuideFor(id)
	if !ok {
		return fmt.Errorf("no token guide for provider %q", id)
	}

	p, err := provider.Get(id)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s — API token setup\n", p.DisplayName())
	fmt.Println()
	fmt.Println("Orbit stores your token in the OS keychain and passes it to the provider CLI.")
	fmt.Println("You create the token on the provider's site — Orbit opens the page for you.")
	fmt.Println()
	fmt.Printf("Permissions needed: %s\n\n", guide.Permissions)

	for i, step := range guide.Steps {
		fmt.Printf("  %d. %s\n", i+1, step)
	}
	fmt.Println()
	fmt.Printf("Create token: %s\n", guide.CreateURL)
	if guide.DocsURL != "" {
		fmt.Printf("Documentation: %s\n", guide.DocsURL)
	}
	fmt.Println()

	openPage := true
	if err := huh.NewConfirm().
		Title("Open the token page in your browser?").
		Description(guide.CreateURL).
		Value(&openPage).
		Run(); err != nil {
		return err
	}
	if openPage {
		if err := openBrowser(guide.CreateURL); err != nil {
			fmt.Fprintf(os.Stderr, "Could not open browser — visit manually:\n  %s\n\n", guide.CreateURL)
		} else {
			fmt.Println("Browser opened. Create your token, then return here.")
		}
	}

	var token string
	if err := huh.NewInput().
		Title(fmt.Sprintf("Paste your %s", guide.TokenLabel)).
		Description("Input is hidden. The token is stored in your OS keychain only.").
		EchoMode(huh.EchoModePassword).
		Value(&token).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("token cannot be empty")
			}
			return nil
		}).
		Run(); err != nil {
		return err
	}
	token = strings.TrimSpace(token)

	if err := credentials.Set(id, token); err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	who, err := p.WhoAmI(ctx)
	if err != nil {
		_ = credentials.Delete(id)
		return err
	}
	if !who.LoggedIn {
		_ = credentials.Delete(id)
		return fmt.Errorf("token rejected: %s", who.Message)
	}

	fmt.Printf("✓ %s token saved to keychain", p.DisplayName())
	if who.Account != "" {
		fmt.Printf(" (%s)", who.Account)
	}
	fmt.Println()
	return nil
}
