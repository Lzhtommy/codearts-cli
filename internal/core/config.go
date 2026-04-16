// Package core provides configuration storage for codearts-cli.
//
// Configuration is persisted as JSON at ~/.codearts-cli/config.json with
// file permissions 0600. Secrets (AK/SK) are stored in-place for now; a
// future revision can promote them to the OS keychain (see the lark-cli
// reference implementation in ../../../../cli).
package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Default endpoints for Huawei Cloud CodeArts. These can be overridden per-
// profile but they match what the web console uses for a cn-south-1 tenant.
const (
	DefaultProjectID = "cd130bd8357b4e7ab293a7979d1c8711"
	DefaultRegion    = "cn-south-1"

	configDirName  = ".codearts-cli"
	configFileName = "config.json"
	configFileMode = 0o600
	configDirMode  = 0o700
)

// Config is the on-disk configuration schema.
//
// Fields are intentionally simple and flat; the CodeArts API model has
// project_id / region scoping but most users operate against a single tenant
// so we avoid the multi-profile complexity of the lark-cli for now.
type Config struct {
	AK        string `json:"ak"`
	SK        string `json:"sk"`
	ProjectID string `json:"projectId"`
	Region    string `json:"region"`
	// UserID is the 32-char IAM user UUID of the caller. Optional overall
	// but required for write APIs that default assignee/author to the caller
	// (e.g. CreateIpdProjectIssue's `assignee` field).
	UserID string `json:"userId,omitempty"`
}

// Validate ensures required credentials are present. Errors include
// actionable hints so the user (or AI agent) knows what to do next.
func (c *Config) Validate() error {
	if c.AK == "" {
		return errors.New("ak is empty — run `codearts-cli config init` to set up credentials")
	}
	if c.SK == "" {
		return errors.New("sk is empty — run `codearts-cli config init` to set up credentials")
	}
	if c.Region == "" {
		return errors.New("region is empty — run `codearts-cli config set region cn-south-1`")
	}
	// ProjectID is no longer universally required (pipeline/repo commands
	// use --project-id flag instead). We only check AK/SK/Region here.
	return nil
}

// Path returns the absolute path of the config file.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, configDirName, configFileName), nil
}

// Load reads the config from disk. Returns a zero-value Config if the file
// does not exist (first-run).
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{ProjectID: DefaultProjectID, Region: DefaultRegion}, nil
		}
		return nil, fmt.Errorf("read config %s: %w", p, err)
	}
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", p, err)
	}
	// Backfill defaults for forward-compat.
	if cfg.ProjectID == "" {
		cfg.ProjectID = DefaultProjectID
	}
	if cfg.Region == "" {
		cfg.Region = DefaultRegion
	}
	return cfg, nil
}

// Save persists the config to disk with restrictive permissions.
func Save(cfg *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), configDirMode); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	// Write atomically: temp file in the same directory then rename.
	tmp, err := os.CreateTemp(filepath.Dir(p), ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Chmod(configFileMode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpName, p); err != nil {
		return fmt.Errorf("rename temp config: %w", err)
	}
	return nil
}

// Redacted returns a copy of the config with secrets masked, safe to print.
func Redacted(cfg *Config) *Config {
	c := *cfg
	c.AK = MaskLeft(c.AK, 4)
	c.SK = "****"
	return &c
}

func MaskLeft(s string, keep int) string {
	if len(s) <= keep {
		return "****"
	}
	return s[:keep] + "****"
}
