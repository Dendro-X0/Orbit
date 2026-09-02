package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/providers/cloudflare"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newShipCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ship",
		Short: "Guided deploy — choose what to deploy, then go",
		Long: `Interactive deploy workflow.

Step 1: project type (API, static site, or full-stack).
Step 2: provider(s) — one for API or frontend, one or two for full-stack.
Then log in, configure, set secrets, or deploy.

Orbit checks existing CLI sessions before prompting for browser OAuth.
Requires an interactive terminal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShipWorkflow(cmd)
		},
	}
}

func newMenuCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "menu",
		Short: "Interactive command picker",
		RunE:  runMenuLoop,
	}
}

func runMenuLoop(cmd *cobra.Command, _ []string) error {
	if !isInteractive() {
		printNonInteractiveHelp()
		return nil
	}
	for {
		action, err := runMainMenu()
		if err != nil {
			return err
		}
		if action == "quit" {
			return nil
		}
		if action == "ship" {
			if err := runShipWorkflow(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		} else if err := runMenuAction(cmd, action); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		}
		fmt.Println()
	}
}

func runShipWorkflow(cmd *cobra.Command) error {
	if !isInteractive() {
		return fmt.Errorf("orbit ship requires an interactive terminal — run individual commands instead (orbit login, orbit configure, orbit deploy)")
	}

	root, err := projectRoot(cmd)
	if err != nil {
		return err
	}
	if err := project.EnsureOrbitDir(root); err != nil {
		return err
	}

	ctx := cmd.Context()
	if len(detectStack(ctx, root)) == 0 {
		return fmt.Errorf("no deployable providers detected in %s", root)
	}

	st, _ := state.Load(statePath(root))
	printRecentFailure(root, resolveStatusScope(ctx, root, st))

	sel, err := runShipDeploySelect(ctx, root)
	if err != nil {
		return err
	}
	if err := saveShipScope(root, &st, sel); err != nil {
		return err
	}

	st, _ = state.Load(statePath(root))
	environment := "production"
	if st.Environment != "" {
		environment = st.Environment
	}

	selected := sel.Providers
	deployLabel := sel.Label

	var previousDeploy *run.DeployRecord
	refreshDeployContext := func() {
		previousDeploy = findPreviousDeploy(root, selected)
		printPreviousDeploy(previousDeploy)
	}
	refreshDeployContext()

	for {
		fmt.Println()
		fmt.Printf("%s %s%s\n",
			ui.label.Render("Scope:"),
			ui.title.Render(deployLabel),
			ui.dim.Render(" ("+environment+")"),
		)
		menuCtx := shipMenuContext{
			PreviousDeploy: previousDeploy,
			HasSecretsGap:  hasSecretsGap(cmd.Context(), root, st, selected),
		}
		action, err := runShipActionMenu(menuCtx)
		if err != nil {
			return err
		}
		if action == "quit" {
			return nil
		}

		switch action {
		case "scope":
			sel, err = runShipDeploySelect(ctx, root)
			if err != nil {
				return err
			}
			if err := saveShipScope(root, &st, sel); err != nil {
				return err
			}
			st, _ = state.Load(statePath(root))
			selected = sel.Providers
			deployLabel = sel.Label
			refreshDeployContext()
		case "full":
			ok, err := confirmRedeploy(previousDeploy)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if err := runShipFullPath(cmd, root, selected, &st, environment); err != nil {
				return err
			}
			return offerOpenDeployURLs(root)
		case "login":
			if err := ensureProvidersLoggedIn(ctx, selected, true); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		case "configure":
			if err := runShipConfigure(ctx, root, selected, &st, environment); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		case "secrets":
			if err := ensureCloudflareSecrets(ctx, root, st, selected); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		case "deploy":
			ok, err := confirmRedeploy(previousDeploy)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if err := runDeployForProviders(cmd, environment, selected); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			} else {
				refreshDeployContext()
			}
		case "open-api":
			if err := openDeployRecordURL(previousDeploy, "api"); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		case "open-docs":
			if err := openDeployRecordURL(previousDeploy, "docs"); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		case "status":
			if err := newStatusCmd().RunE(cmd, nil); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			}
		}
	}
}

func hasSecretsGap(ctx context.Context, root string, st state.Project, selected []string) bool {
	if !stackContains(selected, "cloudflare") {
		return false
	}
	return cloudflareSecretsSummary(ctx, root, st) != ""
}

func runShipFullPath(cmd *cobra.Command, root string, selected []string, st *state.Project, environment string) error {
	ctx := cmd.Context()

	printSection("── Prepare and deploy ──")

	if err := ensureProvidersLoggedIn(ctx, selected, true); err != nil {
		return err
	}
	fmt.Println()

	if err := runShipConfigure(ctx, root, selected, st, environment); err != nil {
		return err
	}
	fmt.Println()

	if err := ensureCloudflareSecrets(ctx, root, *st, selected); err != nil {
		return err
	}
	fmt.Println()

	return runDeployForProviders(cmd, environment, selected)
}

func runShipConfigure(ctx context.Context, root string, selected []string, st *state.Project, environment string) error {
	if err := ensureProvidersLoggedIn(ctx, selected, true); err != nil {
		return err
	}
	var err error
	*st, err = configureStack(ctx, root, selected, *st, nil, environment, false)
	if err != nil {
		return err
	}
	return state.Save(statePath(root), *st)
}

func ensureCloudflareSecrets(ctx context.Context, root string, st state.Project, selected []string) error {
	if !stackContains(selected, "cloudflare") {
		fmt.Println(ui.dim.Render("No Cloudflare worker secrets for this deploy target."))
		return nil
	}

	p, err := provider.Get("cloudflare")
	if err != nil {
		return err
	}
	det, err := p.Detect(ctx, root)
	if err != nil {
		return err
	}
	if !det.Supported {
		return nil
	}

	targetPath := targetPathFor(det, st.TargetFor("cloudflare"))
	env := cloudflare.CmdEnv()
	required, _, missing, err := cloudflare.SecretStatus(ctx, root, targetPath, env)
	if err != nil {
		return err
	}
	if len(required) == 0 {
		fmt.Println(ui.dim.Render("No secrets documented in wrangler.toml."))
		return nil
	}
	if len(missing) == 0 {
		printSuccess("All documented worker secrets are set.")
		return nil
	}

	printWarning(fmt.Sprintf("%d worker secret(s) still missing on Cloudflare.", len(missing)))
	fmt.Println(ui.dim.Render("Orbit will run wrangler secret put for each — you enter values when prompted."))
	fmt.Println()

	workDir := filepath.Join(root, targetPath)
	for _, name := range missing {
		setNow := true
		if err := huh.NewConfirm().
			Title(fmt.Sprintf("Set secret %s?", name)).
			Description("wrangler will prompt for the value (input is hidden).").
			Value(&setNow).
			Run(); err != nil {
			return err
		}
		if !setNow {
			fmt.Printf("  %s %s\n", ui.dim.Render("skipped"), ui.value.Render(name))
			continue
		}
		fmt.Printf("Setting %s…\n", ui.info.Render(name))
		if err := cloudflare.PutSecret(ctx, workDir, name, env); err != nil {
			return err
		}
		printIndentedSuccess(name + " set")
	}

	_, _, stillMissing, err := cloudflare.SecretStatus(ctx, root, targetPath, env)
	if err != nil {
		return err
	}
	if len(stillMissing) == 0 {
		printSuccess("All documented worker secrets are set.")
		return nil
	}

	fmt.Printf("\n%s %s\n", warnMark(), ui.warn.Render(fmt.Sprintf("%d secret(s) still missing: %s", len(stillMissing), strings.Join(stillMissing, ", "))))
	proceed := false
	if err := huh.NewConfirm().
		Title("Continue to deploy anyway?").
		Description("Deploy may fail until secrets are set.").
		Value(&proceed).
		Run(); err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("deploy cancelled — set remaining secrets with: orbit secrets")
	}
	return nil
}

func offerOpenDeployURLs(root string) error {
	open := true
	if err := huh.NewConfirm().
		Title("Open deploy URLs in your browser?").
		Value(&open).
		Run(); err != nil {
		return err
	}
	if !open {
		fmt.Println(ui.dim.Render("\n✓ Ship complete. Run: orbit status"))
		return nil
	}

	opened := false
	for _, target := range []string{"api", "docs"} {
		url, err := resolveOpenURL(root, target)
		if err != nil {
			continue
		}
		if err := openURL(url); err != nil {
			fmt.Fprintf(os.Stderr, "could not open %s: %v\n", url, err)
			continue
		}
		fmt.Printf("Opened %s (%s)\n", ui.info.Render(target), styledURL(url))
		opened = true
	}
	if !opened {
		fmt.Println(ui.dim.Render("No deploy URLs found yet — run: orbit status"))
	} else {
		printSuccess("Ship complete.")
	}
	return nil
}

func runDeploy(cmd *cobra.Command, deployEnv, singleProvider string) error {
	return runDeployForProviders(cmd, deployEnv, nil, singleProvider)
}

func runDeployForProviders(cmd *cobra.Command, deployEnv string, selected []string, singleProvider ...string) error {
	root, err := projectRoot(cmd)
	if err != nil {
		return err
	}
	if err := project.EnsureOrbitDir(root); err != nil {
		return err
	}

	st, _ := state.Load(statePath(root))
	var ids []string
	if len(selected) > 0 {
		ids = selected
	} else {
		single := ""
		if len(singleProvider) > 0 {
			single = singleProvider[0]
		}
		ids, err = resolveDeployProviders(cmd.Context(), root, st, single)
		if err != nil {
			return err
		}
	}

	if deployEnv == "production" && st.Environment != "" {
		deployEnv = st.Environment
	}

	if err := ensureProvidersLoggedIn(cmd.Context(), ids, isInteractive()); err != nil {
		return err
	}

	st, err = ensureConfigured(cmd.Context(), root, st, ids, deployEnv)
	if err != nil {
		return err
	}

	session := run.Session{}
	steps, err := buildDeploySteps(root, st, ids, deployEnv, &session)
	if err != nil {
		return err
	}

	label := providerListLabel(ids)
	if isInteractive() && len(ids) > 0 {
		if record, ok := run.LastSuccessfulDeploy(root, label); ok {
			printPreviousDeploy(record)
			proceed, err := confirmRedeploy(record)
			if err != nil {
				return err
			}
			if !proceed {
				return fmt.Errorf("deploy cancelled")
			}
			fmt.Println()
		}
	}
	fmt.Printf("%s %s (%s)\n\n", ui.title.Render("orbit deploy"), ui.value.Render("—"), styledValue(label+" / "+deployEnv))

	r := &run.Runner{}
	result, err := r.Execute(cmd.Context(), run.Options{
		Root:      root,
		Provider:  label,
		Command:   "deploy",
		PrintLive: true,
		Session:   &session,
	}, steps)

	return printDeployResult(cmd.Context(), root, st, label, result, err)
}

func printNonInteractiveHelp() {
	fmt.Println("orbit — deploy to your cloud, simply.")
	fmt.Println()
	fmt.Println("  orbit ship                         pick project type, then deploy")
	fmt.Println("  orbit menu                         command picker")
	fmt.Println("  orbit deploy --provider cloudflare deploy one provider")
	fmt.Println("  orbit status                       scoped status + deploy history")
	fmt.Println()
	fmt.Println("Run `orbit <command> --help` for details.")
}
