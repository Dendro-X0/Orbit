package cli

import (
	"fmt"

	"github.com/Dendro-X0/Orbit/internal/project"
	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/spf13/cobra"
)

func newRetryCmd() *cobra.Command {
	var fromStep string
	var env string

	cmd := &cobra.Command{
		Use:   "retry",
		Short: "Retry deploy from the last failed step",
		Long:  "Reads the last run's failure.json and resumes deploy from the failed step.",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := projectRoot(cmd)
			if err != nil {
				return err
			}
			if err := project.EnsureOrbitDir(root); err != nil {
				return err
			}

			runDir, err := run.LatestRunDir(root)
			if err != nil {
				return fmt.Errorf("no deploy runs yet — run: orbit deploy")
			}

			stepID := fromStep
			var manifest *run.Manifest
			if stepID == "" {
				failure, err := run.LoadFailure(runDir)
				if err != nil {
					if summary, serr := run.LoadSummary(runDir); serr == nil && summary.OK {
						return fmt.Errorf("last deploy succeeded — use: orbit deploy")
					}
					return fmt.Errorf("no failed step found — run: orbit deploy")
				}
				stepID = failure.FailedStep
			}

			manifest, err = run.LoadManifest(runDir)
			if err != nil {
				return fmt.Errorf("read manifest: %w", err)
			}

			st, _ := state.Load(statePath(root))
			ids := parseProviderIDs(manifest.Provider)
			if len(ids) == 0 {
				ids, err = resolveDeployProviders(cmd.Context(), root, st, providerFlag)
				if err != nil {
					return err
				}
			}

			deployEnv := env
			if deployEnv == "production" {
				if st.Environment != "" {
					deployEnv = st.Environment
				} else {
					deployEnv = "production"
				}
			}

			session := run.Session{}
			allSteps, err := buildDeploySteps(root, st, ids, deployEnv, &session)
			if err != nil {
				return err
			}

			steps, err := sliceStepsFrom(allSteps, stepID)
			if err != nil {
				return err
			}

			label := providerListLabel(ids)
			fmt.Printf("orbit retry — %s from %s (%s)\n\n", label, stepID, deployEnv)

			r := &run.Runner{}
			result, err := r.Execute(cmd.Context(), run.Options{
				Root:      root,
				Provider:  label,
				Command:   "deploy",
				PrintLive: true,
				Session:   &session,
			}, steps)

			return printDeployResult(cmd.Context(), root, st, label, result, err)
		},
	}

	cmd.Flags().StringVar(&fromStep, "from-step", "", "step ID to resume from (default: last failure)")
	cmd.Flags().StringVar(&env, "env", "production", "target environment")
	cmd.Flags().StringVar(&providerFlag, "provider", "", "retry a single provider deploy")
	return cmd
}
