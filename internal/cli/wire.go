package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dendro-X0/Orbit/internal/providers/vercel"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func stackContains(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func wireAPIURLStep(session *run.Session, root string, st state.Project, deployEnv string) run.Step {
	return run.Step{
		ID:    "wire-vite-api-url",
		Title: "[orbit] Wire VITE_API_URL to Vercel",
		Run: func(ctx context.Context, log *run.StepLogger) error {
			logPath := filepath.Join(session.RunDir, "combined.log")
			b, err := os.ReadFile(logPath)
			if err != nil {
				return fmt.Errorf("read deploy log: %w", err)
			}

			apiURL := run.ExtractWorkersURL(string(b))
			if apiURL == "" {
				return fmt.Errorf("no Workers URL found in deploy logs")
			}
			session.APIURL = apiURL
			log.Stdout(fmt.Sprintf("Found API URL: %s", apiURL))

			workDir, err := vercel.WorkDir(root, st.TargetFor("vercel"))
			if err != nil {
				return err
			}

			return vercel.SetEnvVar(ctx, workDir, "VITE_API_URL", apiURL, vercelEnvName(deployEnv), log)
		},
	}
}

func newWireCmd() *cobra.Command {
	var apiURL string
	var env string

	cmd := &cobra.Command{
		Use:   "wire",
		Short: "Wire API URL into Vercel environment variables",
		Long:  "Sets VITE_API_URL on Vercel from the last Workers deploy URL or --api-url.",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}

			st, _ := state.Load(statePath(root))
			if !st.IsConfigured("vercel") {
				return fmt.Errorf("vercel not configured — run: orbit configure --provider vercel")
			}

			url := apiURL
			if url == "" {
				runDir, err := run.LatestRunDir(root)
				if err != nil {
					return fmt.Errorf("no deploy runs found — pass --api-url")
				}
				b, err := os.ReadFile(filepath.Join(runDir, "combined.log"))
				if err != nil {
					return err
				}
				url = run.ExtractWorkersURL(string(b))
			}
			if url == "" {
				return fmt.Errorf("could not determine API URL — pass --api-url")
			}

			workDir, err := vercel.WorkDir(root, st.TargetFor("vercel"))
			if err != nil {
				return err
			}

			wireEnv := env
			if wireEnv == "production" && st.Environment != "" {
				wireEnv = st.Environment
			}
			vercelEnv := vercelEnvName(wireEnv)

			fmt.Printf("Wiring VITE_API_URL=%s (%s)\n", url, vercelEnv)
			if err := vercel.SetEnvVar(cmd.Context(), workDir, "VITE_API_URL", url, vercelEnv, nil); err != nil {
				return err
			}
			fmt.Println("✓ VITE_API_URL set on Vercel")
			return nil
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "", "Workers API URL (default: parse from last deploy log)")
	cmd.Flags().StringVar(&env, "env", "production", "Vercel environment (production or preview)")
	return cmd
}

func vercelEnvName(deployEnv string) string {
	if deployEnv == "preview" {
		return "preview"
	}
	return "production"
}
