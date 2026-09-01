package cli

import (
	"context"
	"testing"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
)

type stubProvider struct {
	id     string
	phases int
}

func (s stubProvider) ID() string          { return s.id }
func (s stubProvider) DisplayName() string { return s.id }
func (s stubProvider) Detect(ctx context.Context, root string) (provider.Detection, error) {
	return provider.Detection{Supported: true, Summary: "stub"}, nil
}
func (s stubProvider) Configure(ctx context.Context, root string, opts provider.ConfigureOptions) (provider.ConfigureResult, error) {
	return provider.ConfigureResult{OK: true}, nil
}
func (s stubProvider) Deploy(ctx context.Context, root string, opts provider.DeployOptions) (provider.DeployResult, error) {
	return provider.DeployResult{OK: true}, nil
}
func (s stubProvider) Login(ctx context.Context) (provider.LoginResult, error) {
	return provider.LoginResult{OK: true}, nil
}
func (s stubProvider) WhoAmI(ctx context.Context) (provider.WhoAmIResult, error) {
	return provider.WhoAmIResult{LoggedIn: true}, nil
}
func (s stubProvider) Doctor(ctx context.Context) ([]provider.Check, error) {
	return nil, nil
}
func (s stubProvider) Phases(root string, opts provider.DeployOptions) []run.Step {
	steps := make([]run.Step, s.phases)
	for i := range steps {
		steps[i] = run.Step{ID: "step", Title: "step"}
	}
	return steps
}

func TestBuildDeployStepsMultiProvider(t *testing.T) {
	provider.Register(stubProvider{id: "stub-a", phases: 2})
	provider.Register(stubProvider{id: "stub-b", phases: 2})

	st := state.Project{}
	st.SetProvider("stub-a", "api", true)
	st.SetProvider("stub-b", "docs", true)

	steps, err := buildDeploySteps(t.TempDir(), st, []string{"stub-a", "stub-b"}, "production")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 4 {
		t.Fatalf("steps = %d, want 4", len(steps))
	}
	if steps[0].ID != "stub-a-step" {
		t.Fatalf("first step id = %q", steps[0].ID)
	}
}

func TestProviderListLabel(t *testing.T) {
	if got := providerListLabel([]string{"cloudflare", "vercel"}); got != "cloudflare+vercel" {
		t.Fatalf("got %q", got)
	}
}
