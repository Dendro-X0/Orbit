package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var failedOnly bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View logs from the last deploy run",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			runDir, err := run.LatestRunDir(root)
			if err != nil {
				return fmt.Errorf("no runs yet — deploy first")
			}

			logPath := filepath.Join(runDir, "combined.log")
			if failedOnly {
				failurePath := filepath.Join(runDir, "failure.json")
				if _, err := os.Stat(failurePath); err == nil {
					b, _ := os.ReadFile(failurePath)
					fmt.Println(string(b))
					return nil
				}
			}

			b, err := os.ReadFile(logPath)
			if err != nil {
				return err
			}
			fmt.Print(string(b))
			return nil
		},
	}

	cmd.Flags().BoolVar(&failedOnly, "failed", false, "show failure.json instead of combined log")
	return cmd
}
