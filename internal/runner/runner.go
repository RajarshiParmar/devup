// Package runner provides an abstraction over os/exec to make all
// command execution testable via dependency injection.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Runner is the interface that wraps command execution.
// All internal packages accept a Runner so they can be unit-tested
// without spawning real processes.
type Runner interface {
	// Run executes name with args, streaming stdout/stderr to the
	// process's own stdout/stderr. It returns a non-nil error when
	// the process exits with a non-zero status.
	Run(ctx context.Context, name string, args ...string) error

	// Output executes name with args and returns the combined
	// stdout+stderr output.
	Output(ctx context.Context, name string, args ...string) ([]byte, error)

	// LookPath reports whether an executable named file can be found
	// in the directories named by the PATH environment variable.
	LookPath(file string) (string, error)
}

// Real is the production Runner that delegates to os/exec.
type Real struct{}

// New returns a new production Runner.
func New() Runner {
	return &Real{}
}

// Run implements Runner.
func (r *Real) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", name, err)
	}
	return nil
}

// Output implements Runner.
func (r *Real) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.Bytes(), fmt.Errorf("run %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// LookPath implements Runner.
func (r *Real) LookPath(file string) (string, error) {
	path, err := exec.LookPath(file)
	if err != nil {
		return "", fmt.Errorf("look path %s: %w", file, err)
	}
	return path, nil
}
