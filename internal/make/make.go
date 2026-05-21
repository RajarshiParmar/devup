// Package make provides helpers for running make targets inside a
// project directory.
package make

import (
	"context"
	"fmt"
	"os"

	"github.com/RajarshiParmar/devup/internal/runner"
)

// Start changes to dir and runs `make start`, streaming output to the
// terminal. It returns a descriptive error if make is not found, the
// directory does not exist, or make exits with a non-zero status.
func Start(ctx context.Context, r runner.Runner, dir string) error {
	if _, err := r.LookPath("make"); err != nil {
		return fmt.Errorf("make not found in PATH: %w", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("project directory %q not found: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path %q is not a directory", dir)
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cannot change to directory %q: %w", dir, err)
	}

	if err := r.Run(ctx, "make", "start"); err != nil {
		return fmt.Errorf("make start failed in %q: %w", dir, err)
	}

	return nil
}
