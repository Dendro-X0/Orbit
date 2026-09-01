package fly

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
)

const ID = "fly"

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) ID() string          { return ID }
func (p *Provider) DisplayName() string { return "Fly.io" }

func (p *Provider) Detect(ctx context.Context, root string) (provider.Detection, error) {
	_ = ctx
	targets, err := findFlyTargets(root)
	if err != nil {
		return provider.Detection{}, err
	}
	if len(targets) == 0 {
		return provider.Detection{Supported: false, Summary: "No fly.toml found"}, nil
	}

	out := make([]provider.Target, 0, len(targets))
	for _, t := range targets {
		out = append(out, provider.Target{ID: t.ID, Path: t.Path, Kind: t.Kind})
	}

	return provider.Detection{
		Supported: true,
		Summary:   fmt.Sprintf("Fly.io app (%d target(s))", len(out)),
		Targets:   out,
	}, nil
}

func (p *Provider) target(root, targetID string) (target, error) {
	det, err := p.Detect(context.Background(), root)
	if err != nil {
		return target{}, err
	}
	if !det.Supported {
		return target{}, fmt.Errorf("no Fly.io project detected")
	}
	if targetID == "" {
		t := det.Targets[0]
		return target{ID: t.ID, Path: t.Path, Kind: t.Kind}, nil
	}
	for _, t := range det.Targets {
		if t.ID == targetID {
			return target{ID: t.ID, Path: t.Path, Kind: t.Kind}, nil
		}
	}
	return target{}, fmt.Errorf("unknown target %q", targetID)
}

func (p *Provider) Configure(ctx context.Context, root string, opts provider.ConfigureOptions) (provider.ConfigureResult, error) {
	t, err := p.target(root, opts.TargetID)
	if err != nil {
		return provider.ConfigureResult{OK: false, Message: err.Error()}, nil
	}

	workDir := filepath.Join(root, t.Path)
	tomlPath := flyTomlPath(root, t.Path)
	appName, err := parseAppName(tomlPath)
	if err != nil {
		return provider.ConfigureResult{}, err
	}

	relToml := filepath.ToSlash(filepath.Join(t.Path, "fly.toml"))
	ready, statusMsg := p.appReady(ctx, workDir, appName)

	if opts.DryRun {
		msg := "Would link Fly.io app via fly launch --no-deploy"
		if ready {
			msg = statusMsg
		} else if appName != "" {
			msg = fmt.Sprintf("Would create or link Fly app %q", appName)
		}
		return provider.ConfigureResult{
			OK:      true,
			Message: msg,
			Changed: []string{relToml},
		}, nil
	}

	if ready {
		return provider.ConfigureResult{OK: true, Message: statusMsg}, nil
	}

	who, err := p.WhoAmI(ctx)
	if err != nil {
		return provider.ConfigureResult{}, err
	}
	if !who.LoggedIn {
		return provider.ConfigureResult{
			OK:      false,
			Message: "Not logged in to Fly.io",
			Hints: []provider.Hint{{
				Code:    "auth.required",
				Message: who.Message,
				Action:  "orbit login fly",
			}},
		}, nil
	}

	bin, err := p.flyBin()
	if err != nil {
		return provider.ConfigureResult{OK: false, Message: err.Error()}, nil
	}

	args := []string{"launch", "--no-deploy", "--yes", "--copy-config"}
	if appName != "" {
		args = append(args, "--name", appName)
	}
	if err := run.RunInteractive(ctx, bin, args, workDir, p.cmdEnv()...); err != nil {
		return provider.ConfigureResult{
			OK:      false,
			Message: "Fly launch failed",
			Hints: []provider.Hint{{
				Code:    "fly.launch_failed",
				Message: err.Error(),
				Action:  fmt.Sprintf("cd %s && fly launch --no-deploy", t.Path),
			}},
		}, nil
	}

	return provider.ConfigureResult{
		OK:      true,
		Message: "Linked Fly.io app",
		Changed: []string{relToml},
	}, nil
}

func (p *Provider) appReady(ctx context.Context, workDir, appName string) (bool, string) {
	bin, err := p.flyBin()
	if err != nil {
		return false, ""
	}
	args := []string{"status"}
	if appName != "" {
		args = append(args, "-a", appName)
	}
	if _, err := run.Capture(ctx, bin, args, workDir, p.cmdEnv()...); err != nil {
		return false, ""
	}
	if appName != "" {
		return true, fmt.Sprintf("Fly app %q is linked", appName)
	}
	return true, "Fly.io app is linked"
}

func (p *Provider) Deploy(ctx context.Context, root string, opts provider.DeployOptions) (provider.DeployResult, error) {
	_ = ctx
	_ = root
	_ = opts
	return provider.DeployResult{OK: true, Message: "Use orbit deploy for phased fly execution"}, nil
}

func (p *Provider) Phases(root string, opts provider.DeployOptions) []run.Step {
	t, err := p.target(root, opts.TargetID)
	if err != nil {
		return nil
	}
	workDir := filepath.Join(root, t.Path)
	tomlPath := flyTomlPath(root, t.Path)
	appName, _ := parseAppName(tomlPath)

	bin, err := p.flyBin()
	if err != nil {
		return nil
	}

	statusArgs := []string{"status"}
	deployArgs := []string{"deploy", "--yes"}
	if appName != "" {
		statusArgs = append(statusArgs, "-a", appName)
		deployArgs = append(deployArgs, "-a", appName)
	}

	title := "Deploy to Fly.io"
	if opts.Environment != "production" {
		title = "Deploy to Fly.io (preview)"
	}

	return []run.Step{
		{
			ID:    "whoami",
			Title: "Verify Fly.io authentication",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: bin, Args: []string{"auth", "whoami"}, Dir: workDir, Env: p.cmdEnv()})
			},
		},
		{
			ID:    "status",
			Title: "Verify Fly.io app",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: bin, Args: statusArgs, Dir: workDir, Env: p.cmdEnv()})
			},
		},
		{
			ID:    "deploy",
			Title: title,
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: bin, Args: deployArgs, Dir: workDir, Env: p.cmdEnv()})
			},
		},
	}
}

func (p *Provider) Login(ctx context.Context) (provider.LoginResult, error) {
	bin, err := p.flyBin()
	if err != nil {
		return provider.LoginResult{OK: false, Message: err.Error()}, nil
	}
	if err := run.RunInteractive(ctx, bin, []string{"auth", "login"}, ""); err != nil {
		return provider.LoginResult{OK: false, Message: err.Error()}, nil
	}
	who, _ := p.WhoAmI(ctx)
	return provider.LoginResult{OK: who.LoggedIn, Account: who.Account, Message: who.Message}, nil
}

func (p *Provider) WhoAmI(ctx context.Context) (provider.WhoAmIResult, error) {
	bin, err := p.flyBin()
	if err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: err.Error()}, nil
	}
	out, err := run.Capture(ctx, bin, []string{"auth", "whoami"}, "", p.cmdEnv()...)
	if err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: err.Error()}, nil
	}
	account := strings.TrimSpace(out)
	if account == "" {
		return provider.WhoAmIResult{LoggedIn: false, Message: "not authenticated"}, nil
	}
	return provider.WhoAmIResult{LoggedIn: true, Account: account, Message: "authenticated"}, nil
}

func (p *Provider) Doctor(ctx context.Context) ([]provider.Check, error) {
	_ = ctx
	checks := []provider.Check{
		{Name: "fly CLI", OK: false, Message: "not found", Fix: "https://fly.io/docs/hands-on/install-flyctl/"},
	}
	if _, err := p.flyBin(); err == nil {
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

func (p *Provider) flyBin() (string, error) {
	if _, err := run.LookPath("fly"); err == nil {
		return "fly", nil
	}
	if _, err := run.LookPath("flyctl"); err == nil {
		return "flyctl", nil
	}
	return "", fmt.Errorf("fly not found on PATH — install: https://fly.io/docs/hands-on/install-flyctl/")
}

func (p *Provider) cmdEnv() []string {
	env, _ := credentials.Env(ID)
	return env
}

// CmdEnv returns environment variables for fly subprocesses.
func CmdEnv() []string {
	return (&Provider{}).cmdEnv()
}

func init() {
	provider.Register(New())
	provider.RegisterAuthGuide(provider.AuthGuide{
		ProviderID:  ID,
		TokenLabel:  "Fly.io access token",
		CreateURL:   "https://fly.io/user/personal_access_tokens",
		DocsURL:     "https://fly.io/docs/reference/api/#authentication",
		Permissions: "Deploy apps in your Fly.io organization",
		OAuthSteps: []string{
			"Orbit runs fly auth login — your browser opens to Fly.io",
			"Sign in or create a Fly.io account",
			"Approve access when prompted",
			"Return to this terminal — Orbit verifies with fly auth whoami",
		},
		Steps: []string{
			"Create a new token on the Fly.io tokens page",
			"Copy the token — it is shown only once",
		},
	})
}
