package cli

import (
	"context"
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/charmbracelet/huh"
)

// printProviderAuthStatus checks each provider CLI session (wrangler, vercel, etc.).
// If you are already logged in on the web, OAuth is usually one click — Orbit does not
// read browser cookies; it trusts the official CLI's whoami after login.
func printProviderAuthStatus(ctx context.Context, ids []string) {
	for _, id := range ids {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		who, err := p.WhoAmI(ctx)
		if err != nil {
			fmt.Printf("  %s %-12s error checking session\n", ui.dim.Render("•"), p.DisplayName())
			continue
		}
		if who.LoggedIn {
			account := who.Account
			if account == "" {
				account = "logged in"
			}
			fmt.Printf("  %s %-12s %s\n", okMark(), styledProvider(p.DisplayName()), ui.value.Render(account))
		} else {
			fmt.Printf("  %s %-12s %s\n", failMark(), styledProvider(p.DisplayName()), ui.dim.Render("not logged in (browser OAuth if needed)"))
		}
	}
}

// ensureProvidersLoggedIn runs OAuth only for providers without an active CLI session.
func ensureProvidersLoggedIn(ctx context.Context, ids []string, prompt bool) error {
	printProviderAuthStatus(ctx, ids)
	fmt.Println()

	pending, err := providersNeedingAuth(ctx, ids)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		printSuccess("All selected providers have an active CLI session.")
		return nil
	}

	if !prompt || !isInteractive() {
		return authRequiredError(pending)
	}

	for _, id := range pending {
		p, err := provider.Get(id)
		if err != nil {
			return err
		}
		fmt.Printf("%s has no CLI session yet.\n", styledProvider(p.DisplayName()))
		fmt.Println(ui.dim.Render("If you are already signed in on the web, OAuth is usually one click."))
		start := true
		if err := huh.NewConfirm().
			Title(fmt.Sprintf("Log in to %s now?", p.DisplayName())).
			Description("Opens the provider's official browser OAuth flow.").
			Value(&start).
			Run(); err != nil {
			return err
		}
		if !start {
			return authRequiredError(pending)
		}
		if err := runOAuthLoginWizard(ctx, id, p); err != nil {
			return err
		}
		printProviderAuthStatus(ctx, []string{id})
		fmt.Println()
	}
	return nil
}

func providersNeedingAuth(ctx context.Context, ids []string) ([]string, error) {
	var pending []string
	for _, id := range ids {
		p, err := provider.Get(id)
		if err != nil {
			return nil, err
		}
		who, err := p.WhoAmI(ctx)
		if err != nil {
			return nil, err
		}
		if !who.LoggedIn {
			pending = append(pending, id)
		}
	}
	return pending, nil
}

func authRequiredError(ids []string) error {
	if len(ids) == 1 {
		return fmt.Errorf("not logged in to %s — run: orbit login %s", ids[0], ids[0])
	}
	msg := "not logged in — run:"
	for _, id := range ids {
		msg += fmt.Sprintf("\n  orbit login %s", id)
	}
	return fmt.Errorf("%s", msg)
}

func configureNeedsAuth(res provider.ConfigureResult) bool {
	for _, h := range res.Hints {
		if h.Code == "auth.required" {
			return true
		}
	}
	return false
}

func offerLoginAndRetryConfigure(
	ctx context.Context,
	root string,
	st *state.Project,
	provID, targetID, environment string,
) error {
	p, err := provider.Get(provID)
	if err != nil {
		return err
	}
	if !isInteractive() {
		return fmt.Errorf("configure failed for %s", provID)
	}
	retry := true
	if err := huh.NewConfirm().
		Title(fmt.Sprintf("Log in to %s and retry configure?", p.DisplayName())).
		Value(&retry).
		Run(); err != nil {
		return err
	}
	if !retry {
		return fmt.Errorf("configure failed for %s", provID)
	}
	if err := runOAuthLoginWizard(ctx, provID, p); err != nil {
		return err
	}
	return configureOne(ctx, root, st, provID, targetID, environment, false)
}

func providerAuthCheck(ctx context.Context, p provider.Provider) (*provider.Check, bool) {
	who, err := p.WhoAmI(ctx)
	if err != nil {
		return &provider.Check{
			Name:    "authentication",
			OK:      false,
			Message: err.Error(),
			Fix:     fmt.Sprintf("orbit login %s", p.ID()),
		}, false
	}
	if who.LoggedIn {
		msg := "logged in"
		if who.Account != "" {
			msg = who.Account
		}
		return &provider.Check{Name: "authentication", OK: true, Message: msg}, true
	}
	return &provider.Check{
		Name:    "authentication",
		OK:      false,
		Message: "not logged in",
		Fix:     fmt.Sprintf("orbit login %s", p.ID()),
	}, false
}
