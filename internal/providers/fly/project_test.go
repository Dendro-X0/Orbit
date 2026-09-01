package fly

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFlyTargets(t *testing.T) {
	root := t.TempDir()
	api := filepath.Join(root, "apps", "api")
	if err := os.MkdirAll(api, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(api, "fly.toml"), []byte(`app = "my-api"`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(api, "Dockerfile"), []byte(`FROM alpine`), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := findFlyTargets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].ID != "api" || targets[0].Kind != "container" {
		t.Fatalf("got %+v", targets[0])
	}
}

func TestParseAppName(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "fly.toml")
	content := `# fly.toml
app = 'assess-api'

[build]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	name, err := parseAppName(path)
	if err != nil {
		t.Fatal(err)
	}
	if name != "assess-api" {
		t.Fatalf("got %q", name)
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
