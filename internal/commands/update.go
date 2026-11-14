package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the update command.
func NewUpdateCmd(deps Deps) *cobra.Command {
	var (
		checkOnly bool
		force     bool
		targetVer string
		checksum  string
		timeout   time.Duration
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update ocli to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.Updater == nil {
				return fmt.Errorf("updater is not configured")
			}
			exe := ""
			if deps.Executable != nil {
				p, err := deps.Executable()
				if err != nil {
					return err
				}
				exe = p
			} else {
				p, err := os.Executable()
				if err != nil {
					return err
				}
				exe = p
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			opts := selfupdate.UpdateOptions{
				Current:        deps.Version,
				ExecutablePath: exe,
				TargetTag:      targetVer,
				PinnedSHA256:   checksum,
				CheckOnly:      checkOnly,
				Force:          force,
			}
			res, err := deps.Updater.Update(ctx, opts)
			out := cmd.OutOrStdout()
			if err != nil {
				if errors.Is(err, selfupdate.ErrUpdateAvailable) {
					fmt.Fprintf(out, "update available: %s -> %s\n", res.Current, res.Latest)
					return selfupdate.ErrUpdateAvailable
				}
				return err
			}
			if res.Updated {
				fmt.Fprintf(out, "updated ocli %s -> %s (sha256: %s)\n", res.Current, res.Latest, res.SHA256)
				return nil
			}
			if res.Available {
				fmt.Fprintf(out, "update available: %s -> %s\n", res.Current, res.Latest)
				return nil
			}
			fmt.Fprintf(out, "up to date: %s (sha256: %s)\n", res.Latest, res.SHA256)
			return nil
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates without applying them")
	cmd.Flags().BoolVar(&force, "force", false, "Force update even if already up to date")
	cmd.Flags().StringVar(&targetVer, "version", "", "Install a specific release tag")
	cmd.Flags().StringVar(&checksum, "checksum", "", "Expected SHA-256 checksum")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Total network timeout")
	setCmdIO(cmd, deps)
	return cmd
}
