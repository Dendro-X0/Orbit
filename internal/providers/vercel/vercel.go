package vercel

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
)

const ID = "vercel"

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) ID() string          { return ID }
func (p *Provider) DisplayName() string { return "Vercel" }

func (p *Provider) Detect(ctx context.Context, root string) (provider.Detection, error) {
	_ = ctx
	targets, err := findVercelTargets(root)
	if err != nil {
		return provider.Detection{}, err
	}
	if len(targets) == 0 {
		return provider.Detection{Supported: false, Summary: "No vercel.json found"}, nil
	}

	out := make([]provider.Target, 0, len(targets))
	for _, t := range targets {
		out = append(out, provider.Target{ID: t.ID, Path: t.Path, Kind: t.Kind})
	}

	return provider.Detection{
		Supported: true,
		Summary:   fmt.Sprintf("Vercel project (%d target(s))", len(out)),
		Targets:   out,
	}, nil
}

func (p *Provider) target(root, targetID string) (target, error) {
	det, err := p.Detect(context.Background(), root)
	if err != nil {
		return target{}, err
	}
	if !det.Supported {
		return target{}, fmt.Errorf("no Vercel project detected")
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

	link, linked, err := readProjectLink(root, t.Path)
	if err != nil {
		return provider.ConfigureResult{}, err
	}

	required := requiredViteEnvVars(root, t.Path)
	missing := []string{}

	if linked && len(required) > 0 {
		envOut, err := run.Capture(ctx, "vercel", []string{"env", "ls"}, workDir)
		if err == nil {
			present := parseEnvList(envOut)
			for _, key := range required {
				if !present[key] {
					missing = append(missing, key)
				}
			}
		}
	}

	if opts.DryRun {
		msg := "Would link Vercel project"
		if linked {
			msg = fmt.Sprintf("Vercel project already linked (%s)", link.ProjectID)
		}
		if len(missing) > 0 {
			msg += fmt.Sprintf("; would prompt for env: %s", strings.Join(missing, ", "))
		}
		return provider.ConfigureResult{
			OK:      true,
			Message: msg,
			Changed: []string{filepath.ToSlash(filepath.Join(t.Path, ".vercel"))},
		}, nil
	}

	if linked && len(missing) == 0 {
		who, err := p.WhoAmI(ctx)
		if err != nil {
			return provider.ConfigureResult{}, err
		}
		if !who.LoggedIn {
			return provider.ConfigureResult{
				OK:      false,
				Message: "Not logged in to Vercel",
				Hints: []provider.Hint{{
					Code: "auth.required", Message: who.Message, Action: "orbit login vercel",
				}},
			}, nil
		}
		return provider.ConfigureResult{
			OK:      true,
			Message: fmt.Sprintf("Vercel project already linked (%s)", link.ProjectID),
		}, nil
	}

	who, err := p.WhoAmI(ctx)
	if err != nil {
		return provider.ConfigureResult{}, err
	}
	if !who.LoggedIn {
		return provider.ConfigureResult{
			OK:      false,
			Message: "Not logged in to Vercel",
			Hints: []provider.Hint{{
				Code:    "auth.required",
				Message: who.Message,
				Action:  "orbit login vercel",
			}},
		}, nil
	}

	changed := []string{}
	if !linked {
		if err := run.RunInteractive(ctx, "vercel", []string{"link", "--yes"}, workDir); err != nil {
			return provider.ConfigureResult{
				OK:      false,
				Message: "Vercel link failed",
				Hints: []provider.Hint{{
					Code:    "vercel.link_failed",
					Message: err.Error(),
					Action:  fmt.Sprintf("cd %s && vercel link", t.Path),
				}},
			}, nil
		}
		changed = append(changed, filepath.ToSlash(filepath.Join(t.Path, ".vercel", "project.json")))
	}

	var hints []provider.Hint
	if len(missing) > 0 {
		hints = append(hints, provider.Hint{
			Code:    "vercel.env.missing",
			Message: fmt.Sprintf("Missing env vars: %s", strings.Join(missing, ", ")),
			Action:  fmt.Sprintf("cd %s && vercel env add <name> production", t.Path),
		})
	}

	msg := "Vercel project configured"
	if !linked {
		msg = "Linked Vercel project"
	}
	if len(missing) > 0 {
		msg += " (some env vars still missing)"
	}

	return provider.ConfigureResult{
		OK:      true,
		Message: msg,
		Changed: changed,
		Hints:   hints,
	}, nil
}

func (p *Provider) Deploy(ctx context.Context, root string, opts provider.DeployOptions) (provider.DeployResult, error) {
	_ = ctx
	_ = root
	_ = opts
	return provider.DeployResult{OK: true, Message: "Use orbit deploy for phased vercel execution"}, nil
}

func (p *Provider) Phases(root string, opts provider.DeployOptions) []run.Step {
	t, err := p.target(root, opts.TargetID)
	if err != nil {
		return nil
	}
	workDir := filepath.Join(root, t.Path)

	deployArgs := []string{"deploy", "--yes"}
	title := "Deploy preview"
	if opts.Environment == "production" {
		deployArgs = append(deployArgs, "--prod")
		title = "Deploy production"
	}

	return []run.Step{
		{
			ID:    "whoami",
			Title: "Verify Vercel authentication",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "vercel", Args: []string{"whoami"}, Dir: workDir})
			},
		},
		{
			ID:    "deploy",
			Title: title,
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "vercel", Args: deployArgs, Dir: workDir})
			},
		},
	}
}

func (p *Provider) Login(ctx context.Context) (provider.LoginResult, error) {
	if _, err := run.LookPath("vercel"); err != nil {
		return provider.LoginResult{OK: false, Message: "vercel not found on PATH — install: npm i -g vercel"}, nil
	}
	if err := run.RunInteractive(ctx, "vercel", []string{"login"}, ""); err != nil {
		return provider.LoginResult{OK: false, Message: err.Error()}, nil
	}
	who, _ := p.WhoAmI(ctx)
	return provider.LoginResult{OK: who.LoggedIn, Account: who.Account, Message: who.Message}, nil
}

func (p *Provider) WhoAmI(ctx context.Context) (provider.WhoAmIResult, error) {
	if _, err := run.LookPath("vercel"); err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: "vercel not installed"}, nil
	}
	out, err := run.Capture(ctx, "vercel", []string{"whoami"}, "")
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
		{Name: "vercel CLI", OK: false, Message: "not found", Fix: "npm i -g vercel"},
	}
	if _, err := run.LookPath("vercel"); err == nil {
		checks[0].OK = true
		checks[0].Message = "found on PATH"
		checks[0].Fix = ""
	}
	return checks, nil
}

func init() {
	provider.Register(New())
}
