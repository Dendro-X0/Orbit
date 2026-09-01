package cli

import (
	"context"

	"github.com/Dendro-X0/Orbit/internal/provider"
)

func pickProvider(root string) string {
	if providerFlag != "" {
		return providerFlag
	}
	for _, p := range provider.All() {
		det, err := p.Detect(context.Background(), root)
		if err == nil && det.Supported {
			return p.ID()
		}
	}
	return ""
}
