package odoo

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

// RunOptions carries the stdio wiring for a child process.
type RunOptions struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Runner executes odoo-bin commands.
type Runner interface {
	Run(ctx context.Context, bin string, args []string, opts RunOptions) error
	CombinedOutput(ctx context.Context, bin string, args []string) ([]byte, error)
}

// ExecRunner runs commands via os/exec with context support.
type ExecRunner struct{}

// NewExecRunner creates an ExecRunner.
func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

// Run executes the command wiring the provided stdio.
func (r *ExecRunner) Run(ctx context.Context, bin string, args []string, opts RunOptions) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	}
	cmd.Env = helperEnv(args)
	cmd.WaitDelay = 5 * time.Second
	return cmd.Run()
}

// CombinedOutput executes the command and returns its combined output.
func (r *ExecRunner) CombinedOutput(ctx context.Context, bin string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = helperEnv(args)
	cmd.WaitDelay = 5 * time.Second
	return cmd.CombinedOutput()
}

func helperEnv(args []string) []string {
	for _, a := range args {
		if a == "-test.run=TestHelperProcess" {
			return append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		}
	}
	return nil
}
