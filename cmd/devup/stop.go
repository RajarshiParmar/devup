package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/RajarshiParmar/devup/internal/docker"
	"github.com/RajarshiParmar/devup/internal/orbstack"
	"github.com/RajarshiParmar/devup/internal/runner"
	"github.com/RajarshiParmar/devup/internal/ui"
)

func newStopCmd(f *globalFlags) *cobra.Command {
	var all bool
	var project bool

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop running services",
		Long: `Interactively select what to stop, or use flags to skip prompts:
  --all       Stop project containers and OrbStack
  --project   Stop project containers only`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStop(cmd.Context(), f, all, project)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "stop everything (containers + OrbStack)")
	cmd.Flags().BoolVar(&project, "project", false, "stop project containers only")

	return cmd
}

func runStop(ctx context.Context, f *globalFlags, all, project bool) error {
	u := ui.New(f.noColor)
	r := runner.New()
	dir := viper.GetString("dir")

	u.Header("devup stop")

	// Determine what to stop.
	stopContainers := false
	stopOrb := false

	switch {
	case all:
		stopContainers = true
		stopOrb = true
	case project:
		stopContainers = true
	default:
		// Interactive: ask the user what to stop.
		var choices []string
		options := []huh.Option[string]{
			huh.NewOption("Project containers (docker compose down)", "containers"),
		}
		if orbstack.IsRunning(ctx, r) {
			options = append(options, huh.NewOption("OrbStack", "orbstack"))
		}

		err := huh.NewMultiSelect[string]().
			Title("What would you like to stop?").
			Options(options...).
			Value(&choices).
			Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if len(choices) == 0 {
			u.Warn("nothing selected — no changes made")
			return nil
		}

		for _, c := range choices {
			switch c {
			case "containers":
				stopContainers = true
			case "orbstack":
				stopOrb = true
			}
		}
	}

	// Execute stop actions.
	if stopContainers {
		_, err := u.SpinStep("Stopping project containers", func() error {
			return docker.StopContainers(ctx, r, dir)
		})
		if err != nil {
			return err
		}
	}

	if stopOrb {
		_, err := u.SpinStep("Stopping OrbStack", func() error {
			return orbstack.Stop(ctx, r)
		})
		if err != nil {
			return err
		}
	}

	fmt.Println()
	u.Success("Done")
	return nil
}
