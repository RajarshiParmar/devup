package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/RajarshiParmar/devup/internal/docker"
	"github.com/RajarshiParmar/devup/internal/orbstack"
	"github.com/RajarshiParmar/devup/internal/runner"
	"github.com/RajarshiParmar/devup/internal/ui"
)

func newStatusCmd(f *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the state of all managed services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd.Context(), f)
		},
	}
}

func runStatus(ctx context.Context, f *globalFlags) error {
	u := ui.New(f.noColor)
	r := runner.New()

	dir := viper.GetString("dir")

	u.Header("devup status")
	fmt.Println()

	// OrbStack status.
	if orbstack.IsInstalled(r) {
		if orbstack.IsRunning(ctx, r) {
			u.DoctorRow(true, "OrbStack", "running")
		} else {
			u.DoctorRow(false, "OrbStack", "stopped")
		}
	} else {
		u.DoctorWarnRow("OrbStack", "not installed")
	}

	// Docker status.
	if docker.IsReady(ctx, r) {
		u.DoctorRow(true, "Docker", "ready")

		// Docker context.
		currentCtx, err := docker.CurrentContext(ctx, r)
		if err == nil {
			u.DoctorRow(true, "Context", currentCtx)
		}

		// Running containers.
		containers, err := docker.RunningContainers(ctx, r, dir)
		if err == nil && len(containers) > 0 {
			summary := fmt.Sprintf("%d running: %s", len(containers), strings.Join(containers, ", "))
			u.DoctorRow(true, "Containers", summary)
		} else if err == nil {
			u.DoctorWarnRow("Containers", "none running")
		}
	} else {
		u.DoctorRow(false, "Docker", "not responding")
	}

	// Project directory.
	if dir != "" {
		u.DoctorRow(true, "Project dir", dir)
	} else {
		u.DoctorWarnRow("Project dir", "not configured (run 'devup init')")
	}

	fmt.Println()
	return nil
}
