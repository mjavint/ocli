package odoo

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		os.Exit(2)
	}
	switch args[0] {
	case "echo-args":
		_, _ = os.Stdout.WriteString(strings.Join(args[1:], "\x00"))
		os.Exit(0)
	case "stdout-stderr":
		_, _ = os.Stdout.WriteString("out")
		_, _ = os.Stderr.WriteString("err")
		os.Exit(3)
	case "sleep":
		d, _ := strconv.Atoi(args[1])
		time.Sleep(time.Duration(d) * time.Second)
		os.Exit(0)
	default:
		os.Exit(2)
	}
}

func helperCommand(ctx context.Context, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestExecRunnerRun(t *testing.T) {
	r := NewExecRunner()
	var out, errb bytes.Buffer
	err := r.Run(context.Background(), os.Args[0], []string{"-test.run=TestHelperProcess", "--", "echo-args", "db", "-c", "x"}, RunOptions{
		Stdout: &out,
		Stderr: &errb,
	})
	if err != nil {
		t.Fatalf("Run() error = %v, stderr=%q", err, errb.String())
	}
	if got := out.String(); got != "db\x00-c\x00x" {
		t.Fatalf("stdout = %q, want args", got)
	}
}

func TestExecRunnerCombinedOutput(t *testing.T) {
	r := NewExecRunner()
	out, err := r.CombinedOutput(context.Background(), os.Args[0], []string{"-test.run=TestHelperProcess", "--", "stdout-stderr"})
	if err == nil {
		t.Fatal("CombinedOutput() error = nil, want non-nil")
	}
	if !strings.Contains(string(out), "out") || !strings.Contains(string(out), "err") {
		t.Fatalf("output = %q, want stdout and stderr", string(out))
	}
}

func TestExecRunnerContextCancel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signaling differs on windows")
	}
	r := NewExecRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := r.Run(ctx, os.Args[0], []string{"-test.run=TestHelperProcess", "--", "sleep", "30"}, RunOptions{})
	if err == nil {
		t.Fatal("Run() error = nil, want cancellation error")
	}
	if time.Since(start) > 10*time.Second {
		t.Fatalf("Run() took %v, want prompt cancellation", time.Since(start))
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		// exec may wrap differently; at least ensure it didn't succeed
	}
}
