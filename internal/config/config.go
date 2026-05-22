// Package config provides helpers for reading and writing the devup
// configuration file at ~/.config/devup/config.yaml.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config represents the devup configuration file schema.
type Config struct {
	Dir           string `yaml:"dir"`
	Timeout       int    `yaml:"timeout"`
	Upgrade       bool   `yaml:"upgrade"`
	Verbose       bool   `yaml:"verbose"`
	DockerContext string `yaml:"docker_context"`
	Notify        bool   `yaml:"notify"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	cwd, _ := os.Getwd()
	return Config{
		Dir:           cwd,
		Timeout:       60,
		Upgrade:       true,
		Verbose:       false,
		DockerContext: "orbstack",
		Notify:        true,
	}
}

// Path returns the default config file path: ~/.config/devup/config.yaml.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".config", "devup", "config.yaml"), nil
}

// Exists returns true if the config file already exists at the default path.
func Exists() (bool, error) {
	p, err := Path()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// Write serializes the config to YAML and writes it to the default path.
// It creates parent directories if they don't exist.
func Write(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := marshalConfig(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(p, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Read loads and parses the config from the default path.
func Read() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file: %w", err)
	}

	return parseConfig(data)
}

// marshalConfig serializes a Config to YAML bytes.
func marshalConfig(cfg Config) ([]byte, error) {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}
	return data, nil
}

// parseConfig deserializes YAML bytes into a Config.
func parseConfig(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config file: %w", err)
	}
	return cfg, nil
}
