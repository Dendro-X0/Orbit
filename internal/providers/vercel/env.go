package vercel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/run"
)

// WorkDir resolves the filesystem path for a Vercel deploy target.
func WorkDir(root, targetID string) (string, error) {
	p := &Provider{}
	t, err := p.target(root, targetID)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, t.Path), nil
}

// SetEnvVar sets a Vercel environment variable (replaces if present).
func SetEnvVar(ctx context.Context, workDir, key, value, environment string, log *run.StepLogger) error {
	if _, err := run.LookPath("vercel"); err != nil {
		return fmt.Errorf("vercel not installed")
	}

	env, _ := credentials.Env(ID)
	_, _ = run.Capture(ctx, "vercel", []string{"env", "rm", key, environment, "--yes"}, workDir, env...)

	cmd := exec.CommandContext(ctx, "vercel", "env", "add", key, environment)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(value + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	if log != nil {
		log.Stdout(fmt.Sprintf("Setting %s=%s for %s", key, value, environment))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vercel env add %s: %w", key, err)
	}
	return nil
}
