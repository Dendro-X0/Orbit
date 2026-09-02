package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
)

func configureOne(
	ctx context.Context,
	root string,
	st *state.Project,
	provID, targetID, environment string,
	dryRun bool,
) error {
	p, err := provider.Get(provID)
	if err != nil {
		return err
	}

	fmt.Printf("Configuring %s at %s\n\n", styledProvider(p.DisplayName()), styledPath(root))
	res, err := p.Configure(ctx, root, provider.ConfigureOptions{
		Environment: environment,
		TargetID:    targetID,
		DryRun:      dryRun,
	})
	if err != nil {
		return err
	}

	printConfigureResult(res)

	if !res.OK && configureNeedsAuth(res) {
		return offerLoginAndRetryConfigure(ctx, root, st, provID, targetID, environment)
	}

	if res.OK && !dryRun {
		st.SetProvider(provID, targetID, true)
		if environment != "" {
			st.Environment = environment
		}
	} else if !res.OK {
		return fmt.Errorf("configure failed for %s", provID)
	}
	return nil
}

func configureStack(
	ctx context.Context,
	root string,
	ids []string,
	st state.Project,
	targetByProvider map[string]string,
	environment string,
	dryRun bool,
) (state.Project, error) {
	for _, id := range ids {
		targetID := targetByProvider[id]
		if targetID == "" {
			targetID = st.TargetFor(id)
		}
		if err := configureOne(ctx, root, &st, id, targetID, environment, dryRun); err != nil {
			return st, err
		}
		if !dryRun {
			if err := state.Save(statePath(root), st); err != nil {
				return st, err
			}
		}
		fmt.Println()
	}
	return st, nil
}

func ensureConfigured(
	ctx context.Context,
	root string,
	st state.Project,
	ids []string,
	env string,
) (state.Project, error) {
	var pending []string
	for _, id := range ids {
		if !st.IsConfigured(id) {
			pending = append(pending, id)
		}
	}
	if len(pending) == 0 {
		return st, nil
	}

	fmt.Printf("%s\n\n", ui.info.Render(fmt.Sprintf("Configuring %d provider(s) before deploy…", len(pending))))
	return configureStack(ctx, root, pending, st, nil, env, false)
}

func printDeployResult(ctx context.Context, root string, st state.Project, label string, result *run.Result, err error) error {
	if result != nil && result.Summary != nil {
		fmt.Println()
		printSuccess("Deployed in " + result.Summary.Duration)

		secretsGap := stackContains(parseProviderIDs(label), "cloudflare") &&
			cloudflareSecretsSummary(ctx, root, st) != ""
		if secretsGap {
			printWarning("Live with gaps — worker secrets still missing")
		} else {
			printSuccess("Live")
		}

		if result.Summary.APIURL != "" {
			printKV("API URL", result.Summary.APIURL)
		}
		if result.Summary.DocsURL != "" {
			printKV("Docs URL", result.Summary.DocsURL)
		}
		if result.Summary.APIURL != "" || result.Summary.DocsURL != "" {
			printKV("Open", "orbit open --target api|docs")
		}
		printKV("Logs", result.Summary.RunDir)

		scope := parseProviderIDs(label)
		printRecommendedNext(ctx, root, st, scope, result.Summary)
	}
	if result != nil && result.Failure != nil {
		fmt.Fprintf(os.Stderr, "\n%s %s\n", failMark(), ui.error.Render("Deploy failed at step "+result.Failure.FailedStep))
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.label.Render("Error:"), ui.error.Render(result.Failure.Message))
		if result.Failure.Hint != nil && result.Failure.Hint.Action != "" {
			fmt.Fprintf(os.Stderr, "  %s %s\n", ui.label.Render("Action:"), highlightCmdLine(result.Failure.Hint.Action))
		}
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.label.Render("Retry:"), highlightCmdLine("orbit retry"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.label.Render("Logs:"), styledPath(result.Failure.LogPaths.Combined))
	}
	return err
}
