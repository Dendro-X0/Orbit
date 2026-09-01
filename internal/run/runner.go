package run

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Step struct {
	ID    string
	Title string
	Run   func(ctx context.Context, log *StepLogger) error
}

type Options struct {
	Root      string
	Provider  string
	Command   string
	RunID     string
	PrintLive bool
	Session   *Session
}

type Result struct {
	Manifest *Manifest
	Summary  *Summary
	Failure  *Failure
	RunDir   string
}

type Runner struct{}

func (r *Runner) Execute(ctx context.Context, opts Options, steps []Step) (*Result, error) {
	if opts.RunID == "" {
		opts.RunID = time.Now().UTC().Format("2006-01-02T15-04-05Z")
	}
	runDir := filepath.Join(opts.Root, ".orbit", "runs", opts.RunID)
	if opts.Session != nil {
		opts.Session.RunDir = runDir
	}
	if err := os.MkdirAll(filepath.Join(runDir, "steps"), 0o755); err != nil {
		return nil, err
	}

	combinedPath := filepath.Join(runDir, "combined.log")
	combined, err := os.Create(combinedPath)
	if err != nil {
		return nil, err
	}
	defer combined.Close()

	manifest := &Manifest{
		ID:        opts.RunID,
		StartedAt: time.Now().UTC(),
		Provider:  opts.Provider,
		Command:   opts.Command,
		Root:      opts.Root,
		Steps:     make([]StepRecord, 0, len(steps)),
	}

	started := time.Now()
	var failed *Failure

	for i, step := range steps {
		stepDir := filepath.Join(runDir, "steps")
		stdoutPath := filepath.Join(stepDir, fmt.Sprintf("%02d-%s.stdout.log", i+1, step.ID))
		stderrPath := filepath.Join(stepDir, fmt.Sprintf("%02d-%s.stderr.log", i+1, step.ID))

		stepLog := newStepLogger(stdoutPath, stderrPath, combined, opts.PrintLive)
		stepStarted := time.Now()

		writeCombined(combined, fmt.Sprintf("=== step %d/%d: %s (%s) ===\n", i+1, len(steps), step.Title, step.ID))

		err := step.Run(ctx, stepLog)
		stepEnded := time.Now()

		record := StepRecord{
			ID:        step.ID,
			Title:     step.Title,
			OK:        err == nil,
			StartedAt: stepStarted.UTC(),
			EndedAt:   stepEnded.UTC(),
			Duration:  stepEnded.Sub(stepStarted).Round(time.Millisecond).String(),
			StdoutLog: rel(opts.Root, stdoutPath),
			StderrLog: rel(opts.Root, stderrPath),
		}
		if err != nil {
			record.Error = err.Error()
			record.Hint = classifyError(err)
			manifest.OK = false
			manifest.Steps = append(manifest.Steps, record)
			manifest.EndedAt = time.Now().UTC()

			tail, _ := tailFile(stderrPath, 20)
			failed = &Failure{
				OK:          false,
				Final:       true,
				Command:     opts.Command,
				Provider:    opts.Provider,
				FailedStep:  step.ID,
				Message:     err.Error(),
				Hint:        record.Hint,
				ProviderRaw: tail,
				RunDir:      rel(opts.Root, runDir),
				LogPaths: Paths{
					Combined: rel(opts.Root, combinedPath),
					Stdout:   record.StdoutLog,
					Stderr:   record.StderrLog,
				},
			}
			break
		}
		manifest.Steps = append(manifest.Steps, record)
	}

	if failed == nil {
		manifest.OK = true
		manifest.EndedAt = time.Now().UTC()
	}

	if err := writeJSON(filepath.Join(runDir, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	_ = writeLatest(opts.Root, runDir)

	duration := time.Since(started).Round(time.Millisecond).String()
	result := &Result{Manifest: manifest, RunDir: runDir}

	if failed != nil {
		if err := writeJSON(filepath.Join(runDir, "failure.json"), failed); err != nil {
			return nil, err
		}
		result.Failure = failed
		return result, fmt.Errorf("%s", failed.Message)
	}

	summary := &Summary{
		OK:       true,
		Final:    true,
		Command:  opts.Command,
		Provider: opts.Provider,
		RunDir:   rel(opts.Root, runDir),
		Duration: duration,
	}
	if opts.Session != nil && opts.Session.APIURL != "" {
		summary.URL = opts.Session.APIURL
	}
	if err := writeJSON(filepath.Join(runDir, "summary.json"), summary); err != nil {
		return nil, err
	}
	result.Summary = summary
	return result, nil
}

type StepLogger struct {
	stdout    *os.File
	stderr    *os.File
	combined  *os.File
	printLive bool
}

// NewStepLogger creates a logger for a run step (used by providers in tests).
func NewStepLogger(stdoutPath, stderrPath string, combined *os.File, printLive bool) *StepLogger {
	return newStepLogger(stdoutPath, stderrPath, combined, printLive)
}

func newStepLogger(stdoutPath, stderrPath string, combined *os.File, printLive bool) *StepLogger {
	stdout, _ := os.Create(stdoutPath)
	stderr, _ := os.Create(stderrPath)
	return &StepLogger{stdout: stdout, stderr: stderr, combined: combined, printLive: printLive}
}

func (l *StepLogger) Stdout(line string) {
	l.write(l.stdout, line, os.Stdout)
}

func (l *StepLogger) Stderr(line string) {
	l.write(l.stderr, line, os.Stderr)
}

func (l *StepLogger) write(file *os.File, line string, live io.Writer) {
	redacted := RedactLine(line)
	if !strings.HasSuffix(redacted, "\n") {
		redacted += "\n"
	}
	_, _ = file.WriteString(redacted)
	_, _ = l.combined.WriteString(redacted)
	if l.printLive && live != nil {
		_, _ = io.WriteString(live, redacted)
	}
}

func writeCombined(f *os.File, line string) {
	_, _ = f.WriteString(line)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func rel(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}

func tailFile(path string, lines int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		buf = append(buf, sc.Text())
		if len(buf) > lines {
			buf = buf[1:]
		}
	}
	return strings.Join(buf, "\n"), sc.Err()
}

func writeLatest(root, runDir string) error {
	latest := filepath.Join(root, ".orbit", "latest")
	return os.WriteFile(latest, []byte(filepath.ToSlash(rel(root, runDir))+"\n"), 0o644)
}

func LatestRunDir(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, ".orbit", "latest"))
	if err != nil {
		return "", err
	}
	return filepath.Join(root, strings.TrimSpace(string(b))), nil
}

func classifyError(err error) *Hint {
	msg := err.Error()
	switch {
	case strings.Contains(strings.ToLower(msg), "not logged in"), strings.Contains(strings.ToLower(msg), "authentication"):
		return &Hint{Code: "auth.required", Message: "Provider authentication required", Action: "Run: orbit login <provider>"}
	case strings.Contains(strings.ToLower(msg), "not found"), strings.Contains(strings.ToLower(msg), "enoent"):
		return &Hint{Code: "cli.missing", Message: "Required CLI not found on PATH", Action: "Install the provider CLI and re-run orbit doctor"}
	default:
		return &Hint{Code: "deploy.failed", Message: msg}
	}
}
