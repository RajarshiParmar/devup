package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Timeout != 60 {
		t.Errorf("expected default timeout 60, got %d", cfg.Timeout)
	}
	if cfg.DockerContext != "orbstack" {
		t.Errorf("expected default docker_context 'orbstack', got %q", cfg.DockerContext)
	}
	if !cfg.Upgrade {
		t.Error("expected default upgrade to be true")
	}
	if !cfg.Notify {
		t.Error("expected default notify to be true")
	}
}

func TestWriteAndRead(t *testing.T) {
	// Use a temp dir to avoid writing to real config path.
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.yaml")

	// Override the Path function behavior by writing directly.
	cfg := Config{
		Dir:           "/tmp/test-project",
		Timeout:       30,
		Upgrade:       false,
		Verbose:       true,
		DockerContext: "default",
		Notify:        true,
	}

	// Write manually to temp location.
	data, err := marshalConfig(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Read it back.
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	got, err := parseConfig(readData)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got.Dir != cfg.Dir {
		t.Errorf("Dir: got %q, want %q", got.Dir, cfg.Dir)
	}
	if got.Timeout != cfg.Timeout {
		t.Errorf("Timeout: got %d, want %d", got.Timeout, cfg.Timeout)
	}
	if got.DockerContext != cfg.DockerContext {
		t.Errorf("DockerContext: got %q, want %q", got.DockerContext, cfg.DockerContext)
	}
}
