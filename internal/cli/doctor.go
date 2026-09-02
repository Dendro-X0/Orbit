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
			printTitle("orbit doctor")
			fmt.Println()

			ok := true
			if _, err := run.LookPath("git"); err != nil {
				fmt.Printf("%s %s\n", failMark(), ui.dim.Render("git not found (optional)"))
			} else {
				printSuccess("git")
			}

			for _, p := range provider.All() {
				checks, err := p.Doctor(cmd.Context())
				if err != nil {
					return err
				}
				fmt.Printf("\n%s\n", styledProvider(p.DisplayName())+ui.dim.Render(":"))
				for _, c := range checks {
					if c.OK {
						fmt.Printf("  %s %s %s %s\n", okMark(), ui.label.Render(c.Name), ui.dim.Render("—"), ui.value.Render(c.Message))
					} else {
						ok = false
						fmt.Printf("  %s %s %s %s\n", failMark(), ui.label.Render(c.Name), ui.dim.Render("—"), ui.error.Render(c.Message))
					}
					if !c.OK && c.Fix != "" {
						fmt.Printf("    %s %s\n", ui.label.Render("fix:"), highlightCmdLine(c.Fix))
					}
				}
				if authCheck, aok := providerAuthCheck(cmd.Context(), p); authCheck != nil {
					if aok {
						fmt.Printf("  %s %s %s %s\n", okMark(), ui.label.Render(authCheck.Name), ui.dim.Render("—"), ui.value.Render(authCheck.Message))
					} else {
						ok = false
						fmt.Printf("  %s %s %s %s\n", failMark(), ui.label.Render(authCheck.Name), ui.dim.Render("—"), ui.error.Render(authCheck.Message))
					}
					if !aok && authCheck.Fix != "" {
						fmt.Printf("    %s %s\n", ui.label.Render("fix:"), highlightCmdLine(authCheck.Fix))
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
						fmt.Printf("\n%s %s\n", ui.warn.Render("Secrets:"), ui.warn.Render(msg))
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
