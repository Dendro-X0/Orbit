package run

import (
	"path/filepath"
	"testing"
)

func TestLastFailedRun(t *testing.T) {
	root := t.TempDir()
	failDir := filepath.Join(root, ".orbit", "runs", "2026-09-01T10-00-00Z")
	writeTestFailure(t, failDir, "cloudflare", "cloudflare-whoami", "not logged in")

	record, ok := LastFailedRun(root)
	if !ok || record.FailedStep != "cloudflare-whoami" {
		t.Fatalf("record = %+v ok=%v", record, ok)
	}
}

func writeTestFailure(t *testing.T, dir, provider, step, msg string) {
	t.Helper()
	writeTestSummary(t, filepath.Join(dir, "summary.json"), &Summary{OK: false, Provider: provider})
	if err := writeJSON(filepath.Join(dir, "failure.json"), &Failure{
		Provider: provider, FailedStep: step, Message: msg,
	}); err != nil {
		t.Fatal(err)
	}
}
