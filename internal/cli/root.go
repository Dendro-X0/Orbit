package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootDir      string
	providerFlag string
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orbit",
		Short: "Deploy to your cloud — login, configure, deploy",
		RunE:  runMenu,
	}

	cmd.PersistentFlags().StringVar(&rootDir, "path", "", "project root (default: auto-detect)")

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newWhoamiCmd())
	cmd.AddCommand(newConfigureCmd())
	cmd.AddCommand(newDeployCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newLogsCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func runMenu(cmd *cobra.Command, _ []string) error {
	fmt.Println("orbit — deploy to your cloud, simply.")
	fmt.Println()
	fmt.Println("  [1] Deploy this project     orbit deploy")
	fmt.Println("  [2] Log in to a provider    orbit login")
	fmt.Println("  [3] Check setup             orbit doctor")
	fmt.Println("  [4] View last run logs      orbit logs")
	fmt.Println()
	fmt.Println("Run `orbit <command> --help` for details.")
	return nil
}

func projectRoot(cmd *cobra.Command) (string, error) {
	if rootDir != "" {
		return rootDir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	_ = cmd
	return findRoot(cwd)
}
