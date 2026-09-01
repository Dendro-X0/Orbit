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
