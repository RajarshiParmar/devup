package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/RajarshiParmar/devup/internal/brew"
	"github.com/RajarshiParmar/devup/internal/docker"
	devmake "github.com/RajarshiParmar/devup/internal/make"
	"github.com/RajarshiParmar/devup/internal/orbstack"
	"github.com/RajarshiParmar/devup/internal/runner"
	"github.com/RajarshiParmar/devup/internal/ui"
)

func newRestartCmd(f *globalFlags) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Stop and restart the development environment",
		Long: `Stops project containers (and optionally OrbStack), then runs the
full startup sequence again.

  --all   Also restart OrbStack (not just containers)`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRestart(cmd.Context(), f, all)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "also restart OrbStack")

	return cmd
}

func runRestart(ctx context.Context, f *globalFlags, all bool) error {
	u := ui.New(f.noColor)
	r := runner.New()

	dir := viper.GetString("dir")
	timeout := time.Duration(viper.GetInt("timeout")) * time.Second
	doUpgrade := viper.GetBool("upgrade")

	u.Header("devup restart")

	totalStart := time.Now()

	// ── Stop phase ──────────────────────────────────────────────────────────

	_, err := u.SpinStep("Stopping project containers", func() error {
		return docker.StopContainers(ctx, r, dir)
	})
	if err != nil {
		return err
	}

	if all {
		_, err := u.SpinStep("Stopping OrbStack", func() error {
			return orbstack.Stop(ctx, r)
		})
		if err != nil {
			return err
		}
	}

	// ── Start phase ─────────────────────────────────────────────────────────

	fmt.Println()
	u.Step("Restarting…")
	fmt.Println()

	// Optional brew upgrade.
	if doUpgrade {
		_, err := u.SpinStep("Updating Homebrew", func() error {
			return brew.UpdateAndUpgrade(ctx, r)
		})
		if err != nil {
			return err
		}
	}

	// Switch Docker context.
	_, _ = u.SpinStep("Switching Docker context", func() error {
		return docker.SwitchContext(ctx, r, "orbstack")
	})

	// Ensure Docker engine is running.
	if !docker.IsReady(ctx, r) {
		_, err := u.SpinStep("Starting OrbStack", func() error {
			return orbstack.Start(ctx, r)
		})
		if err != nil {
			return err
		}

		_, err = u.SpinStep(fmt.Sprintf("Waiting for Docker engine (timeout: %s)", timeout), func() error {
			return docker.WaitReady(ctx, r, timeout, func(_ time.Duration) {})
		})
		if err != nil {
			return err
		}
	}
	u.Success("Docker engine ready")

	// Run make start.
	_, err = u.SpinStep(fmt.Sprintf("Running make start in %s", dir), func() error {
		return devmake.Start(ctx, r, dir)
	})
	if err != nil {
		return err
	}

	// Summary.
	totalElapsed := time.Since(totalStart)
	fmt.Println()
	u.Success("Restarted in %s", formatDuration(totalElapsed))

	return nil
}
