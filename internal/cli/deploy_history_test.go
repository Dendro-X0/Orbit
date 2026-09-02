package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindPreviousDeployMatchesScope(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, ".orbit", "runs", "2026-09-02T06-35-20Z")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := `{
  "ok": true,
  "provider": "cloudflare",
  "apiUrl": "https://assess-api.example.workers.dev",
  "duration": "14.95s"
}`
	if err := os.WriteFile(filepath.Join(runDir, "summary.json"), []byte(summary), 0o644); err != nil {
		t.Fatal(err)
	}

	record := findPreviousDeploy(root, []string{"cloudflare"})
	if record == nil || record.APIURL != "https://assess-api.example.workers.dev" {
		t.Fatalf("record = %+v", record)
	}
	if record := findPreviousDeploy(root, []string{"vercel"}); record != nil {
		t.Fatalf("expected no vercel deploy, got %+v", record)
	}
}

func TestFormatTimeAgo(t *testing.T) {
	if got := formatTimeAgo(time.Now().Add(-30 * time.Minute)); got != "30 minutes ago" {
		t.Fatalf("got %q", got)
	}
	if got := formatTimeAgo(time.Time{}); got != "unknown" {
		t.Fatalf("got %q", got)
	}
}
