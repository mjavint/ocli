package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type Odoo struct {
	odooBin    string
	configPath string
}

// NewStartOdooCmd creates the start command.
func NewStartOdooCmd(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Odoo Start Server",
		Long:  `Start the Odoo server with specified addons`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgVals := deps.cfg()
			odooCfg := &Odoo{
				odooBin:    cfgVals.Odoo.OdooBin,
				configPath: cfgVals.Odoo.ConfigFile,
			}
			return odooCfg.startOdooServer(cmd.Context(), deps)
		},
	}
	setCmdIO(cmd, deps)
	return cmd
}

func (cfg *Odoo) startOdooServer(parent context.Context, deps Deps) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(signalChan)

	cmdExec := exec.CommandContext(ctx, cfg.odooBin, "-c", cfg.configPath)
	cmdExec.Stdout = deps.stdout()
	cmdExec.Stderr = deps.stderr()
	cmdExec.Stdin = os.Stdin

	if err := cmdExec.Start(); err != nil {
		return fmt.Errorf("failed to start Odoo: %w", err)
	}

	fmt.Fprintf(deps.stdout(), "✅ Odoo started with PID %d\n", cmdExec.Process.Pid)
	fmt.Fprintln(deps.stdout(), "Press Ctrl+C to stop...")

	errChan := make(chan error, 1)
	go func() {
		errChan <- cmdExec.Wait()
	}()

	select {
	case sig := <-signalChan:
		return cfg.handleShutdown(sig, cmdExec, errChan, cancel, deps)
	case err := <-errChan:
		return cfg.handleCompletion(err, deps)
	}
}

func (cfg *Odoo) handleShutdown(sig os.Signal, cmd *exec.Cmd, errChan <-chan error, cancel context.CancelFunc, deps Deps) error {
	fmt.Fprintf(deps.stdout(), "\n🛑 Received signal %v. Shutting down Odoo...\n", sig)

	cancel()

	const shutdownTimeout = 10 * time.Second
	select {
	case err := <-errChan:
		return cfg.logShutdownResult(err, deps)
	case <-time.After(shutdownTimeout):
		fmt.Fprintln(deps.stdout(), "⚠️ Shutdown timeout exceeded, forcing termination...")
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill process: %w", err)
			}
		}
		return <-errChan
	}
}

func (cfg *Odoo) handleCompletion(err error, deps Deps) error {
	if err == nil {
		fmt.Fprintln(deps.stdout(), "✅ Odoo finished successfully")
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			fmt.Fprintf(deps.stdout(), "🔴 Odoo exited with code %d\n", status.ExitStatus())
		}
	}
	return fmt.Errorf("Odoo execution failed: %w", err)
}

func (cfg *Odoo) logShutdownResult(err error, deps Deps) error {
	if err != nil {
		fmt.Fprintf(deps.stdout(), "⚠️ Odoo terminated with error: %v\n", err)
		return err
	}
	fmt.Fprintln(deps.stdout(), "✅ Odoo shutdown gracefully")
	return nil
}
