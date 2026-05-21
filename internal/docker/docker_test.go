package docker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RajarshiParmar/devup/internal/docker"
	"github.com/RajarshiParmar/devup/internal/runner"
)

func TestCurrentContext(t *testing.T) {
	m := runner.NewMock()
	m.SetOutput([]byte("orbstack\n"), nil, "docker", "context", "show")

	ctx, err := docker.CurrentContext(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx != "orbstack" {
		t.Errorf("expected orbstack, got %q", ctx)
	}
}

func TestCurrentContext_Error(t *testing.T) {
	m := runner.NewMock()
	m.SetOutput(nil, errors.New("docker not found"), "docker", "context", "show")

	_, err := docker.CurrentContext(context.Background(), m)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSwitchContext_AlreadySet(t *testing.T) {
	m := runner.NewMock()
	m.SetOutput([]byte("orbstack\n"), nil, "docker", "context", "show")

	err := docker.SwitchContext(context.Background(), m, "orbstack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// docker context use should NOT have been called.
	for _, c := range m.Calls() {
		if c.Name == "docker" && len(c.Args) > 1 && c.Args[0] == "context" && c.Args[1] == "use" {
			t.Error("docker context use was called even though context was already orbstack")
		}
	}
}

func TestSwitchContext_Switches(t *testing.T) {
	m := runner.NewMock()
	m.SetOutput([]byte("default\n"), nil, "docker", "context", "show")

	err := docker.SwitchContext(context.Background(), m, "orbstack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// docker context use orbstack should have been called.
	found := false
	for _, c := range m.Calls() {
		if c.Name == "docker" && len(c.Args) == 3 && c.Args[1] == "use" && c.Args[2] == "orbstack" {
			found = true
		}
	}
	if !found {
		t.Error("expected docker context use orbstack to be called")
	}
}

func TestIsReady_True(t *testing.T) {
	m := runner.NewMock()
	// no error configured → Run returns nil → IsReady = true
	if !docker.IsReady(context.Background(), m) {
		t.Error("expected IsReady to be true")
	}
}

func TestIsReady_False(t *testing.T) {
	m := runner.NewMock()
	m.SetRunError(errors.New("not running"), "docker", "info")
	if docker.IsReady(context.Background(), m) {
		t.Error("expected IsReady to be false")
	}
}

func TestWaitReady_ImmediatelyReady(t *testing.T) {
	m := runner.NewMock()
	// docker info succeeds on first call
	err := docker.WaitReady(context.Background(), m, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitReady_Timeout(t *testing.T) {
	m := runner.NewMock()
	m.SetRunError(errors.New("not running"), "docker", "info")

	err := docker.WaitReady(context.Background(), m, 1*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWaitReady_ContextCancelled(t *testing.T) {
	m := runner.NewMock()
	m.SetRunError(errors.New("not running"), "docker", "info")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := docker.WaitReady(ctx, m, 30*time.Second, nil)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}
