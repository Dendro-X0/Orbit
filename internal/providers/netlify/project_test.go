package netlify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNetlifyTargets(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "apps", "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "netlify.toml"), []byte(`[build]\n  publish = "dist"`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "vite.config.ts"), []byte(`export default {}`), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := findNetlifyTargets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].ID != "docs" || targets[0].Kind != "vite" {
		t.Fatalf("got %+v", targets[0])
	}
}

func TestReadSiteLink(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".netlify")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte(`{"siteId":"abc-123"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link, ok, err := readSiteLink(root, ".")
	if err != nil || !ok || link.SiteID != "abc-123" {
		t.Fatalf("link=%+v ok=%v err=%v", link, ok, err)
	}
}

func TestParseLoggedInUser(t *testing.T) {
	out := `Logged in as alice on My Team
Current site: my-site`
	if got := parseLoggedInUser(out); got != "alice" {
		t.Fatalf("got %q", got)
	}
}

func TestDetectUnsupported(t *testing.T) {
	p := New()
	det, err := p.Detect(t.Context(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if det.Supported {
		t.Fatal("expected unsupported")
	}
}
