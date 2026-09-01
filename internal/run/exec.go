package run

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type CmdOptions struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

// RunCommand executes an external process and streams output through the step logger.
func RunCommand(ctx context.Context, log *StepLogger, opts CmdOptions) error {
	name := opts.Name
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		if _, err := exec.LookPath(name + ".cmd"); err == nil {
			// Let exec resolve .cmd shims when present.
		}
	}

	cmd := exec.CommandContext(ctx, name, opts.Args...)
	cmd.Dir = opts.Dir
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	copyLines := func(r io.Reader, emit func(string)) {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			emit(sc.Text())
		}
	}

	done := make(chan struct{}, 2)
	go func() {
		copyLines(stdout, log.Stdout)
		done <- struct{}{}
	}()
	go func() {
		copyLines(stderr, log.Stderr)
		done <- struct{}{}
	}()
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(opts.Args, " "), err)
	}
	return nil
}

func LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FindFile(root string, parts ...string) string {
	return filepath.Join(append([]string{root}, parts...)...)
}

// RunInteractive runs a command attached to the user's terminal (for login flows).
func RunInteractive(ctx context.Context, name string, args []string, dir string, extraEnv ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// Capture runs a command and returns combined stdout.
func Capture(ctx context.Context, name string, args []string, dir string, extraEnv ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
