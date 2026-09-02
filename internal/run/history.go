package run

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DeployRecord is a successful deploy captured in .orbit/runs/.
type DeployRecord struct {
	RunDir     string
	Provider   string
	Command    string
	APIURL     string
	DocsURL    string
	URL        string
	Duration   string
	DeployedAt time.Time
}

// LastSuccessfulDeploy returns the most recent successful deploy for the given
// provider label (e.g. "cloudflare" or "cloudflare+vercel").
func LastSuccessfulDeploy(root, providerLabel string) (*DeployRecord, bool) {
	records, err := ListSuccessfulDeploys(root)
	if err != nil {
		return nil, false
	}
	for _, record := range records {
		if record.Provider == providerLabel {
			return &record, true
		}
	}
	return nil, false
}

// ListSuccessfulDeploys returns successful deploy records newest first.
func ListSuccessfulDeploys(root string) ([]DeployRecord, error) {
	runsDir := filepath.Join(root, ".orbit", "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var records []DeployRecord
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runDir := filepath.Join(runsDir, entry.Name())
		summary, err := LoadSummary(runDir)
		if err != nil || summary == nil || !summary.OK {
			continue
		}
		record := DeployRecord{
			RunDir:   filepath.ToSlash(rel(root, runDir)),
			Provider: summary.Provider,
			Command:  summary.Command,
			APIURL:   summary.APIURL,
			DocsURL:  summary.DocsURL,
			URL:      summary.URL,
			Duration: summary.Duration,
		}
		if t, ok := ParseRunDirTime(entry.Name()); ok {
			record.DeployedAt = t
		}
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].DeployedAt.After(records[j].DeployedAt)
	})
	return records, nil
}

// ParseRunDirTime parses timestamps from run directory names like 2026-09-02T06-35-20Z.
func ParseRunDirTime(name string) (time.Time, bool) {
	name = strings.TrimSpace(name)
	t, err := time.Parse("2006-01-02T15-04-05Z", name)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}
