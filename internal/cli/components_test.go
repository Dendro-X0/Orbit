package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/Dendro-X0/Orbit/internal/providers/cloudflare"
	_ "github.com/Dendro-X0/Orbit/internal/providers/vercel"
)

func TestDetectDeployComponentsAssessLike(t *testing.T) {
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
	if len(components) != 2 {
		t.Fatalf("components = %d, want 2 (api + frontend)", len(components))
	}
	if components[0].Label != "API / backend" {
		t.Fatalf("first = %q", components[0].Label)
	}
	if components[1].Label != "Docs / web frontend" {
		t.Fatalf("second = %q", components[1].Label)
	}
	if components[1].Kind != "vite" {
		t.Fatalf("frontend kind = %q", components[1].Kind)
	}
}

func TestDetectDeployComponentsSingleAPI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "wrangler.toml"), []byte(`name="api"`), 0o644); err != nil {
		t.Fatal(err)
	}
	components := detectDeployComponents(t.Context(), root)
	if len(components) != 1 {
		t.Fatalf("components = %d", len(components))
	}
	if components[0].Providers[0] != "cloudflare" {
		t.Fatalf("providers = %v", components[0].Providers)
	}
}

func TestDetectDeployIntentsAssessLike(t *testing.T) {
	components := []deployComponent{
		{ID: "api:api", Label: "API / backend", Providers: []string{"cloudflare"}},
		{ID: "frontend:docs", Label: "Docs / web frontend", Providers: []string{"vercel"}, Kind: "vite"},
	}
	intents := detectDeployIntents(components)
	if len(intents) != 3 {
		t.Fatalf("intents = %d, want 3", len(intents))
	}
	if intents[0].ID != "api_backend" {
		t.Fatalf("first intent = %q, want api_backend", intents[0].ID)
	}
	if !strings.Contains(intents[0].Description, "Cloudflare only") {
		t.Fatalf("api scope = %q", intents[0].Description)
	}
	if intents[2].ID != "full_stack" {
		t.Fatalf("full_stack intent = %q", intents[2].ID)
	}
	if !strings.Contains(intents[2].Description, "2 providers") {
		t.Fatalf("full_stack scope = %q", intents[2].Description)
	}
}

func TestOrderDeployIntentsPutsAPIFirst(t *testing.T) {
	intents := []deployIntent{
		{ID: "full_stack", Label: "Full-stack"},
		{ID: "api_backend", Label: "API"},
		{ID: "web_frontend", Label: "Frontend"},
	}
	ordered := orderDeployIntents(intents)
	if ordered[0].ID != "api_backend" {
		t.Fatalf("first = %q", ordered[0].ID)
	}
}

func TestBuildFullStackPairingsAssessLike(t *testing.T) {
	components := []deployComponent{
		{ID: "api:api", Label: "API / backend", Providers: []string{"cloudflare"}},
		{ID: "frontend:docs", Label: "Docs / web frontend", Providers: []string{"vercel"}, Kind: "vite"},
	}
	pairings := buildFullStackPairings(components)
	if len(pairings) != 1 {
		t.Fatalf("pairings = %d, want 1", len(pairings))
	}
	if len(pairings[0].Providers) != 2 {
		t.Fatalf("providers = %v", pairings[0].Providers)
	}
	if pairings[0].Providers[0] != "cloudflare" || pairings[0].Providers[1] != "vercel" {
		t.Fatalf("providers = %v", pairings[0].Providers)
	}
}

func TestBuildFullStackPairingsNextJSOnly(t *testing.T) {
	components := []deployComponent{
		{ID: "frontend:root", Label: "Web application (Next.js)", Providers: []string{"vercel"}, Kind: "nextjs"},
	}
	pairings := buildFullStackPairings(components)
	if len(pairings) != 1 {
		t.Fatalf("pairings = %d", len(pairings))
	}
	if pairings[0].Providers[0] != "vercel" {
		t.Fatalf("providers = %v", pairings[0].Providers)
	}
}
