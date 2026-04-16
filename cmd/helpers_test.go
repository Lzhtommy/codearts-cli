package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// ---- parseRepoID ----

func TestParseRepoID_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},
		{"8147520", 8147520},
		{"2147483647", 2147483647},
	}
	for _, tt := range tests {
		got, err := parseRepoID(tt.input)
		if err != nil {
			t.Errorf("parseRepoID(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("parseRepoID(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseRepoID_Invalid(t *testing.T) {
	tests := []struct {
		input string
		hint  string
	}{
		{"0", "positive"},
		{"-1", "positive"},
		{"abc", "positive"},
		{"759278abbfb14b098eeddc548741f38b", "NOT the 32-char project UUID"},
		{"", "positive"},
		{"12.5", "positive"},
	}
	for _, tt := range tests {
		_, err := parseRepoID(tt.input)
		if err == nil {
			t.Errorf("parseRepoID(%q) should fail", tt.input)
			continue
		}
		if !containsStr(err.Error(), tt.hint) {
			t.Errorf("parseRepoID(%q) error should contain %q, got: %s", tt.input, tt.hint, err)
		}
	}
}

// ---- extractStringFromResp ----

func TestExtractStringFromResp_TopLevel(t *testing.T) {
	resp := map[string]interface{}{"id": "123"}
	got := extractStringFromResp(resp, "id")
	if got != "123" {
		t.Errorf("extractStringFromResp top-level = %q, want 123", got)
	}
}

func TestExtractStringFromResp_ResultArray(t *testing.T) {
	// Huawei envelope: {"result": [{"id": "456"}]}
	resp := map[string]interface{}{
		"result": []interface{}{
			map[string]interface{}{"id": "456", "title": "bug"},
		},
	}
	got := extractStringFromResp(resp, "id")
	if got != "456" {
		t.Errorf("extractStringFromResp result array = %q, want 456", got)
	}
}

func TestExtractStringFromResp_ResultMap(t *testing.T) {
	resp := map[string]interface{}{
		"result": map[string]interface{}{"id": "789"},
	}
	got := extractStringFromResp(resp, "id")
	if got != "789" {
		t.Errorf("extractStringFromResp result map = %q, want 789", got)
	}
}

func TestExtractStringFromResp_NotFound(t *testing.T) {
	resp := map[string]interface{}{"other": "val"}
	got := extractStringFromResp(resp, "id")
	if got != "" {
		t.Errorf("extractStringFromResp missing key = %q, want empty", got)
	}
}

func TestExtractStringFromResp_EmptyResult(t *testing.T) {
	resp := map[string]interface{}{"result": []interface{}{}}
	got := extractStringFromResp(resp, "id")
	if got != "" {
		t.Errorf("extractStringFromResp empty array = %q, want empty", got)
	}
}

func TestExtractStringFromResp_NonStringValue(t *testing.T) {
	resp := map[string]interface{}{"id": 12345}
	got := extractStringFromResp(resp, "id")
	if got != "" {
		t.Errorf("extractStringFromResp non-string = %q, want empty", got)
	}
}

// ---- firstNonEmpty ----

func TestFirstNonEmpty_Inline(t *testing.T) {
	got, err := firstNonEmpty("--body", `{"a":1}`, "--body-file", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `{"a":1}` {
		t.Errorf("got %q, want inline value", got)
	}
}

func TestFirstNonEmpty_BothSet(t *testing.T) {
	_, err := firstNonEmpty("--body", "val", "--body-file", "file.json")
	if err == nil {
		t.Fatal("should error when both inline and file are set")
	}
	if !containsStr(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutual exclusivity, got: %s", err)
	}
}

func TestFirstNonEmpty_File(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.json")
	os.WriteFile(tmpFile, []byte(`  {"b":2}  `), 0o644)

	got, err := firstNonEmpty("--body", "", "--body-file", tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `{"b":2}` {
		t.Errorf("got %q, want trimmed file content", got)
	}
}

func TestFirstNonEmpty_FileMissing(t *testing.T) {
	_, err := firstNonEmpty("--body", "", "--body-file", "/tmp/nonexistent_codearts_test.json")
	if err == nil {
		t.Fatal("should error on missing file")
	}
	if !containsStr(err.Error(), "nonexistent_codearts_test.json") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestFirstNonEmpty_FileEmpty(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.json")
	os.WriteFile(tmpFile, []byte("   "), 0o644)

	_, err := firstNonEmpty("--body", "", "--body-file", tmpFile)
	if err == nil {
		t.Fatal("should error on empty file")
	}
	if !containsStr(err.Error(), "empty") {
		t.Errorf("error should mention empty, got: %s", err)
	}
}

func TestFirstNonEmpty_BothEmpty(t *testing.T) {
	got, err := firstNonEmpty("--body", "", "--body-file", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
