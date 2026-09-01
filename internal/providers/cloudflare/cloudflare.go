package cloudflare

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
)

const ID = "cloudflare"

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) ID() string          { return ID }
func (p *Provider) DisplayName() string { return "Cloudflare" }

func (p *Provider) Detect(ctx context.Context, root string) (provider.Detection, error) {
	_ = ctx
	var targets []provider.Target

	apiWrangler := filepath.Join(root, "apps", "api", "wrangler.toml")
	if run.FileExists(apiWrangler) {
		targets = append(targets, provider.Target{ID: "api", Path: "apps/api", Kind: "workers"})
	}
	if run.FileExists(filepath.Join(root, "wrangler.toml")) {
		targets = append(targets, provider.Target{ID: "worker", Path: ".", Kind: "workers"})
	}

	if len(targets) == 0 {
		return provider.Detection{Supported: false, Summary: "No wrangler.toml found"}, nil
	}

	return provider.Detection{
		Supported: true,
		Summary:   fmt.Sprintf("Cloudflare Workers project (%d target(s))", len(targets)),
		Targets:   targets,
	}, nil
}

func (p *Provider) target(root, targetID string) (provider.Target, error) {
	det, err := p.Detect(context.Background(), root)
	if err != nil {
		return provider.Target{}, err
	}
	if !det.Supported {
		return provider.Target{}, fmt.Errorf("no Cloudflare project detected")
	}
	if targetID == "" {
		return det.Targets[0], nil
	}
	for _, t := range det.Targets {
		if t.ID == targetID {
			return t, nil
		}
	}
	return provider.Target{}, fmt.Errorf("unknown target %q", targetID)
}

func (p *Provider) Configure(ctx context.Context, root string, opts provider.ConfigureOptions) (provider.ConfigureResult, error) {
	target, err := p.target(root, opts.TargetID)
	if err != nil {
		return provider.ConfigureResult{OK: false, Message: err.Error()}, nil
	}

	path := wranglerPath(root, target.Path)
	cfg, err := readD1Config(path)
	if err != nil {
		return provider.ConfigureResult{}, err
	}

	if !needsD1Link(cfg.DatabaseID) {
		return provider.ConfigureResult{OK: true, Message: "Cloudflare project already configured"}, nil
	}

	relPath := filepath.ToSlash(filepath.Join(target.Path, "wrangler.toml"))
	if opts.DryRun {
		return provider.ConfigureResult{
			OK:      true,
			Message: fmt.Sprintf("Would create D1 database %q and update %s", cfg.DatabaseName, relPath),
			Changed: []string{relPath},
		}, nil
	}

	who, err := p.WhoAmI(ctx)
	if err != nil {
		return provider.ConfigureResult{}, err
	}
	if !who.LoggedIn {
		return provider.ConfigureResult{
			OK:      false,
			Message: "Not logged in to Cloudflare",
			Hints: []provider.Hint{{
				Code:    "auth.required",
				Message: who.Message,
				Action:  "orbit login cloudflare",
			}},
		}, nil
	}

	workDir := filepath.Join(root, target.Path)
	out, err := run.Capture(ctx, "wrangler", []string{"d1", "create", cfg.DatabaseName}, workDir, p.cmdEnv()...)
	if err != nil {
		return provider.ConfigureResult{
			OK:      false,
			Message: "D1 create failed",
			Hints: []provider.Hint{{
				Code:    "cloudflare.d1.create_failed",
				Message: err.Error(),
			}},
		}, nil
	}

	databaseID, err := parseDatabaseIDFromCreateOutput(out)
	if err != nil {
		return provider.ConfigureResult{}, err
	}
	if err := patchDatabaseID(path, databaseID); err != nil {
		return provider.ConfigureResult{}, err
	}

	return provider.ConfigureResult{
		OK:      true,
		Message: fmt.Sprintf("Linked D1 database %s", cfg.DatabaseName),
		Changed: []string{relPath},
	}, nil
}

func (p *Provider) Deploy(ctx context.Context, root string, opts provider.DeployOptions) (provider.DeployResult, error) {
	_ = ctx
	_ = root
	_ = opts
	return provider.DeployResult{OK: true, Message: "Use orbit deploy for phased wrangler execution"}, nil
}

func (p *Provider) Phases(root string, opts provider.DeployOptions) []run.Step {
	target, err := p.target(root, opts.TargetID)
	if err != nil {
		return nil
	}
	workDir := filepath.Join(root, target.Path)

	cfg, _ := readD1Config(wranglerPath(root, target.Path))
	migrateArgs := []string{"d1", "migrations", "apply", cfg.DatabaseName, "--remote"}
	if opts.Environment == "preview" {
		migrateArgs = []string{"d1", "migrations", "apply", cfg.DatabaseName, "--local"}
	}

	steps := []run.Step{
		{
			ID:    "whoami",
			Title: "Verify Cloudflare authentication",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "wrangler", Args: []string{"whoami"}, Dir: workDir, Env: p.cmdEnv()})
			},
		},
	}

	if cfg.DatabaseName != "" {
		steps = append(steps, run.Step{
			ID:    "migrate",
			Title: "Apply D1 migrations",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "wrangler", Args: migrateArgs, Dir: workDir, Env: p.cmdEnv()})
			},
		})
	}

	steps = append(steps, run.Step{
		ID:    "deploy",
		Title: "Deploy Worker",
		Run: func(ctx context.Context, log *run.StepLogger) error {
			return run.RunCommand(ctx, log, run.CmdOptions{Name: "wrangler", Args: []string{"deploy"}, Dir: workDir, Env: p.cmdEnv()})
		},
	})

	return steps
}

func (p *Provider) Login(ctx context.Context) (provider.LoginResult, error) {
	if _, err := run.LookPath("wrangler"); err != nil {
		return provider.LoginResult{OK: false, Message: "wrangler not found on PATH — install: npm i -g wrangler"}, nil
	}
	err := run.RunInteractive(ctx, "wrangler", []string{"login"}, "")
	if err != nil {
		return provider.LoginResult{OK: false, Message: err.Error()}, nil
	}
	who, _ := p.WhoAmI(ctx)
	return provider.LoginResult{OK: who.LoggedIn, Account: who.Account, Message: who.Message}, nil
}

func (p *Provider) WhoAmI(ctx context.Context) (provider.WhoAmIResult, error) {
	if _, err := run.LookPath("wrangler"); err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: "wrangler not installed"}, nil
	}
	out, err := run.Capture(ctx, "wrangler", []string{"whoami"}, "", p.cmdEnv()...)
	if err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: err.Error()}, nil
	}
	return provider.WhoAmIResult{LoggedIn: true, Account: strings.TrimSpace(out), Message: "authenticated"}, nil
}

func (p *Provider) Doctor(ctx context.Context) ([]provider.Check, error) {
	_ = ctx
	checks := []provider.Check{
		{Name: "wrangler CLI", OK: false, Message: "not found", Fix: "npm i -g wrangler"},
	}
	if _, err := run.LookPath("wrangler"); err == nil {
		checks[0].OK = true
		checks[0].Message = "found on PATH"
		checks[0].Fix = ""
	}
	if credentials.Has(ID) {
		checks = append(checks, provider.Check{
			Name:    "API token",
			OK:      true,
			Message: "stored in OS keychain",
		})
	}
	return checks, nil
}

func (p *Provider) cmdEnv() []string {
	env, _ := credentials.Env(ID)
	return env
}

func init() {
	provider.Register(New())
	provider.RegisterAuthGuide(provider.AuthGuide{
		ProviderID:  ID,
		TokenLabel:  "Cloudflare API token",
		CreateURL:   "https://dash.cloudflare.com/profile/api-tokens",
		DocsURL:     "https://developers.cloudflare.com/fundamentals/api/get-started/create-token/",
		Permissions: "Workers Scripts (edit), Workers KV (edit), D1 (edit), Account Settings (read)",
		OAuthSteps: []string{
			"Orbit runs wrangler login — your browser opens to Cloudflare",
			"Sign in to Cloudflare if you are not already",
			"On the Wrangler screen, review permissions and click Authorize",
			"Return to this terminal — Orbit verifies with wrangler whoami",
		},
		Steps: []string{
			"Click \"Create Token\" on the API Tokens page",
			"Use the \"Edit Cloudflare Workers\" template, or create a custom token with Workers + D1 permissions",
			"Copy the token — it is shown only once",
		},
	})
}
