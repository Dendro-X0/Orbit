package cli

import (
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check tool and provider health",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("orbit doctor")
			fmt.Println()

			ok := true
			if _, err := run.LookPath("git"); err != nil {
				fmt.Println("✗ git not found (optional)")
			} else {
				fmt.Println("✓ git")
			}

			for _, p := range provider.All() {
				checks, err := p.Doctor(cmd.Context())
				if err != nil {
					return err
				}
				fmt.Printf("\n%s:\n", p.DisplayName())
				for _, c := range checks {
					mark := "✗"
					if c.OK {
						mark = "✓"
					} else {
						ok = false
					}
					fmt.Printf("  %s %s — %s\n", mark, c.Name, c.Message)
					if !c.OK && c.Fix != "" {
						fmt.Printf("    fix: %s\n", c.Fix)
					}
				}
				if authCheck, aok := providerAuthCheck(cmd.Context(), p); authCheck != nil {
					mark := "✗"
					if aok {
						mark = "✓"
					} else {
						ok = false
					}
					fmt.Printf("  %s %s — %s\n", mark, authCheck.Name, authCheck.Message)
					if !aok && authCheck.Fix != "" {
						fmt.Printf("    fix: %s\n", authCheck.Fix)
					}
				}
			}

			root, _ := projectRoot(cmd)
			if root != "" {
				printToolkitHints(root)
				st, _ := state.Load(statePath(root))
				stack := detectStack(cmd.Context(), root)
				if len(stack) > 0 {
					if msg := cloudflareSecretsSummary(cmd.Context(), root, st); msg != "" {
						fmt.Printf("\nSecrets: %s\n", msg)
					}
				}
			}

			if !ok {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		},
	}
}
