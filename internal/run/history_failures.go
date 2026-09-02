package run

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// FailureRecord is a failed deploy run from .orbit/runs/.
type FailureRecord struct {
	RunDir     string
	Failure    *Failure
	FailedAt   time.Time
	Provider   string
	FailedStep string
}

// LastFailedRun returns the most recent failed deploy run, if any.
func LastFailedRun(root string) (*FailureRecord, bool) {
	records, err := ListFailedRuns(root)
	if err != nil || len(records) == 0 {
		return nil, false
	}
	return &records[0], true
}

// LastFailedRunForProvider returns the latest failure for a provider label.
func LastFailedRunForProvider(root, providerLabel string) (*FailureRecord, bool) {
	records, err := ListFailedRuns(root)
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

// ListFailedRuns returns failed deploy records newest first.
func ListFailedRuns(root string) ([]FailureRecord, error) {
	runsDir := filepath.Join(root, ".orbit", "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var records []FailureRecord
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runDir := filepath.Join(runsDir, entry.Name())
		summary, _ := LoadSummary(runDir)
		if summary != nil && summary.OK {
			continue
		}
		failure, err := LoadFailure(runDir)
		if err != nil || failure == nil {
			continue
		}
		record := FailureRecord{
			RunDir:     filepath.ToSlash(rel(root, runDir)),
			Failure:    failure,
			Provider:   failure.Provider,
			FailedStep: failure.FailedStep,
		}
		if t, ok := ParseRunDirTime(entry.Name()); ok {
			record.FailedAt = t
		}
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].FailedAt.After(records[j].FailedAt)
	})
	return records, nil
}

// LatestRunFailed reports whether the most recent run directory has a failure.
func LatestRunFailed(root string) bool {
	runDir, err := LatestRunDir(root)
	if err != nil {
		return false
	}
	summary, _ := LoadSummary(runDir)
	if summary != nil && summary.OK {
		return false
	}
	failure, _ := LoadFailure(runDir)
	return failure != nil
}
