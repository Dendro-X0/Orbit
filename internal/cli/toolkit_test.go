package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Dendro-X0/Orbit/internal/run"
)

func TestSliceStepsFrom(t *testing.T) {
	steps := []run.Step{
		{ID: "cloudflare-whoami"},
		{ID: "cloudflare-deploy"},
		{ID: "wire-vite-api-url"},
	}
	got, err := sliceStepsFrom(steps, "wire-vite-api-url")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "wire-vite-api-url" {
		t.Fatalf("got %#v", got)
	}
}

func TestHasSignetProject(t *testing.T) {
	dir := t.TempDir()
	if hasSignetProject(dir) {
		t.Fatal("expected false")
	}
	if err := os.WriteFile(filepath.Join(dir, "signet.toml"), []byte("[project]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !hasSignetProject(dir) {
		t.Fatal("expected true")
	}
}

func TestParseProviderIDs(t *testing.T) {
	got := parseProviderIDs("cloudflare+vercel")
	if len(got) != 2 || got[0] != "cloudflare" || got[1] != "vercel" {
		t.Fatalf("got %#v", got)
	}
}
