package provider

import (
	"context"

	"github.com/Dendro-X0/Orbit/internal/run"
)

// Provider owns stack detection, configuration, and deployment for a cloud platform.
// The core CLI orchestrates UX, auth shortcuts, and synchronized logging.
type Provider interface {
	ID() string
	DisplayName() string

	Detect(ctx context.Context, root string) (Detection, error)
	Configure(ctx context.Context, root string, opts ConfigureOptions) (ConfigureResult, error)
	Deploy(ctx context.Context, root string, opts DeployOptions) (DeployResult, error)

	// Login runs the provider's built-in auth shortcut (browser OAuth or token).
	Login(ctx context.Context) (LoginResult, error)
	// WhoAmI returns the active account if authenticated.
	WhoAmI(ctx context.Context) (WhoAmIResult, error)
	// Doctor checks provider CLI presence and auth.
	Doctor(ctx context.Context) ([]Check, error)
}

type Detection struct {
	Supported bool     `json:"supported"`
	Summary   string   `json:"summary,omitempty"`
	Targets   []Target `json:"targets,omitempty"`
	Hints     []Hint   `json:"hints,omitempty"`
}

type Target struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type ConfigureOptions struct {
	Environment string
	DryRun      bool
}

type ConfigureResult struct {
	OK       bool   `json:"ok"`
	Message  string `json:"message,omitempty"`
	Changed  []string `json:"changed,omitempty"`
	Hints    []Hint `json:"hints,omitempty"`
}

type DeployOptions struct {
	Environment string
	TargetID    string
	FromStep    string
}

type DeployResult struct {
	OK      bool   `json:"ok"`
	URL     string `json:"url,omitempty"`
	Message string `json:"message,omitempty"`
	LogsURL string `json:"logsUrl,omitempty"`
	Hints   []Hint `json:"hints,omitempty"`
}

type LoginResult struct {
	OK      bool   `json:"ok"`
	Account string `json:"account,omitempty"`
	Message string `json:"message,omitempty"`
}

type WhoAmIResult struct {
	LoggedIn bool   `json:"loggedIn"`
	Account  string `json:"account,omitempty"`
	Message  string `json:"message,omitempty"`
}

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Fix     string `json:"fix,omitempty"`
}

type Hint struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Action  string `json:"action,omitempty"`
}

// PhaseProvider exposes deploy phases for synchronized logging.
type PhaseProvider interface {
	Provider
	Phases(root string, opts DeployOptions) []run.Step
}
