package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
)

// Deploy order: backend/edge before frontend.
var stackOrder = []string{"cloudflare", "fly", "vercel", "netlify", "railway"}

func detectStack(ctx context.Context, root string) []string {
	supported := map[string]bool{}
	for _, p := range provider.All() {
		det, err := p.Detect(ctx, root)
		if err == nil && det.Supported {
			supported[p.ID()] = true
		}
	}

	var ordered []string
	for _, id := range stackOrder {
		if supported[id] {
			ordered = append(ordered, id)
			delete(supported, id)
		}
	}
	rest := make([]string, 0, len(supported))
	for id := range supported {
		rest = append(rest, id)
	}
	sort.Strings(rest)
	return append(ordered, rest...)
}

func resolveDeployProviders(ctx context.Context, root string, st state.Project, single string) ([]string, error) {
	if single != "" {
		if _, err := provider.Get(single); err != nil {
			return nil, err
		}
		return []string{single}, nil
	}

	stack := detectStack(ctx, root)
	if len(stack) == 0 {
		return nil, fmt.Errorf("no deployable providers detected — run: orbit configure")
	}
	if len(stack) == 1 {
		return stack, nil
	}
	return stack, nil
}

func buildDeploySteps(root string, st state.Project, ids []string, env string, session *run.Session) ([]run.Step, error) {
	var steps []run.Step
	for _, id := range ids {
		p, err := provider.Get(id)
		if err != nil {
			return nil, err
		}
		pp, ok := p.(provider.PhaseProvider)
		if !ok {
			return nil, fmt.Errorf("provider %s does not support phased deploy yet", id)
		}
		phases := pp.Phases(root, provider.DeployOptions{
			Environment: env,
			TargetID:    st.TargetFor(id),
		})
		for _, phase := range phases {
			steps = append(steps, run.Step{
				ID:    fmt.Sprintf("%s-%s", id, phase.ID),
				Title: fmt.Sprintf("[%s] %s", p.DisplayName(), phase.Title),
				Run:   phase.Run,
			})
		}
		if id == "cloudflare" && stackContains(ids, "vercel") && session != nil {
			steps = append(steps, wireAPIURLStep(session, root, st, env))
		}
	}
	return steps, nil
}

func providerListLabel(ids []string) string {
	return strings.Join(ids, "+")
}
