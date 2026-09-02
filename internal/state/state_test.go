package state

import "testing"

func TestNormalizeMigratesLegacyFields(t *testing.T) {
	p := Project{
		Provider:    "cloudflare",
		TargetID:    "api",
		Configured:  true,
		Environment: "production",
	}
	p.Normalize()
	if !p.IsConfigured("cloudflare") {
		t.Fatal("expected cloudflare configured")
	}
	if p.TargetFor("cloudflare") != "api" {
		t.Fatalf("target = %q", p.TargetFor("cloudflare"))
	}
}

func TestSetProvider(t *testing.T) {
	var p Project
	p.SetProvider("vercel", "docs", true)
	if !p.IsConfigured("vercel") || p.TargetFor("vercel") != "docs" {
		t.Fatalf("state = %+v", p)
	}
}

func TestSetShipScope(t *testing.T) {
	var p Project
	p.SetShipScope("api_backend", "API / backend — Cloudflare", []string{"cloudflare"})
	if p.ShipLabel() != "API / backend — Cloudflare" {
		t.Fatalf("label = %q", p.ShipLabel())
	}
	if len(p.ShipProviders()) != 1 || p.ShipProviders()[0] != "cloudflare" {
		t.Fatalf("providers = %v", p.ShipProviders())
	}
}
