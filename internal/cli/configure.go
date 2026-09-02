package cli

import (
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newConfigureCmd() *cobra.Command {
	var dryRun bool
	var env string
	var yes bool
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Interactive project setup for your provider(s)",
		Long:  "With --all or multiple detected providers, configures the full stack (e.g. Cloudflare + Vercel).",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}
			if err := project.EnsureOrbitDir(root); err != nil {
				return err
			}

			st, _ := state.Load(statePath(root))
			environment := env
			if environment == "" {
				environment = "production"
			}

			stack := detectStack(cmd.Context(), root)
			if len(stack) == 0 {
				return fmt.Errorf("no supported provider detected in this project")
			}

			activeScope := resolveStatusScope(cmd.Context(), root, st)

			targets := map[string]string{}
			var configureIDs []string

			useWizard := isInteractive() && !yes && providerFlag == "" && !allFlag && !dryRun

			switch {
			case allFlag:
				configureIDs = stack
			case providerFlag != "":
				configureIDs = []string{providerFlag}
				targets[providerFlag] = st.TargetFor(providerFlag)
			case useWizard:
				wizard, err := runConfigureWizard(cmd.Context(), root, st, stack, activeScope, environment)
				if err != nil {
					return err
				}
				if wizard.ConfigureAll {
					configureIDs = wizard.ScopeProviders
					if len(configureIDs) == 0 {
						configureIDs = stack
					}
					environment = wizard.Environment
				} else {
					configureIDs = []string{wizard.ProviderID}
					targets[wizard.ProviderID] = wizard.TargetID
					environment = wizard.Environment
				}
			default:
				if len(activeScope) > 0 && len(activeScope) < len(stack) {
					configureIDs = activeScope
				} else {
					configureIDs = stack
				}
			}

			if len(configureIDs) > 1 {
				fmt.Printf("Configuring stack: %s\n\n", providerListLabel(configureIDs))
			}

			if !dryRun {
				if err := ensureProvidersLoggedIn(cmd.Context(), configureIDs, isInteractive() && !yes); err != nil {
					return err
				}
			}

			st, err = configureStack(cmd.Context(), root, configureIDs, st, targets, environment, dryRun)
			if err != nil {
				return err
			}

			if !dryRun {
				return state.Save(statePath(root), st)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned changes without applying")
	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	cmd.Flags().StringVar(&providerFlag, "provider", "", "configure a single provider")
	cmd.Flags().BoolVar(&allFlag, "all", false, "configure all detected providers")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip interactive prompts")
	return cmd
}

func printConfigureResult(res provider.ConfigureResult) {
	if res.Message != "" {
		if res.OK {
			printSuccess(res.Message)
		} else {
			printError(res.Message)
		}
	}
	for _, change := range res.Changed {
		fmt.Printf("  %s %s\n", ui.dim.Render("updated"), styledPath(change))
	}
	for _, h := range res.Hints {
		fmt.Printf("  %s [%s]: %s\n", ui.info.Render("hint"), h.Code, h.Message)
		if h.Action != "" {
			fmt.Printf("  %s %s\n", ui.label.Render("action:"), highlightCmdLine(h.Action))
		}
	}
}
