package cli

import (
	"context"
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/charmbracelet/huh"
)

type configureWizardInput struct {
	ConfigureAll    bool
	ScopeProviders  []string
	ProviderID      string
	TargetID        string
	Environment     string
	Confirm         bool
}

func runConfigureWizard(
	ctx context.Context,
	root string,
	st state.Project,
	stack []string,
	activeScope []string,
	defaultEnv string,
) (configureWizardInput, error) {
	input := configureWizardInput{Environment: defaultEnv}
	if input.Environment == "" {
		input.Environment = "production"
	}

	targetStack := stack
	if len(activeScope) > 0 {
		targetStack = activeScope
		if label := st.ShipLabel(); label != "" && len(activeScope) < len(stack) {
			fmt.Printf("%s %s\n\n", ui.label.Render("Active scope:"), ui.title.Render(label))
		} else if len(activeScope) < len(stack) {
			fmt.Printf("%s %s\n\n", ui.label.Render("Active scope:"), ui.provider.Render(providerListLabel(activeScope)))
		}
	}

	if len(targetStack) == 1 {
		return runSingleProviderConfigureWizard(ctx, root, st, targetStack[0], input)
	}

	if len(targetStack) > 1 && len(targetStack) < len(stack) {
		input.ConfigureAll = true
		input.ScopeProviders = append([]string(nil), targetStack...)
		return runScopedStackConfigureWizard(input, targetStack)
	}

	if len(stack) > 1 {
		scope := "all"
		if err := huh.NewSelect[string]().
			Title("What do you want to configure?").
			Description("Prefer orbit ship to set a deploy scope — configure will follow it when set.").
			Options(
				huh.NewOption(fmt.Sprintf("All detected providers (%s)", providerListLabel(stack)), "all"),
				huh.NewOption("Single provider", "one"),
			).
			Value(&scope).
			Run(); err != nil {
			return input, err
		}
		if scope == "all" {
			input.ConfigureAll = true
			input.ScopeProviders = append([]string(nil), stack...)
			return runScopedStackConfigureWizard(input, stack)
		}
	}

	return runSingleProviderPickWizard(ctx, root, st, stack, input)
}

func runScopedStackConfigureWizard(input configureWizardInput, providers []string) (configureWizardInput, error) {
	if err := huh.NewSelect[string]().
		Title("Deployment environment").
		Options(
			huh.NewOption("Production", "production"),
			huh.NewOption("Preview / local", "preview"),
		).
		Value(&input.Environment).
		Run(); err != nil {
		return input, err
	}
	summary := fmt.Sprintf("Providers: %s\nEnvironment: %s", providerListLabel(providers), input.Environment)
	if err := huh.NewConfirm().
		Title("Apply configuration?").
		Description(summary).
		Value(&input.Confirm).
		Run(); err != nil {
		return input, err
	}
	if !input.Confirm {
		return input, fmt.Errorf("configure cancelled")
	}
	return input, nil
}

func runSingleProviderConfigureWizard(
	ctx context.Context,
	root string,
	st state.Project,
	providerID string,
	input configureWizardInput,
) (configureWizardInput, error) {
	input.ProviderID = providerID
	input.TargetID = st.TargetFor(providerID)

	p, err := provider.Get(providerID)
	if err != nil {
		return input, err
	}
	det, err := p.Detect(ctx, root)
	if err != nil {
		return input, err
	}
	if !det.Supported {
		return input, fmt.Errorf("%s does not detect a deployable project here", p.DisplayName())
	}

	fmt.Printf("\nDetected: %s\n", det.Summary)
	for _, t := range det.Targets {
		fmt.Printf("  • %s (%s) at %s\n", t.ID, t.Kind, t.Path)
	}

	if len(det.Targets) > 1 && input.TargetID == "" {
		targetOptions := make([]huh.Option[string], 0, len(det.Targets))
		for _, t := range det.Targets {
			targetOptions = append(targetOptions, huh.NewOption(fmt.Sprintf("%s — %s", t.ID, t.Path), t.ID))
		}
		if err := huh.NewSelect[string]().
			Title("Choose deploy target").
			Options(targetOptions...).
			Value(&input.TargetID).
			Run(); err != nil {
			return input, err
		}
	} else if input.TargetID == "" && len(det.Targets) > 0 {
		input.TargetID = det.Targets[0].ID
	}

	if err := huh.NewSelect[string]().
		Title("Deployment environment").
		Options(
			huh.NewOption("Production", "production"),
			huh.NewOption("Preview / local", "preview"),
		).
		Value(&input.Environment).
		Run(); err != nil {
		return input, err
	}

	summary := fmt.Sprintf(
		"Provider: %s\nTarget: %s\nEnvironment: %s",
		p.DisplayName(),
		input.TargetID,
		input.Environment,
	)
	if err := huh.NewConfirm().
		Title("Apply configuration?").
		Description(summary).
		Value(&input.Confirm).
		Run(); err != nil {
		return input, err
	}
	if !input.Confirm {
		return input, fmt.Errorf("configure cancelled")
	}
	return input, nil
}

func runSingleProviderPickWizard(
	ctx context.Context,
	root string,
	st state.Project,
	stack []string,
	input configureWizardInput,
) (configureWizardInput, error) {
	input.ProviderID = st.Provider
	if input.ProviderID == "" && len(stack) == 1 {
		input.ProviderID = stack[0]
	}
	input.TargetID = st.TargetFor(input.ProviderID)

	providerOptions := make([]huh.Option[string], 0, len(stack))
	for _, id := range stack {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		label := p.DisplayName()
		det, err := p.Detect(ctx, root)
		if err == nil && det.Supported {
			label += " — " + det.Summary
		}
		providerOptions = append(providerOptions, huh.NewOption(label, id))
	}

	if input.ProviderID == "" {
		if len(providerOptions) == 1 {
			input.ProviderID = stack[0]
		} else {
			if err := huh.NewSelect[string]().
				Title("Choose provider for this project").
				Description("Matches your ship scope when set — otherwise pick one provider.").
				Options(providerOptions...).
				Value(&input.ProviderID).
				Run(); err != nil {
				return input, err
			}
		}
	}

	return runSingleProviderConfigureWizard(ctx, root, st, input.ProviderID, input)
}

func runLoginWizard() (string, error) {
	ids := provider.IDs()
	if len(ids) == 0 {
		return "", fmt.Errorf("no providers registered")
	}
	if len(ids) == 1 {
		return ids[0], nil
	}

	var id string
	options := make([]huh.Option[string], 0, len(ids))
	for _, p := range provider.All() {
		options = append(options, huh.NewOption(p.DisplayName(), p.ID()))
	}
	if err := huh.NewSelect[string]().
		Title("Set up authentication for").
		Options(options...).
		Value(&id).
		Run(); err != nil {
		return "", err
	}
	return id, nil
}

func runMainMenu() (string, error) {
	var action string
	if err := huh.NewSelect[string]().
		Title("orbit — deploy to your cloud").
		Options(
			huh.NewOption("Ship to production (guided)", "ship"),
			huh.NewOption("Deploy this project", "deploy"),
			huh.NewOption("Configure project(s)", "configure"),
			huh.NewOption("Log in to a provider", "login"),
			huh.NewOption("Check setup (doctor)", "doctor"),
			huh.NewOption("Set worker secrets", "secrets"),
			huh.NewOption("Project status", "status"),
			huh.NewOption("View last run logs", "logs"),
			huh.NewOption("Quit", "quit"),
		).
		Value(&action).
		Run(); err != nil {
		return "", err
	}
	return action, nil
}

type shipMenuContext struct {
	PreviousDeploy *run.DeployRecord
	HasSecretsGap  bool
}

func runShipActionMenu(ctx shipMenuContext) (string, error) {
	var action string
	options := []huh.Option[string]{}

	if ctx.PreviousDeploy != nil {
		if ctx.PreviousDeploy.APIURL != "" {
			options = append(options, huh.NewOption("Open live API URL", "open-api"))
		}
		if ctx.PreviousDeploy.DocsURL != "" {
			options = append(options, huh.NewOption("Open live docs URL", "open-docs"))
		}
		options = append(options, huh.NewOption("View status", "status"))
		if ctx.HasSecretsGap {
			options = append(options, huh.NewOption("Set worker secrets", "secrets"))
		}
		options = append(options,
			huh.NewOption("Change scope (project type / providers)", "scope"),
			huh.NewOption("Log in", "login"),
			huh.NewOption("Configure", "configure"),
			huh.NewOption("Re-deploy (not recommended)", "full"),
			huh.NewOption("Quit", "quit"),
		)
	} else {
		options = append(options,
			huh.NewOption("Ship — prepare and deploy", "full"),
			huh.NewOption("Change scope (project type / providers)", "scope"),
			huh.NewOption("Log in", "login"),
			huh.NewOption("Configure", "configure"),
			huh.NewOption("Set worker secrets", "secrets"),
			huh.NewOption("Deploy", "deploy"),
			huh.NewOption("View status", "status"),
			huh.NewOption("Quit", "quit"),
		)
	}

	if err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(options...).
		Value(&action).
		Run(); err != nil {
		return "", err
	}
	return action, nil
}
