package orbstack_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RajarshiParmar/devup/internal/orbstack"
	"github.com/RajarshiParmar/devup/internal/runner"
)

func TestStart_OrbCLISucceeds(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("orb", "/usr/local/bin/orb")
	err := orbstack.Start(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range m.Calls() {
		if c.Name == "orb" && len(c.Args) == 1 && c.Args[0] == "start" {
			found = true
		}
	}
	if !found {
		t.Error("expected orb start to be called")
	}
}
func TestStart_OrbCLIFailsFallsBackToOpen(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("orb", "/usr/local/bin/orb")
	m.SetRunError(errors.New("orb start failed"), "orb", "start")
	err := orbstack.Start(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range m.Calls() {
		if c.Name == "open" && len(c.Args) == 2 && c.Args[0] == "-a" && c.Args[1] == "OrbStack" {
			found = true
		}
	}
	if !found {
		t.Error("expected open -a OrbStack fallback to be called")
	}
}
func TestStart_NoOrbFallsBackToOpen(t *testing.T) {
	m := runner.NewMock()
	err := orbstack.Start(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range m.Calls() {
		if c.Name == "open" {
			found = true
		}
	}
	if !found {
		t.Error("expected open to be called as fallback")
	}
}
func TestStart_BothFail(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("orb", "/usr/local/bin/orb")
	m.SetRunError(errors.New("orb failed"), "orb", "start")
	m.SetRunError(errors.New("open failed"), "open", "-a", "OrbStack")
	err := orbstack.Start(context.Background(), m)
	if err == nil {
		t.Fatal("expected error when both orb and open fail")
	}
}
func TestIsInstalled_WithOrb(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("orb", "/usr/local/bin/orb")
	if !orbstack.IsInstalled(m) {
		t.Error("expected IsInstalled to be true when orb is in PATH")
	}
}
