package cli

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/Dendro-X0/Orbit/internal/providers/cloudflare"
	_ "github.com/Dendro-X0/Orbit/internal/providers/vercel"
)

func TestDetectProjectProfileAssessLike(t *testing.T) {
	root := t.TempDir()
	api := filepath.Join(root, "apps", "api")
	docs := filepath.Join(root, "apps", "docs")
	if err := os.MkdirAll(api, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(api, "wrangler.toml"), []byte(`name="api"`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "vercel.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "vite.config.ts"), []byte(`export default {}`), 0o644); err != nil {
		t.Fatal(err)
	}

	components := detectDeployComponents(t.Context(), root)
	profile := detectProjectProfile(root, components)
	if profile.Type != "full_web_product" {
		t.Fatalf("type = %q", profile.Type)
	}
	if profile.SuggestedID != "full_stack" {
		t.Fatalf("suggested = %q", profile.SuggestedID)
	}
	if !signalContains(profile.Signals, "monorepo:api+docs") {
		t.Fatalf("signals = %v", profile.Signals)
	}
}

func TestDetectProjectProfileAPIOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "wrangler.toml"), []byte(`name="api"`), 0o644); err != nil {
		t.Fatal(err)
	}
	components := detectDeployComponents(t.Context(), root)
	profile := detectProjectProfile(root, components)
	if profile.Type != "api_backend" {
		t.Fatalf("type = %q", profile.Type)
	}
	if profile.SuggestedID != "api_backend" {
		t.Fatalf("suggested = %q", profile.SuggestedID)
	}
}

func TestReorderIntentsSuggestedFirst(t *testing.T) {
	intents := []deployIntent{
		{ID: "api_backend", Label: "API"},
		{ID: "full_stack", Label: "Full-stack"},
		{ID: "web_frontend", Label: "Frontend"},
	}
	ordered := reorderIntentsSuggestedFirst(intents, "full_stack")
	if ordered[0].ID != "full_stack" {
		t.Fatalf("first = %q", ordered[0].ID)
	}
}

func TestReorderSuggestedFirst(t *testing.T) {
	parts := []deployComponent{
		{ID: "api:api", Label: "API"},
		{ID: "frontend:docs", Label: "Docs"},
	}
	ordered := reorderSuggestedFirst(parts, "frontend:docs")
	if ordered[0].ID != "frontend:docs" {
		t.Fatalf("first = %q", ordered[0].ID)
	}
}
