package netlify

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
)

const ID = "netlify"

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) ID() string          { return ID }
func (p *Provider) DisplayName() string { return "Netlify" }

func (p *Provider) Detect(ctx context.Context, root string) (provider.Detection, error) {
	_ = ctx
	targets, err := findNetlifyTargets(root)
	if err != nil {
		return provider.Detection{}, err
	}
	if len(targets) == 0 {
		return provider.Detection{Supported: false, Summary: "No netlify.toml found"}, nil
	}

	out := make([]provider.Target, 0, len(targets))
	for _, t := range targets {
		out = append(out, provider.Target{ID: t.ID, Path: t.Path, Kind: t.Kind})
	}

	return provider.Detection{
		Supported: true,
		Summary:   fmt.Sprintf("Netlify site (%d target(s))", len(out)),
		Targets:   out,
	}, nil
}

func (p *Provider) target(root, targetID string) (target, error) {
	det, err := p.Detect(context.Background(), root)
	if err != nil {
		return target{}, err
	}
	if !det.Supported {
		return target{}, fmt.Errorf("no Netlify project detected")
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
	link, linked, err := readSiteLink(root, t.Path)
	if err != nil {
		return provider.ConfigureResult{}, err
	}

	required := requiredViteEnvVars(root, t.Path)
	missing := []string{}
	if linked && len(required) > 0 {
		envOut, err := run.Capture(ctx, "netlify", []string{"env:list"}, workDir, p.cmdEnv()...)
		if err == nil {
			present := parseEnvList(envOut)
			for _, key := range required {
				if !present[key] {
					missing = append(missing, key)
				}
			}
		}
	}

	relState := filepath.ToSlash(filepath.Join(t.Path, ".netlify", "state.json"))
	if opts.DryRun {
		msg := "Would link Netlify site"
		if linked {
			msg = fmt.Sprintf("Netlify site already linked (%s)", link.SiteID)
		}
		if len(missing) > 0 {
			msg += fmt.Sprintf("; would prompt for env: %s", strings.Join(missing, ", "))
		}
		return provider.ConfigureResult{
			OK:      true,
			Message: msg,
			Changed: []string{relState},
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
				Message: "Not logged in to Netlify",
				Hints: []provider.Hint{{
					Code: "auth.required", Message: who.Message, Action: "orbit login netlify",
				}},
			}, nil
		}
		return provider.ConfigureResult{
			OK:      true,
			Message: fmt.Sprintf("Netlify site already linked (%s)", link.SiteID),
		}, nil
	}

	who, err := p.WhoAmI(ctx)
	if err != nil {
		return provider.ConfigureResult{}, err
	}
	if !who.LoggedIn {
		return provider.ConfigureResult{
			OK:      false,
			Message: "Not logged in to Netlify",
			Hints: []provider.Hint{{
				Code:    "auth.required",
				Message: who.Message,
				Action:  "orbit login netlify",
			}},
		}, nil
	}

	changed := []string{}
	if !linked {
		if err := run.RunInteractive(ctx, "netlify", []string{"link"}, workDir, p.cmdEnv()...); err != nil {
			return provider.ConfigureResult{
				OK:      false,
				Message: "Netlify link failed",
				Hints: []provider.Hint{{
					Code:    "netlify.link_failed",
					Message: err.Error(),
					Action:  fmt.Sprintf("cd %s && netlify link", t.Path),
				}},
			}, nil
		}
		changed = append(changed, relState)
	}

	var hints []provider.Hint
	if len(missing) > 0 {
		hints = append(hints, provider.Hint{
			Code:    "netlify.env.missing",
			Message: fmt.Sprintf("Missing env vars: %s", strings.Join(missing, ", ")),
			Action:  fmt.Sprintf("cd %s && netlify env:set <name> <value>", t.Path),
		})
	}

	msg := "Netlify site configured"
	if !linked {
		msg = "Linked Netlify site"
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
	return provider.DeployResult{OK: true, Message: "Use orbit deploy for phased netlify execution"}, nil
}

func (p *Provider) Phases(root string, opts provider.DeployOptions) []run.Step {
	t, err := p.target(root, opts.TargetID)
	if err != nil {
		return nil
	}
	workDir := filepath.Join(root, t.Path)

	deployArgs := []string{"deploy"}
	title := "Deploy preview"
	if opts.Environment == "production" {
		deployArgs = append(deployArgs, "--prod")
		title = "Deploy production"
	}

	return []run.Step{
		{
			ID:    "whoami",
			Title: "Verify Netlify authentication",
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "netlify", Args: []string{"status"}, Dir: workDir, Env: p.cmdEnv()})
			},
		},
		{
			ID:    "deploy",
			Title: title,
			Run: func(ctx context.Context, log *run.StepLogger) error {
				return run.RunCommand(ctx, log, run.CmdOptions{Name: "netlify", Args: deployArgs, Dir: workDir, Env: p.cmdEnv()})
			},
		},
	}
}

func (p *Provider) Login(ctx context.Context) (provider.LoginResult, error) {
	if _, err := run.LookPath("netlify"); err != nil {
		return provider.LoginResult{OK: false, Message: "netlify not found on PATH — install: npm i -g netlify-cli"}, nil
	}
	if err := run.RunInteractive(ctx, "netlify", []string{"login"}, ""); err != nil {
		return provider.LoginResult{OK: false, Message: err.Error()}, nil
	}
	who, _ := p.WhoAmI(ctx)
	return provider.LoginResult{OK: who.LoggedIn, Account: who.Account, Message: who.Message}, nil
}

func (p *Provider) WhoAmI(ctx context.Context) (provider.WhoAmIResult, error) {
	if _, err := run.LookPath("netlify"); err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: "netlify not installed"}, nil
	}
	out, err := run.Capture(ctx, "netlify", []string{"status"}, "", p.cmdEnv()...)
	if err != nil {
		return provider.WhoAmIResult{LoggedIn: false, Message: err.Error()}, nil
	}
	account := parseLoggedInUser(out)
	if account == "" {
		return provider.WhoAmIResult{LoggedIn: false, Message: "not authenticated"}, nil
	}
	return provider.WhoAmIResult{LoggedIn: true, Account: account, Message: "authenticated"}, nil
}

func parseLoggedInUser(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Logged in as ") {
			rest := strings.TrimPrefix(line, "Logged in as ")
			if i := strings.Index(rest, " on "); i > 0 {
				return strings.TrimSpace(rest[:i])
			}
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func (p *Provider) Doctor(ctx context.Context) ([]provider.Check, error) {
	_ = ctx
	checks := []provider.Check{
		{Name: "netlify CLI", OK: false, Message: "not found", Fix: "npm i -g netlify-cli"},
	}
	if _, err := run.LookPath("netlify"); err == nil {
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

// CmdEnv returns environment variables for netlify subprocesses.
func CmdEnv() []string {
	return (&Provider{}).cmdEnv()
}

func init() {
	provider.Register(New())
	provider.RegisterAuthGuide(provider.AuthGuide{
		ProviderID:  ID,
		TokenLabel:  "Netlify personal access token",
		CreateURL:   "https://app.netlify.com/user/applications#personal-access-tokens",
		DocsURL:     "https://docs.netlify.com/cli/get-started/#authentication",
		Permissions: "Deploy and manage sites in your Netlify account",
		OAuthSteps: []string{
			"Orbit runs netlify login — your browser opens to Netlify",
			"Sign in or confirm your Netlify account",
			"Approve access when prompted",
			"Return to this terminal — Orbit verifies with netlify status",
		},
		Steps: []string{
			"Create a new personal access token",
			"Copy the token — it is shown only once",
		},
	})
}
