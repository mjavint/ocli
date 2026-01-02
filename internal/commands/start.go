package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

type Odoo struct {
	odooBin    string
	configPath string
}

// initCmd represents the init command
func NewStartOdooCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Odoo Start Server",
		Long:  `Start the Odoo server with specified addons`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := &Odoo{
				odooBin:    config.AppConfig.Odoo.OdooBin,    // Ruta al binario de Odoo
				configPath: config.AppConfig.Odoo.ConfigFile, // Ruta al archivo de configuración
			}
			cfg.startOdooServer()
		},
	}
	return cmd
}

func (cfg *Odoo) startOdooServer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(signalChan)

	// Create and configure command
	cmd := exec.CommandContext(ctx, cfg.odooBin, "-c", cfg.configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Odoo: %w", err)
	}

	fmt.Printf("✅ Odoo started with PID %d\n", cmd.Process.Pid)
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for completion or signal
	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Wait()
	}()

	select {
	case sig := <-signalChan:
		return cfg.handleShutdown(sig, cmd, errChan, cancel)
	case err := <-errChan:
		return cfg.handleCompletion(err)
	}
}

func (cfg *Odoo) handleShutdown(sig os.Signal, cmd *exec.Cmd, errChan <-chan error, cancel context.CancelFunc) error {
	fmt.Printf("\n🛑 Received signal %v. Shutting down Odoo...\n", sig)

	cancel() // Gracefully terminate via context

	const shutdownTimeout = 10 * time.Second
	select {
	case err := <-errChan:
		return cfg.logShutdownResult(err)
	case <-time.After(shutdownTimeout):
		fmt.Println("⚠️ Shutdown timeout exceeded, forcing termination...")
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill process: %w", err)
			}
		}
		return <-errChan
	}
}

func (cfg *Odoo) handleCompletion(err error) error {
	if err == nil {
		fmt.Println("✅ Odoo finished successfully")
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			fmt.Printf("🔴 Odoo exited with code %d\n", status.ExitStatus())
		}
	}
	return fmt.Errorf("Odoo execution failed: %w", err)
}

func (cfg *Odoo) logShutdownResult(err error) error {
	if err != nil {
		fmt.Printf("⚠️ Odoo terminated with error: %v\n", err)
		return err
	}
	fmt.Println("✅ Odoo shutdown gracefully")
	return nil
}
