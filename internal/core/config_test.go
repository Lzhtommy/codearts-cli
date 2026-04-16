package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_OK(t *testing.T) {
	cfg := &Config{AK: "ak", SK: "sk", Region: "cn-south-1"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestValidate_MissingAK(t *testing.T) {
	cfg := &Config{SK: "sk", Region: "cn-south-1"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with empty AK")
	}
	if !containsStr(err.Error(), "config init") {
		t.Errorf("error should hint at config init, got: %s", err)
	}
}

func TestValidate_MissingSK(t *testing.T) {
	cfg := &Config{AK: "ak", Region: "cn-south-1"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with empty SK")
	}
	if !containsStr(err.Error(), "config init") {
		t.Errorf("error should hint at config init, got: %s", err)
	}
}

func TestValidate_MissingRegion(t *testing.T) {
	cfg := &Config{AK: "ak", SK: "sk"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with empty Region")
	}
	if !containsStr(err.Error(), "config set") {
		t.Errorf("error should hint at config set, got: %s", err)
	}
}

func TestValidate_NoProjectID_OK(t *testing.T) {
	// ProjectID is no longer required in Validate — pipeline/repo commands
	// require it via flag.
	cfg := &Config{AK: "ak", SK: "sk", Region: "cn-south-1"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() should not require ProjectID, got: %v", err)
	}
}

func TestRedacted(t *testing.T) {
	cfg := &Config{
		AK:        "HPUA1234567890",
		SK:        "supersecret",
		ProjectID: "proj123",
		Region:    "cn-south-1",
		UserID:    "user123",
	}
	r := Redacted(cfg)
	if r.AK != "HPUA****" {
		t.Errorf("Redacted AK = %q, want HPUA****", r.AK)
	}
	if r.SK != "****" {
		t.Errorf("Redacted SK = %q, want ****", r.SK)
	}
	// Non-secret fields should be unchanged.
	if r.ProjectID != "proj123" {
		t.Errorf("Redacted ProjectID = %q, want proj123", r.ProjectID)
	}
	if r.UserID != "user123" {
		t.Errorf("Redacted UserID = %q, want user123", r.UserID)
	}
	// Original should not be mutated.
	if cfg.AK != "HPUA1234567890" {
		t.Error("Redacted() mutated the original config")
	}
}

func TestMaskLeft(t *testing.T) {
	tests := []struct {
		input string
		keep  int
		want  string
	}{
		{"", 4, "****"},
		{"abc", 4, "****"},       // shorter than keep
		{"abcd", 4, "****"},      // equal to keep
		{"abcde", 4, "abcd****"}, // longer than keep
		{"HPUA1234567890", 4, "HPUA****"},
	}
	for _, tt := range tests {
		got := maskLeft(tt.input, tt.keep)
		if got != tt.want {
			t.Errorf("maskLeft(%q, %d) = %q, want %q", tt.input, tt.keep, got, tt.want)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	// Override HOME so config lands in tmpDir.
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{
		AK:        "testAK",
		SK:        "testSK",
		ProjectID: "proj",
		Region:    "cn-south-1",
		UserID:    "user",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists and has restricted permissions.
	p := filepath.Join(tmpDir, configDirName, configFileName)
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("config file not found: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config file mode = %o, want 600", info.Mode().Perm())
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.AK != "testAK" || loaded.SK != "testSK" {
		t.Errorf("Load() AK=%q SK=%q, want testAK/testSK", loaded.AK, loaded.SK)
	}
	if loaded.ProjectID != "proj" || loaded.Region != "cn-south-1" || loaded.UserID != "user" {
		t.Errorf("Load() fields mismatch: %+v", loaded)
	}
}

func TestLoad_NoFile_ReturnsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error on missing file: %v", err)
	}
	if cfg.ProjectID != DefaultProjectID {
		t.Errorf("default ProjectID = %q, want %q", cfg.ProjectID, DefaultProjectID)
	}
	if cfg.Region != DefaultRegion {
		t.Errorf("default Region = %q, want %q", cfg.Region, DefaultRegion)
	}
}

func TestLoad_IgnoresUnknownFields(t *testing.T) {
	// Simulates old config.json with "endpoint" field (now removed).
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	dir := filepath.Join(tmpDir, configDirName)
	os.MkdirAll(dir, 0o700)
	data := `{"ak":"ak","sk":"sk","projectId":"p","region":"r","endpoint":"https://old.com"}`
	os.WriteFile(filepath.Join(dir, configFileName), []byte(data), 0o600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AK != "ak" {
		t.Errorf("AK = %q, want ak", cfg.AK)
	}
	// endpoint field is silently ignored (no panic).
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
