package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
)

func printRecentFailure(root string, scopeProviders []string) {
	var record *run.FailureRecord
	var ok bool

	if len(scopeProviders) > 0 {
		record, ok = run.LastFailedRunForProvider(root, providerListLabel(scopeProviders))
	}
	if !ok {
		record, ok = run.LastFailedRun(root)
	}
	if !ok || record == nil || record.Failure == nil {
		return
	}

	if record.Provider != "" {
		if success, found := run.LastSuccessfulDeploy(root, record.Provider); found && success != nil {
			if !success.DeployedAt.IsZero() && !record.FailedAt.IsZero() &&
				success.DeployedAt.After(record.FailedAt) {
				return
			}
		}
	}

	fmt.Println()
	printSection("Last deploy failed")
	printKV("Step", record.FailedStep)
	printKV("Error", record.Failure.Message)
	if record.Provider != "" {
		printKV("Provider", record.Provider)
	}
	if !record.FailedAt.IsZero() {
		printKV("When", formatTimeAgo(record.FailedAt))
	}
	if record.RunDir != "" {
		printKV("Logs", record.RunDir)
	}
	printKV("Retry", "orbit retry")
}

func printRecommendedNext(ctx context.Context, root string, st state.Project, scopeProviders []string, summary *run.Summary) {
	steps := buildRecommendedSteps(ctx, root, st, scopeProviders, summary)
	if len(steps) == 0 {
		return
	}
	fmt.Println()
	printSection("Recommended next")
	for _, step := range steps {
		printBullet(step)
	}
}

func buildRecommendedSteps(ctx context.Context, root string, st state.Project, scopeProviders []string, summary *run.Summary) []string {
	var steps []string

	if stackContains(scopeProviders, "cloudflare") {
		if cloudflareSecretsSummary(ctx, root, st) != "" {
			steps = append(steps, ui.cmd.Render("orbit secrets")+ui.dim.Render(" — set missing worker secrets"))
		}
		steps = append(steps, corsRecommendations(ctx, root, st, summary)...)
	}

	if summary != nil && summary.APIURL != "" && !stackContains(scopeProviders, "vercel") {
		stack := detectStack(ctx, root)
		if stackContains(stack, "vercel") {
			steps = append(steps, "Deploy docs — run "+ui.cmd.Render("orbit ship")+ui.dim.Render(" and pick Static site / docs"))
		}
	}

	if summary != nil && summary.APIURL != "" {
		health := apiHealthURL(ctx, root, st, summary.APIURL)
		steps = append(steps, "Health check — "+ui.url.Render(health))
	}

	return dedupeStrings(steps)
}

func corsRecommendations(ctx context.Context, root string, st state.Project, summary *run.Summary) []string {
	wranglerPath := wranglerPathForCloudflare(ctx, root, st)
	if wranglerPath == "" {
		return nil
	}
	relPath, err := filepath.Rel(root, wranglerPath)
	if err != nil {
		relPath = wranglerPath
	}
	relPath = filepath.ToSlash(relPath)

	origins := readWranglerCORSOrigins(wranglerPath)
	if len(origins) == 0 {
		return nil
	}

	docsURL := ""
	if summary != nil {
		docsURL = summary.DocsURL
	}

	var steps []string
	if docsURL != "" && corsMissingOrigin(origins, docsURL) {
		steps = append(steps, fmt.Sprintf(
			"Add %s to CORS_ORIGINS in %s, then re-deploy the API",
			ui.url.Render(docsURL),
			ui.path.Render(relPath),
		))
	} else if corsOnlyLocalDev(origins) && summary != nil && summary.APIURL != "" {
		steps = append(steps, fmt.Sprintf(
			"CORS_ORIGINS in %s is localhost-only — add your production docs URL before shipping the frontend",
			ui.path.Render(relPath),
		))
	}
	return steps
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	var out []string
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
