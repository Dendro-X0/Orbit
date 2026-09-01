package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Project struct {
	Provider    string `json:"provider,omitempty"`
	Environment string `json:"environment,omitempty"`
	TargetID    string `json:"targetId,omitempty"`
	Configured  bool   `json:"configured"`
}

func Load(path string) (Project, error) {
	var p Project
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return p, err
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return p, err
	}
	return p, nil
}

func Save(path string, p Project) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}
