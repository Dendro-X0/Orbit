package cli

import (
	"testing"

	"github.com/Dendro-X0/Orbit/internal/provider"
	_ "github.com/Dendro-X0/Orbit/internal/providers/cloudflare"
	_ "github.com/Dendro-X0/Orbit/internal/providers/vercel"
)

func TestAuthGuidesRegistered(t *testing.T) {
	for _, id := range []string{"cloudflare", "vercel"} {
		guide, ok := provider.AuthGuideFor(id)
		if !ok {
			t.Fatalf("missing auth guide for %s", id)
		}
		if guide.CreateURL == "" || len(guide.Steps) == 0 || len(guide.OAuthSteps) == 0 {
			t.Fatalf("incomplete guide for %s: %+v", id, guide)
		}
	}
}
