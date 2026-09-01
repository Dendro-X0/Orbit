package cli

import (
	"context"
	"testing"

	"github.com/Dendro-X0/Orbit/internal/provider"
)

type authStub struct {
	id      string
	logged  bool
	account string
}

func (s authStub) ID() string          { return s.id }
func (s authStub) DisplayName() string { return s.id }
func (s authStub) Detect(ctx context.Context, root string) (provider.Detection, error) {
	return provider.Detection{}, nil
}
func (s authStub) Configure(ctx context.Context, root string, opts provider.ConfigureOptions) (provider.ConfigureResult, error) {
	return provider.ConfigureResult{}, nil
}
func (s authStub) Deploy(ctx context.Context, root string, opts provider.DeployOptions) (provider.DeployResult, error) {
	return provider.DeployResult{}, nil
}
func (s authStub) Login(ctx context.Context) (provider.LoginResult, error) {
	return provider.LoginResult{OK: true}, nil
}
func (s authStub) WhoAmI(ctx context.Context) (provider.WhoAmIResult, error) {
	return provider.WhoAmIResult{LoggedIn: s.logged, Account: s.account}, nil
}
func (s authStub) Doctor(ctx context.Context) ([]provider.Check, error) {
	return nil, nil
}

func TestProvidersNeedingAuth(t *testing.T) {
	provider.Register(authStub{id: "stub-a", logged: true})
	provider.Register(authStub{id: "stub-b", logged: false})

	pending, err := providersNeedingAuth(context.Background(), []string{"stub-a", "stub-b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0] != "stub-b" {
		t.Fatalf("pending = %#v", pending)
	}
}

func TestConfigureNeedsAuth(t *testing.T) {
	if !configureNeedsAuth(provider.ConfigureResult{
		Hints: []provider.Hint{{Code: "auth.required"}},
	}) {
		t.Fatal("expected auth required")
	}
}

func TestAuthRequiredError(t *testing.T) {
	err := authRequiredError([]string{"cloudflare", "vercel"})
	if err == nil || err.Error() == "" {
		t.Fatal("expected error")
	}
}
