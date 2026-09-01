package run

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadManifestSummaryFailure(t *testing.T) {
	dir := t.TempDir()
	manifest := &Manifest{ID: "test-run", OK: false, Command: "deploy", StartedAt: time.Now().UTC()}
	if err := writeJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	failure := &Failure{FailedStep: "cloudflare-whoami", Message: "auth"}
	if err := writeJSON(filepath.Join(dir, "failure.json"), failure); err != nil {
		t.Fatal(err)
	}

	m, err := LoadManifest(dir)
	if err != nil || m.ID != "test-run" {
		t.Fatalf("manifest: %v %#v", err, m)
	}
	f, err := LoadFailure(dir)
	if err != nil || f.FailedStep != "cloudflare-whoami" {
		t.Fatalf("failure: %v %#v", err, f)
	}
	if _, err := LoadSummary(dir); err == nil {
		t.Fatal("expected missing summary error")
	}
	_ = os.Remove(filepath.Join(dir, "failure.json"))
}
