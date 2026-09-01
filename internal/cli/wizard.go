package cli

import (
	"context"
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/charmbracelet/huh"
)

type configureWizardInput struct {
	ProviderID  string
	TargetID    string
	Environment string
	Confirm     bool
}

func runConfigureWizard(ctx context.Context, root string, stProvider, stTarget, stEnv string) (configureWizardInput, error) {
	input := configureWizardInput{
		ProviderID:  stProvider,
		TargetID:    stTarget,
		Environment: stEnv,
	}
	if input.Environment == "" {
		input.Environment = "production"
	}

	providers := provider.All()
	if len(providers) == 0 {
		return input, fmt.Errorf("no providers registered")
	}

	providerOptions := make([]huh.Option[string], 0, len(providers))
	for _, p := range providers {
		label := p.DisplayName()
		det, err := p.Detect(ctx, root)
		if err == nil && det.Supported {
			label += " — " + det.Summary
		}
		providerOptions = append(providerOptions, huh.NewOption(label, p.ID()))
	}

	if input.ProviderID == "" {
		if len(providerOptions) == 1 {
			input.ProviderID = providers[0].ID()
		} else {
			if err := huh.NewSelect[string]().
				Title("Choose a cloud provider").
				Options(providerOptions...).
				Value(&input.ProviderID).
				Run(); err != nil {
				return input, err
			}
		}
	}

	p, err := provider.Get(input.ProviderID)
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
		Title("Log in to provider").
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
			huh.NewOption("Deploy this project", "deploy"),
			huh.NewOption("Configure project", "configure"),
			huh.NewOption("Log in to a provider", "login"),
			huh.NewOption("Check setup (doctor)", "doctor"),
			huh.NewOption("View last run logs", "logs"),
			huh.NewOption("Quit", "quit"),
		).
		Value(&action).
		Run(); err != nil {
		return "", err
	}
	return action, nil
}
