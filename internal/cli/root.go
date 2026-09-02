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
		RunE:  runDefault,
	}

	cmd.PersistentFlags().StringVar(&rootDir, "path", "", "project root (default: auto-detect)")

	cmd.AddCommand(newShipCmd())
	cmd.AddCommand(newMenuCmd())
	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newWhoamiCmd())
	cmd.AddCommand(newConfigureCmd())
	cmd.AddCommand(newDeployCmd())
	cmd.AddCommand(newRetryCmd())
	cmd.AddCommand(newWireCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newOpenCmd())
	cmd.AddCommand(newSecretsCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newLogsCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func runDefault(cmd *cobra.Command, _ []string) error {
	if isInteractive() {
		return runShipWorkflow(cmd)
	}
	printNonInteractiveHelp()
	return nil
}

func runMenuAction(cmd *cobra.Command, action string) error {
	switch action {
	case "ship":
		return runShipWorkflow(cmd)
	case "deploy":
		return newDeployCmd().RunE(cmd, nil)
	case "configure":
		return newConfigureCmd().RunE(cmd, nil)
	case "login":
		return newLoginCmd().RunE(cmd, nil)
	case "doctor":
		return newDoctorCmd().RunE(cmd, nil)
	case "status":
		return newStatusCmd().RunE(cmd, nil)
	case "secrets":
		return newSecretsCmd().RunE(cmd, nil)
	case "logs":
		return newLogsCmd().RunE(cmd, nil)
	default:
		return fmt.Errorf("unknown action %q", action)
	}
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
