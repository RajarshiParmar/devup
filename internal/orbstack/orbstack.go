// Package orbstack provides helpers for starting the OrbStack engine.
package orbstack

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/RajarshiParmar/devup/internal/runner"
)

// Start attempts to start the OrbStack Docker engine.
//
// Strategy:
//  1. If the orb CLI is available, run `orb start`.
//  2. If `orb start` fails (or orb is not in PATH), fall back to
//     `open -a OrbStack` to launch the GUI app.
func Start(ctx context.Context, r runner.Runner) error {
	if path, err := r.LookPath("orb"); err == nil {
		slog.Debug("found orb CLI", "path", path)
		if err := r.Run(ctx, "orb", "start"); err == nil {
			slog.Debug("orb start succeeded")
			return nil
		} else {
			slog.Debug("orb start failed, falling back to GUI", "err", err)
		}
	} else {
		slog.Debug("orb CLI not found, trying GUI fallback")
	}

	// Fallback: open the OrbStack.app — best effort.
	if err := r.Run(ctx, "open", "-a", "OrbStack"); err != nil {
		return fmt.Errorf("could not start OrbStack via CLI or GUI: %w", err)
	}

	return nil
}

// IsInstalled reports whether OrbStack is installed (either the CLI or
// the macOS application bundle).
func IsInstalled(r runner.Runner) bool {
	if _, err := r.LookPath("orb"); err == nil {
		return true
	}
	// Check for the app bundle via mdfind (Spotlight)
	if _, err := r.LookPath("open"); err == nil {
		return true // open -a will simply fail gracefully if the app is missing
	}
	return false
}

// IsRunning reports whether OrbStack is currently running.
func IsRunning(ctx context.Context, r runner.Runner) bool {
	err := r.Run(ctx, "orb", "status")
	return err == nil
}
