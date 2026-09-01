package cli

import (
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/run"
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
			}

			root, _ := projectRoot(cmd)
			if root != "" {
				printToolkitHints(root)
			}

			if !ok {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		},
	}
}
