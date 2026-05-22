package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/RajarshiParmar/devup/internal/config"
	"github.com/RajarshiParmar/devup/internal/ui"
)

func newInitCmd(f *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard for devup configuration",
		Long: `Generates ~/.config/devup/config.yaml through an interactive prompt.
If a configuration file already exists, you will be asked to confirm overwrite.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInit(f)
		},
	}
}

func runInit(f *globalFlags) error {
	u := ui.New(f.noColor)
	u.Header("devup init")

	// Check if config already exists.
	exists, err := config.Exists()
	if err != nil {
		return fmt.Errorf("checking existing config: %w", err)
	}

	if exists {
		var overwrite bool
		err := huh.NewConfirm().
			Title("Configuration file already exists. Overwrite?").
			Value(&overwrite).
			Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		if !overwrite {
			u.Warn("init cancelled — existing config preserved")
			return nil
		}
	}

	// Start with defaults.
	cfg := config.DefaultConfig()

	// Interactive prompts.
	cwd, _ := os.Getwd()
	dirInput := cwd

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project directory").
				Description("Path to the project where `make start` runs").
				Placeholder(cwd).
				Value(&dirInput),

			huh.NewInput().
				Title("Docker context").
				Description("Docker context to use (usually 'orbstack')").
				Placeholder("orbstack").
				Value(&cfg.DockerContext),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Run brew upgrade on start?").
				Description("Automatically update Homebrew packages each time devup runs").
				Value(&cfg.Upgrade),

			huh.NewConfirm().
				Title("Enable desktop notifications?").
				Description("Show a notification when startup completes").
				Value(&cfg.Notify),
		),
	).Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	// Apply text inputs (handle empty = use default).
	if dirInput != "" {
		cfg.Dir = dirInput
	}
	if cfg.DockerContext == "" {
		cfg.DockerContext = "orbstack"
	}

	// Validate project directory exists.
	if info, statErr := os.Stat(cfg.Dir); statErr != nil || !info.IsDir() {
		u.Warn("directory %q does not exist — config will still be saved", cfg.Dir)
	}

	// Write config.
	if err := config.Write(cfg); err != nil {
		u.Error("failed to write config: %v", err)
		return err
	}

	configPath, _ := config.Path()
	fmt.Println()
	u.Success("Configuration saved to %s", configPath)
	fmt.Println()

	// Print summary.
	u.Step("dir:            %s", cfg.Dir)
	u.Step("timeout:        %ds", cfg.Timeout)
	u.Step("upgrade:        %v", cfg.Upgrade)
	u.Step("docker_context: %s", cfg.DockerContext)
	u.Step("notify:         %v", cfg.Notify)
	fmt.Println()

	u.Success("Run 'devup' to start your environment")
	return nil
}
