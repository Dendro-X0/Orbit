package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show project configuration and last deploy run",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			st, _ := state.Load(statePath(root))
			stack := detectStack(cmd.Context(), root)
			scope := resolveStatusScope(cmd.Context(), root, st)

			printTitle("orbit status")
			fmt.Println()
			printLabeled(13, "Project", root)

			if len(stack) > 0 {
				printLabeled(13, "Detected", strings.Join(stack, ", "))
			}

			env := st.Environment
			if env == "" {
				env = "production"
			}
			printLabeled(13, "Environment", env)

			if label := st.ShipLabel(); label != "" {
				printLabeled(13, "Active scope", label)
			} else if len(scope) > 0 && len(scope) < len(stack) {
				printLabeled(13, "Active scope", "last deploy — "+providerListLabel(scope))
			}

			printAuthStatus(cmd.Context(), scope)
			printPendingSetup(scope, st)
			printConfiguredProviders(scope, st)

			printScopeDeployStatus(root, scope)

			printNextSteps(cmd.Context(), root, scope, st)
			summary := scopeSummary(root, scope)
			printRecommendedNext(cmd.Context(), root, st, scope, summary)
			if stackContains(scope, "cloudflare") {
				if msg := cloudflareSecretsSummary(cmd.Context(), root, st); msg != "" {
					fmt.Println()
					fmt.Printf("%s %s\n", ui.warn.Render("Secrets:"), ui.warn.Render(msg))
				}
			}
			printToolkitHints(root)
			return nil
		},
	}
}

func printConfiguredProviders(scope []string, st state.Project) {
	configured := false
	for _, id := range scope {
		if st.IsConfigured(id) {
			configured = true
			break
		}
	}
	if !configured {
		return
	}
	fmt.Println()
	printSection("Configured")
	for _, id := range scope {
		if !st.IsConfigured(id) {
			continue
		}
		target := st.TargetFor(id)
		if target != "" {
			printBullet(styledProvider(id) + ui.dim.Render(" (") + ui.value.Render(target) + ui.dim.Render(")"))
		} else {
			printBullet(styledProvider(id))
		}
	}
}

func printScopeDeployStatus(root string, scope []string) {
	label := scopeProviderLabel(scope)
	if label == "" {
		return
	}

	record, ok := run.LastSuccessfulDeploy(root, label)
	if ok && record != nil {
		fmt.Println()
		printSection("Deployed (this scope)")
		if !record.DeployedAt.IsZero() {
			printKV("When", formatTimeAgo(record.DeployedAt))
		}
		if record.Duration != "" {
			printKV("Duration", record.Duration)
		}
		if record.APIURL != "" {
			printKV("API URL", record.APIURL)
		}
		if record.DocsURL != "" {
			printKV("Docs URL", record.DocsURL)
		}
		if record.RunDir != "" {
			printKV("Run", record.RunDir)
		}
	}

	runDir, err := run.LatestRunDir(root)
	if err != nil {
		if !ok {
			fmt.Println()
			printKVPlain("Last run", "(none)")
		}
		return
	}

	summary, _ := run.LoadSummary(runDir)
	failure, _ := run.LoadFailure(runDir)
	if summary != nil && summary.OK && summary.Provider == label {
		return
	}
	if failure != nil && failure.Provider != label && failure.Provider != "" {
		return
	}

	relRun, _ := filepath.Rel(root, runDir)
	if failure != nil {
		fmt.Println()
		printKV("Last run", filepath.ToSlash(relRun))
		fmt.Printf("  %s %s\n", ui.label.Render("Status:"), ui.error.Render("failed at "+failure.FailedStep))
		fmt.Printf("  %s %s\n", ui.label.Render("Error:"), ui.error.Render(failure.Message))
		if failure.Hint != nil && failure.Hint.Action != "" {
			fmt.Printf("  %s %s\n", ui.label.Render("Action:"), highlightCmdLine(failure.Hint.Action))
		}
		printKV("Logs", failure.LogPaths.Combined)
		fmt.Printf("  %s %s\n", ui.label.Render("Retry:"), highlightCmdLine("orbit retry"))
	} else if !ok {
		fmt.Println()
		printKVPlain("Last run", filepath.ToSlash(relRun))
	}
}

func printAuthStatus(ctx context.Context, scope []string) {
	if len(scope) == 0 {
		return
	}
	fmt.Println()
	printSection("Authentication")
	for _, id := range scope {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		who, err := p.WhoAmI(ctx)
		if err != nil {
			fmt.Printf("  %s %-12s %s\n", ui.dim.Render("•"), id, ui.error.Render(fmt.Sprintf("error: %v", err)))
			continue
		}
		if who.LoggedIn {
			line := "logged in"
			if who.Account != "" {
				line = who.Account
			}
			fmt.Printf("  %s %-12s %s\n", okMark(), styledProvider(id), ui.value.Render(line))
		} else {
			fmt.Printf("  %s %-12s %s\n", failMark(), styledProvider(id), highlightCmdLine("orbit login "+id))
		}
	}
}

func printPendingSetup(scope []string, st state.Project) {
	var pending []string
	for _, id := range scope {
		if !st.IsConfigured(id) {
			pending = append(pending, id)
		}
	}
	if len(pending) == 0 {
		return
	}
	fmt.Println()
	printSection("Pending setup")
	for _, id := range pending {
		fmt.Printf("  %s %s\n", styledProvider(id), highlightCmdLine("→ orbit configure --provider "+id))
	}
}

func scopeSummary(root string, scope []string) *run.Summary {
	record, ok := run.LastSuccessfulDeploy(root, scopeProviderLabel(scope))
	if !ok || record == nil {
		return nil
	}
	return &run.Summary{
		APIURL:  record.APIURL,
		DocsURL: record.DocsURL,
	}
}

func printNextSteps(ctx context.Context, root string, scope []string, st state.Project) {
	if len(scope) == 0 {
		return
	}

	label := scopeProviderLabel(scope)
	record, hasDeploy := run.LastSuccessfulDeploy(root, label)

	var steps []string
	for _, id := range scope {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		who, _ := p.WhoAmI(ctx)
		if !who.LoggedIn {
			steps = append(steps, fmt.Sprintf("orbit login %s", id))
			break
		}
	}
	for _, id := range scope {
		if !st.IsConfigured(id) {
			if len(scope) == 1 {
				steps = append(steps, fmt.Sprintf("orbit configure --provider %s", scope[0]))
			} else {
				steps = append(steps, "orbit ship")
			}
			break
		}
	}

	if !hasDeploy && len(st.ConfiguredProviders()) > 0 {
		steps = append(steps, "orbit ship")
	}
	if hasDeploy && record != nil {
		if record.APIURL != "" {
			steps = append(steps, "orbit open --target api")
		}
		if record.DocsURL != "" {
			steps = append(steps, "orbit open --target docs")
		}
		if stackContains(scope, "cloudflare") && cloudflareSecretsSummary(ctx, root, st) != "" {
			steps = append(steps, "orbit secrets")
		}
	}
	if len(steps) == 0 {
		return
	}
	fmt.Println()
	printSection("Next")
	for _, s := range steps {
		printNextStep(s)
	}
}
