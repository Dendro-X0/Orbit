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

	fmt.Printf("Configuring %s at %s\n\n", p.DisplayName(), root)
	res, err := p.Configure(ctx, root, provider.ConfigureOptions{
		Environment: environment,
		TargetID:    targetID,
		DryRun:      dryRun,
	})
	if err != nil {
		return err
	}

	printConfigureResult(res)

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

	fmt.Printf("Configuring %d provider(s) before deploy…\n\n", len(pending))
	return configureStack(ctx, root, pending, st, nil, env, false)
}

func printDeployResult(result *run.Result, err error) error {
	if result != nil && result.Summary != nil {
		fmt.Printf("\n✓ Deployed in %s\n", result.Summary.Duration)
		fmt.Printf("  Logs: %s\n", result.Summary.RunDir)
	}
	if result != nil && result.Failure != nil {
		fmt.Fprintf(os.Stderr, "\n✗ Deploy failed at step %s\n", result.Failure.FailedStep)
		fmt.Fprintf(os.Stderr, "  %s\n", result.Failure.Message)
		if result.Failure.Hint != nil && result.Failure.Hint.Action != "" {
			fmt.Fprintf(os.Stderr, "  action: %s\n", result.Failure.Hint.Action)
		}
		fmt.Fprintf(os.Stderr, "  logs: %s\n", result.Failure.LogPaths.Combined)
	}
	return err
}
