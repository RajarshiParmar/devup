// Package brew provides helpers for running Homebrew update and upgrade.
package brew

import (
	"context"
	"fmt"

	"github.com/RajarshiParmar/devup/internal/runner"
)

// UpdateAndUpgrade runs `brew update` followed by `brew upgrade`.
// Both commands stream their output to the terminal.
func UpdateAndUpgrade(ctx context.Context, r runner.Runner) error {
	if _, err := r.LookPath("brew"); err != nil {
		return fmt.Errorf("brew not found in PATH: %w", err)
	}

	if err := r.Run(ctx, "brew", "update"); err != nil {
		return fmt.Errorf("brew update: %w", err)
	}

	if err := r.Run(ctx, "brew", "upgrade"); err != nil {
		return fmt.Errorf("brew upgrade: %w", err)
	}

	return nil
}

// IsInstalled reports whether Homebrew is available in PATH.
func IsInstalled(r runner.Runner) bool {
	_, err := r.LookPath("brew")
	return err == nil
}
