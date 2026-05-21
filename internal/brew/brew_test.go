package brew_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RajarshiParmar/devup/internal/brew"
	"github.com/RajarshiParmar/devup/internal/runner"
)

func TestUpdateAndUpgrade_Success(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("brew", "/opt/homebrew/bin/brew")
	// brew update and brew upgrade both succeed by default

	err := brew.UpdateAndUpgrade(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := m.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls (update + upgrade), got %d", len(calls))
	}
	if calls[0].Name != "brew" || calls[0].Args[0] != "update" {
		t.Errorf("expected brew update, got %v %v", calls[0].Name, calls[0].Args)
	}
	if calls[1].Name != "brew" || calls[1].Args[0] != "upgrade" {
		t.Errorf("expected brew upgrade, got %v %v", calls[1].Name, calls[1].Args)
	}
}

func TestUpdateAndUpgrade_BrewNotFound(t *testing.T) {
	m := runner.NewMock()
	// brew not in PATH

	err := brew.UpdateAndUpgrade(context.Background(), m)
	if err == nil {
		t.Fatal("expected error when brew not in PATH")
	}
}

func TestUpdateAndUpgrade_UpdateFails(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("brew", "/opt/homebrew/bin/brew")
	m.SetRunError(errors.New("network error"), "brew", "update")

	err := brew.UpdateAndUpgrade(context.Background(), m)
	if err == nil {
		t.Fatal("expected error when brew update fails")
	}
}

func TestUpdateAndUpgrade_UpgradeFails(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("brew", "/opt/homebrew/bin/brew")
	m.SetRunError(errors.New("formula conflict"), "brew", "upgrade")

	err := brew.UpdateAndUpgrade(context.Background(), m)
	if err == nil {
		t.Fatal("expected error when brew upgrade fails")
	}
}

func TestIsInstalled_Found(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("brew", "/opt/homebrew/bin/brew")
	if !brew.IsInstalled(m) {
		t.Error("expected IsInstalled to return true")
	}
}

func TestIsInstalled_NotFound(t *testing.T) {
	m := runner.NewMock()
	if brew.IsInstalled(m) {
		t.Error("expected IsInstalled to return false when brew not in PATH")
	}
}
