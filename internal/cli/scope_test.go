package cli

import (
	"testing"
)

func TestFormatScopeLabel(t *testing.T) {
	intent := deployIntent{ID: "api_backend", Label: "API / backend"}
	got := formatScopeLabel(intent, []string{"cloudflare"}, "Cloudflare")
	if got != "API / backend — Cloudflare" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatScopeLabelFullStack(t *testing.T) {
	intent := deployIntent{ID: "full_stack", Label: "Full-stack"}
	detail := "Cloudflare (API) + Vercel (frontend)"
	got := formatScopeLabel(intent, []string{"cloudflare", "vercel"}, detail)
	if got != "Full-stack — Cloudflare (API) + Vercel (frontend)" {
		t.Fatalf("got %q", got)
	}
}
