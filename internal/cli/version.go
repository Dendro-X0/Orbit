package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("orbit " + Version)
			if Commit != "" {
				fmt.Print(" " + strings.TrimPrefix(Commit, "v"))
			}
			if BuildDate != "" {
				fmt.Print(" (" + BuildDate + ")")
			}
			fmt.Println()
		},
	}
}
