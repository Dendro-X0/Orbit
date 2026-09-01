package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/credentials"
	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout [provider]",
		Short: "Remove stored API tokens from the OS keychain",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				removed := false
				for _, p := range provider.All() {
					if credentials.Has(p.ID()) {
						if err := credentials.Delete(p.ID()); err != nil {
							return err
						}
						fmt.Printf("✓ Removed %s token from keychain\n", p.DisplayName())
						removed = true
					}
				}
				if !removed {
					fmt.Println("No stored tokens found")
				}
				return nil
			}

			id := args[0]
			if _, err := provider.Get(id); err != nil {
				return err
			}
			if !credentials.Has(id) {
				fmt.Printf("No stored token for %s\n", id)
				return nil
			}
			if err := credentials.Delete(id); err != nil {
				return err
			}
			fmt.Printf("✓ Removed %s token from keychain\n", id)
			return nil
		},
	}
	return cmd
}

func readLoginToken(flag string) (string, error) {
	if flag == "" {
		return "", nil
	}
	if flag == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	return flag, nil
}
