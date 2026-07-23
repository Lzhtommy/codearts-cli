package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/horizon0514/codearts-cli/internal/core"
)

func TestValidate_OK(t *testing.T) {
	cfg := &core.Config{AK: "ak", SK: "sk", Region: "cn-south-1", Gateway: "http://gw:8099"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestValidate_MissingAK(t *testing.T) {
	cfg := &core.Config{SK: "sk", Region: "cn-south-1", Gateway: "http://gw:8099"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("should fail with empty AK")
	}
	if !strings.Contains(err.Error(), "config init") {
		t.Errorf("error should hint at config init, got: %s", err)
	}
}

func TestValidate_MissingSK(t *testing.T) {
	cfg := &core.Config{AK: "ak", Region: "cn-south-1", Gateway: "http://gw:8099"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("should fail with empty SK")
	}
	if !strings.Contains(err.Error(), "config init") {
		t.Errorf("error should hint at config init, got: %s", err)
	}
}

func TestValidate_MissingRegion(t *testing.T) {
	cfg := &core.Config{AK: "ak", SK: "sk", Gateway: "http://gw:8099"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("should fail with empty Region")
	}
	if !strings.Contains(err.Error(), "region") {
		t.Errorf("error should hint at region, got: %s", err)
	}
}

// Gateway is optional now: empty means talk directly to the region's public
// endpoints, so Validate must accept a config with no gateway.
func TestValidate_NoGateway_OK(t *testing.T) {
	cfg := &core.Config{AK: "ak", SK: "sk", Region: "cn-north-4"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("should not require Gateway, got: %v", err)
	}
}

func TestValidate_NoProjectID_OK(t *testing.T) {
	cfg := &core.Config{AK: "ak", SK: "sk", Region: "cn-south-1", Gateway: "http://gw:8099"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("should not require ProjectID, got: %v", err)
	}
}

func TestRedacted(t *testing.T) {
	cfg := &core.Config{
		AK: "HPUA1234567890", SK: "supersecret",
		ProjectID: "proj123", Gateway: "http://gw:8099", UserID: "user123",
	}
	r := core.Redacted(cfg)
	if r.AK != "HPUA****" {
		t.Errorf("Redacted AK = %q, want HPUA****", r.AK)
	}
	if r.SK != "****" {
		t.Errorf("Redacted SK = %q, want ****", r.SK)
	}
	if r.ProjectID != "proj123" || r.UserID != "user123" {
		t.Errorf("non-secret fields mutated: %+v", r)
	}
	if cfg.AK != "HPUA1234567890" {
		t.Error("Redacted() mutated original")
	}
}

func TestMaskLeft(t *testing.T) {
	tests := []struct {
		input string
		keep  int
		want  string
	}{
		{"", 4, "****"},
		{"abc", 4, "****"},
		{"abcd", 4, "****"},
		{"abcde", 4, "abcd****"},
		{"HPUA1234567890", 4, "HPUA****"},
	}
	for _, tt := range tests {
		got := core.MaskLeft(tt.input, tt.keep)
		if got != tt.want {
			t.Errorf("maskLeft(%q, %d) = %q, want %q", tt.input, tt.keep, got, tt.want)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &core.Config{
		AK: "testAK", SK: "testSK", ProjectID: "proj",
		Gateway: "http://gw:8099", UserID: "user",
	}
	if err := core.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	p := filepath.Join(tmpDir, ".codearts-cli", "config.json")
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("config file not found: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config file mode = %o, want 600", info.Mode().Perm())
	}

	loaded, err := core.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.AK != "testAK" || loaded.SK != "testSK" || loaded.UserID != "user" || loaded.Gateway != "http://gw:8099" {
		t.Errorf("Load() fields mismatch: %+v", loaded)
	}
}

func TestLoad_NoFile_ReturnsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := core.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ProjectID != core.DefaultProjectID {
		t.Errorf("default ProjectID = %q, want %q", cfg.ProjectID, core.DefaultProjectID)
	}
	if cfg.Gateway != core.DefaultGateway {
		t.Errorf("default Gateway = %q, want %q", cfg.Gateway, core.DefaultGateway)
	}
	if cfg.Region != core.DefaultRegion {
		t.Errorf("default Region = %q, want %q", cfg.Region, core.DefaultRegion)
	}
}

func TestLoad_IgnoresUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	dir := filepath.Join(tmpDir, ".codearts-cli")
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"ak":"ak","sk":"sk","projectId":"p","gateway":"http://gw:8099","endpoint":"https://old.com"}`), 0o600)

	cfg, err := core.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AK != "ak" {
		t.Errorf("AK = %q, want ak", cfg.AK)
	}
	if cfg.Gateway != "http://gw:8099" {
		t.Errorf("Gateway = %q, want http://gw:8099", cfg.Gateway)
	}
}

// TestLoad_GatewaylessConfig verifies that a config file with a region but no
// gateway loads with an empty Gateway (i.e. it is NOT backfilled to the
// internal default). An empty gateway is the intended "hit the region's public
// endpoints directly" state, which a public multi-region user relies on.
func TestLoad_GatewaylessConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	dir := filepath.Join(tmpDir, ".codearts-cli")
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"ak":"ak","sk":"sk","projectId":"p","region":"cn-north-4"}`), 0o600)

	cfg, err := core.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Gateway != "" {
		t.Errorf("gatewayless config should keep Gateway empty, got %q", cfg.Gateway)
	}
	if cfg.Region != "cn-north-4" {
		t.Errorf("Region = %q, want cn-north-4", cfg.Region)
	}
}

// TestLoad_BackfillsRegion verifies that a pre-region config (no region field)
// backfills Region to the default so it keeps hitting Guangzhou.
func TestLoad_BackfillsRegion(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	dir := filepath.Join(tmpDir, ".codearts-cli")
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"ak":"ak","sk":"sk","projectId":"p","gateway":"http://gw:8099"}`), 0o600)

	cfg, err := core.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Region != core.DefaultRegion {
		t.Errorf("pre-region config should backfill Region, got %q", cfg.Region)
	}
}
