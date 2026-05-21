package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RajarshiParmar/devup/internal/runner"
)

func TestMockRunner_Run_NoError(t *testing.T) {
	m := runner.NewMock()
	err := m.Run(context.Background(), "docker", "info")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMockRunner_Run_ConfiguredError(t *testing.T) {
	m := runner.NewMock()
	want := errors.New("exit status 1")
	m.SetRunError(want, "docker", "info")

	err := m.Run(context.Background(), "docker", "info")
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestMockRunner_Output(t *testing.T) {
	m := runner.NewMock()
	m.SetOutput([]byte("orbstack\n"), nil, "docker", "context", "show")

	out, err := m.Output(context.Background(), "docker", "context", "show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "orbstack\n" {
		t.Fatalf("expected %q, got %q", "orbstack\n", string(out))
	}
}

func TestMockRunner_LookPath_Found(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("docker", "/usr/local/bin/docker")

	path, err := m.LookPath("docker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/usr/local/bin/docker" {
		t.Fatalf("expected /usr/local/bin/docker, got %q", path)
	}
}

func TestMockRunner_LookPath_NotFound(t *testing.T) {
	m := runner.NewMock()
	_, err := m.LookPath("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
}

func TestMockRunner_RecordsCalls(t *testing.T) {
	m := runner.NewMock()
	_ = m.Run(context.Background(), "docker", "info")
	_, _ = m.Output(context.Background(), "docker", "context", "show")

	calls := m.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Name != "docker" {
		t.Errorf("call[0].Name = %q, want docker", calls[0].Name)
	}
}
