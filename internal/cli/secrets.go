package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/providers/cloudflare"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newSecretsCmd() *cobra.Command {
	var providerID string
	var putName string

	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Check and set Cloudflare Worker secrets",
		Long:  "Reads secret names from wrangler.toml comments and compares with wrangler secret list.",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			id := providerID
			if id == "" {
				id = "cloudflare"
			}
			if id != "cloudflare" {
				return fmt.Errorf("secrets command currently supports cloudflare only")
			}

			st, _ := state.Load(statePath(root))
			p, err := provider.Get(id)
			if err != nil {
				return err
			}
			det, err := p.Detect(cmd.Context(), root)
			if err != nil {
				return err
			}
			if !det.Supported {
				return fmt.Errorf("no Cloudflare project detected")
			}

			targetID := st.TargetFor(id)
			if targetID == "" && len(det.Targets) > 0 {
				targetID = det.Targets[0].ID
			}
			targetPath := targetPathFor(det, targetID)
			if targetPath == "" {
				return fmt.Errorf("could not resolve Cloudflare target")
			}

			env := cloudflare.CmdEnv()

			if putName != "" {
				workDir := filepath.Join(root, targetPath)
				fmt.Printf("Setting secret %s (wrangler will prompt for value)…\n", putName)
				if err := cloudflare.PutSecret(cmd.Context(), workDir, putName, env); err != nil {
					return err
				}
				fmt.Printf("✓ %s set\n", putName)
				return printSecretsStatus(cmd, root, targetPath, env)
			}

			return printSecretsStatus(cmd, root, targetPath, env)
		},
	}

	cmd.Flags().StringVar(&providerID, "provider", "cloudflare", "provider (cloudflare)")
	cmd.Flags().StringVar(&putName, "put", "", "set a secret via wrangler secret put")
	return cmd
}

func printSecretsStatus(cmd *cobra.Command, root, targetPath string, env []string) error {
	required, set, missing, err := cloudflare.SecretStatus(cmd.Context(), root, targetPath, env)
	if err != nil {
		return err
	}

	fmt.Println("orbit secrets — Cloudflare Worker")
	fmt.Println()
	if len(required) == 0 {
		fmt.Println("No secrets documented in wrangler.toml comments.")
		fmt.Println("Add a comment block, e.g.:")
		fmt.Println("  # Secrets: API_KEY_PEPPER, GITHUB_TOKEN")
		return nil
	}

	fmt.Println("Documented secrets:")
	for _, name := range required {
		mark := "✗"
		status := "missing"
		for _, s := range set {
			if s == name {
				mark = "✓"
				status = "set"
				break
			}
		}
		fmt.Printf("  %s %-24s %s\n", mark, name, status)
	}

	if len(missing) == 0 {
		fmt.Println("\n✓ All documented secrets are set on the worker.")
		return nil
	}

	fmt.Println("\nMissing secrets:")
	for _, name := range missing {
		fmt.Printf("  orbit secrets --put %s\n", name)
	}
	fmt.Printf("\nOr manually: cd %s && wrangler secret put <NAME>\n", targetPath)
	return nil
}

func targetPathFor(det provider.Detection, targetID string) string {
	for _, t := range det.Targets {
		if t.ID == targetID {
			return t.Path
		}
	}
	if len(det.Targets) > 0 {
		return det.Targets[0].Path
	}
	return ""
}

func printSecretsReminder(ctx context.Context, root string, st state.Project, providerLabel string) {
	if !strings.Contains(providerLabel, "cloudflare") {
		return
	}
	p, err := provider.Get("cloudflare")
	if err != nil {
		return
	}
	det, err := p.Detect(ctx, root)
	if err != nil || !det.Supported {
		return
	}
	targetID := st.TargetFor("cloudflare")
	targetPath := targetPathFor(det, targetID)
	_, _, missing, err := cloudflare.SecretStatus(ctx, root, targetPath, cloudflare.CmdEnv())
	if err != nil || len(missing) == 0 {
		return
	}
	fmt.Printf("\n%s %s\n", ui.warn.Render("Secrets:"), ui.warn.Render(fmt.Sprintf("%d missing on worker — run: orbit secrets", len(missing))))
}

func cloudflareSecretsSummary(ctx context.Context, root string, st state.Project) string {
	if !stackContains(detectStack(ctx, root), "cloudflare") {
		return ""
	}
	p, err := provider.Get("cloudflare")
	if err != nil {
		return ""
	}
	det, err := p.Detect(ctx, root)
	if err != nil || !det.Supported {
		return ""
	}
	targetPath := targetPathFor(det, st.TargetFor("cloudflare"))
	_, _, missing, err := cloudflare.SecretStatus(ctx, root, targetPath, cloudflare.CmdEnv())
	if err != nil || len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("%d worker secret(s) missing → orbit secrets", len(missing))
}
