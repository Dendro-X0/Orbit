package run

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// LoadManifest reads manifest.json from a run directory.
func LoadManifest(runDir string) (*Manifest, error) {
	var m Manifest
	if err := loadJSON(filepath.Join(runDir, "manifest.json"), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// LoadSummary reads summary.json from a run directory.
func LoadSummary(runDir string) (*Summary, error) {
	var s Summary
	if err := loadJSON(filepath.Join(runDir, "summary.json"), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadFailure reads failure.json from a run directory.
func LoadFailure(runDir string) (*Failure, error) {
	var f Failure
	if err := loadJSON(filepath.Join(runDir, "failure.json"), &f); err != nil {
		return nil, err
	}
	return &f, nil
}
