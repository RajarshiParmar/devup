// Package docker provides helpers for managing Docker context and
// waiting for the Docker engine to become ready.
package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/RajarshiParmar/devup/internal/runner"
)

const (
	pollInterval = 2 * time.Second
)

// CurrentContext returns the name of the active Docker context.
func CurrentContext(ctx context.Context, r runner.Runner) (string, error) {
	out, err := r.Output(ctx, "docker", "context", "show")
	if err != nil {
		return "", fmt.Errorf("docker context show: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SwitchContext switches the active Docker context to target.
// It is a no-op when the current context already matches.
func SwitchContext(ctx context.Context, r runner.Runner, target string) error {
	current, err := CurrentContext(ctx, r)
	if err != nil {
		// docker may not be reachable yet; log and continue
		slog.Debug("could not read docker context", "err", err)
		return nil
	}

	if current == target {
		slog.Debug("docker context already set", "context", target)
		return nil
	}

	slog.Debug("switching docker context", "from", current, "to", target)
	if err := r.Run(ctx, "docker", "context", "use", target); err != nil {
		return fmt.Errorf("docker context use %s: %w", target, err)
	}
	return nil
}

// IsReady reports whether the Docker engine is accepting connections.
func IsReady(ctx context.Context, r runner.Runner) bool {
	err := r.Run(ctx, "docker", "info")
	return err == nil
}

// RunningContainers returns the names of running containers.
// If dir is non-empty, it filters to containers from a compose project in that directory.
func RunningContainers(ctx context.Context, r runner.Runner, dir string) ([]string, error) {
	args := []string{"ps", "--format", "{{.Names}}"}
	if dir != "" {
		args = []string{"compose", "-f", dir + "/docker-compose.yml", "ps", "--format", "{{.Names}}", "--status", "running"}
	}

	out, err := r.Output(ctx, "docker", args...)
	if err != nil {
		// If compose file doesn't exist, fall back to all containers.
		if dir != "" {
			out, err = r.Output(ctx, "docker", "ps", "--format", "{{.Names}}")
			if err != nil {
				return nil, fmt.Errorf("docker ps: %w", err)
			}
		} else {
			return nil, fmt.Errorf("docker ps: %w", err)
		}
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

// WaitReady polls the Docker engine every 2 s until it is ready or
// timeout is exceeded. It returns an error if the deadline is reached
// or the context is cancelled.
func WaitReady(ctx context.Context, r runner.Runner, timeout time.Duration, progress func(elapsed time.Duration)) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	start := time.Now()
	for {
		if IsReady(ctx, r) {
			return nil
		}

		elapsed := time.Since(start)
		if progress != nil {
			progress(elapsed)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("docker engine did not become ready within %s", timeout)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled while waiting for docker engine: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
