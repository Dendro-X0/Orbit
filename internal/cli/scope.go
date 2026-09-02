package cli

import (
	"context"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
)

// resolveStatusScope returns providers relevant to the user's active ship intent.
// Falls back to the latest successful deploy, then the full detected stack.
func resolveStatusScope(ctx context.Context, root string, st state.Project) []string {
	if providers := st.ShipProviders(); len(providers) > 0 {
		return providers
	}
	if records, err := run.ListSuccessfulDeploys(root); err == nil && len(records) > 0 && records[0].Provider != "" {
		if ids := parseProviderIDs(records[0].Provider); len(ids) > 0 {
			return ids
		}
	}
	return detectStack(ctx, root)
}

func scopeProviderLabel(providers []string) string {
	if len(providers) == 0 {
		return ""
	}
	return providerListLabel(providers)
}
