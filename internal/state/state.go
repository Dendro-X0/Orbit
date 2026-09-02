package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ProviderConfig holds per-provider setup for a project.
type ProviderConfig struct {
	TargetID   string `json:"targetId,omitempty"`
	Configured bool   `json:"configured"`
}

// ShipScope is the last deploy intent chosen in orbit ship.
type ShipScope struct {
	IntentID  string   `json:"intentId,omitempty"`
	Label     string   `json:"label,omitempty"`
	Providers []string `json:"providers,omitempty"`
}

// Project is persisted at .orbit/state.json.
type Project struct {
	Environment string                    `json:"environment,omitempty"`
	Providers   map[string]ProviderConfig `json:"providers,omitempty"`
	Ship        *ShipScope                `json:"ship,omitempty"`

	// Legacy single-provider fields (migrated into Providers on load).
	Provider   string `json:"provider,omitempty"`
	TargetID   string `json:"targetId,omitempty"`
	Configured bool   `json:"configured"`
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
	p.Normalize()
	return p, nil
}

func Save(path string, p Project) error {
	p.Normalize()
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

// Normalize migrates legacy single-provider state into Providers.
func (p *Project) Normalize() {
	if p.Providers == nil {
		p.Providers = map[string]ProviderConfig{}
	}
	if p.Provider != "" {
		if _, ok := p.Providers[p.Provider]; !ok {
			p.Providers[p.Provider] = ProviderConfig{
				TargetID:   p.TargetID,
				Configured: p.Configured,
			}
		}
	}
}

func (p *Project) SetProvider(id, targetID string, configured bool) {
	p.Normalize()
	cfg := p.Providers[id]
	if targetID != "" {
		cfg.TargetID = targetID
	}
	cfg.Configured = configured
	p.Providers[id] = cfg
	p.Provider = id
	p.TargetID = cfg.TargetID
	p.Configured = configured
}

func (p *Project) IsConfigured(id string) bool {
	p.Normalize()
	return p.Providers[id].Configured
}

func (p *Project) TargetFor(id string) string {
	p.Normalize()
	return p.Providers[id].TargetID
}

func (p *Project) ConfiguredProviders() []string {
	p.Normalize()
	var out []string
	for id, cfg := range p.Providers {
		if cfg.Configured {
			out = append(out, id)
		}
	}
	return out
}

func (p *Project) SetShipScope(intentID, label string, providers []string) {
	p.Normalize()
	if intentID == "" && len(providers) == 0 {
		p.Ship = nil
		return
	}
	copied := append([]string(nil), providers...)
	p.Ship = &ShipScope{
		IntentID:  intentID,
		Label:     label,
		Providers: copied,
	}
}

func (p *Project) ShipProviders() []string {
	p.Normalize()
	if p.Ship == nil || len(p.Ship.Providers) == 0 {
		return nil
	}
	return append([]string(nil), p.Ship.Providers...)
}

func (p *Project) ShipLabel() string {
	p.Normalize()
	if p.Ship == nil {
		return ""
	}
	return p.Ship.Label
}
