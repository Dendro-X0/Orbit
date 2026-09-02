package run

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestSummary(t *testing.T, path string, summary *Summary) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(path, summary); err != nil {
		t.Fatal(err)
	}
}

func TestLastSuccessfulDeployMatchesProvider(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, ".orbit", "runs")
	oldRun := filepath.Join(runs, "2026-09-01T10-00-00Z")
	newRun := filepath.Join(runs, "2026-09-02T06-35-20Z")
	writeTestSummary(t, filepath.Join(oldRun, "summary.json"), &Summary{
		OK: true, Provider: "cloudflare", APIURL: "https://old.example", Duration: "10s",
	})
	writeTestSummary(t, filepath.Join(newRun, "summary.json"), &Summary{
		OK: true, Provider: "cloudflare", APIURL: "https://new.example", Duration: "14s",
	})

	record, ok := LastSuccessfulDeploy(root, "cloudflare")
	if !ok || record == nil {
		t.Fatal("expected match")
	}
	if record.APIURL != "https://new.example" {
		t.Fatalf("got %q", record.APIURL)
	}
}

func TestLastSuccessfulDeploySkipsOtherProviders(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, ".orbit", "runs", "2026-09-02T06-35-20Z")
	writeTestSummary(t, filepath.Join(runDir, "summary.json"), &Summary{
		OK: true, Provider: "cloudflare+vercel", APIURL: "https://api.example", DocsURL: "https://docs.example",
	})

	if _, ok := LastSuccessfulDeploy(root, "cloudflare"); ok {
		t.Fatal("expected no match for narrower scope")
	}
	record, ok := LastSuccessfulDeploy(root, "cloudflare+vercel")
	if !ok || record.DocsURL == "" {
		t.Fatalf("record = %+v ok=%v", record, ok)
	}
}

func TestParseRunDirTime(t *testing.T) {
	tm, ok := ParseRunDirTime("2026-09-02T06-35-20Z")
	if !ok {
		t.Fatal("expected parse ok")
	}
	if tm.Hour() != 6 || tm.Minute() != 35 {
		t.Fatalf("got %v", tm)
	}
}

func TestListSuccessfulDeploysNewestFirst(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, ".orbit", "runs")
	for _, name := range []string{"2026-09-01T10-00-00Z", "2026-09-03T10-00-00Z", "2026-09-02T06-35-20Z"} {
		dir := filepath.Join(runs, name)
		writeTestSummary(t, filepath.Join(dir, "summary.json"), &Summary{OK: true, Provider: "cloudflare"})
	}
	records, err := ListSuccessfulDeploys(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Fatalf("len = %d", len(records))
	}
	if records[0].DeployedAt.Before(records[1].DeployedAt) {
		t.Fatalf("order wrong: %v then %v", records[0].DeployedAt, records[1].DeployedAt)
	}
	_ = time.Now()
}
