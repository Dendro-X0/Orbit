package cloudflare

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (p *Provider) Configure(ctx context.Context, root string, opts provider.ConfigureOptions) (provider.ConfigureResult, error) {
	_ = ctx
	det, err := p.Detect(ctx, root)
	if err != nil {
		return provider.ConfigureResult{}, err
	}
	if !det.Supported {
		return provider.ConfigureResult{OK: false, Message: "No Cloudflare project detected"}, nil
	}

	target := det.Targets[0]
	wranglerPath := filepath.Join(root, target.Path, "wrangler.toml")
	content, err := os.ReadFile(wranglerPath)
	if err != nil {
		return provider.ConfigureResult{}, err
	}

	if strings.Contains(string(content), "REPLACE_AFTER_CREATE") {
		if opts.DryRun {
			return provider.ConfigureResult{
				OK:      true,
				Message: "Would create D1 database and update wrangler.toml",
				Changed: []string{filepath.ToSlash(filepath.Join(target.Path, "wrangler.toml"))},
			}, nil
		}
		return provider.ConfigureResult{
			OK:      true,
			Message: "D1 database not linked yet",
			Hints: []provider.Hint{{
				Code:    "cloudflare.d1.unlinked",
				Message: "database_id is still REPLACE_AFTER_CREATE",
				Action:  "wrangler d1 create <name> then set database_id in wrangler.toml",
			}},
		}, nil
	}

	return provider.ConfigureResult{OK: true, Message: "Cloudflare project already configured"}, nil
}

func (p *Provider) Deploy(ctx context.Context, root string, opts provider.DeployOptions) (provider.DeployResult, error) {
	_ = ctx
	_ = root
	_ = opts
	return provider.DeployResult{OK: true, Message: "Use orbit deploy for phased wrangler execution"}, nil
}

func (p *Provider) Phases(root string, opts provider.DeployOptions) []run.Step {
	det, _ := p.Detect(context.Background(), root)
	targetPath := "."
	if len(det.Targets) > 0 {
		targetPath = det.Targets[0].Path
	}
	if opts.TargetID != "" {
		for _, t := range det.Targets {
			if t.ID == opts.TargetID {
				targetPath = t.Path
				break
			}
		}
	}
	workDir := filepath.Join(root, targetPath)

	return []run.Step{
		{
			ID:    "whoami",
			Title: "Verify Cloudflare authentication",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "wrangler", Args: []string{"whoami"}, Dir: workDir})
			},
		},
		{
			ID:    "deploy",
			Title: "Deploy Worker",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "wrangler", Args: []string{"deploy"}, Dir: workDir})
			},
		},
	}
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
	out, err := run.Capture(ctx, "wrangler", []string{"whoami"}, "")
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
	return checks, nil
}

func init() {
	provider.Register(New())
}
