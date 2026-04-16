package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lzhtommy/codearts-cli/cmd"
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
		got, err := cmd.ParseRepoID(tt.input)
		if err != nil {
			t.Errorf("parseRepoID(%q) error: %v", tt.input, err)
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
		_, err := cmd.ParseRepoID(tt.input)
		if err == nil {
			t.Errorf("parseRepoID(%q) should fail", tt.input)
			continue
		}
		if !strings.Contains(err.Error(), tt.hint) {
			t.Errorf("parseRepoID(%q) error should contain %q, got: %s", tt.input, tt.hint, err)
		}
	}
}

// ---- extractStringFromResp ----

func TestExtractStringFromResp_TopLevel(t *testing.T) {
	resp := map[string]interface{}{"id": "123"}
	if got := cmd.ExtractStringFromResp(resp, "id"); got != "123" {
		t.Errorf("top-level = %q, want 123", got)
	}
}

func TestExtractStringFromResp_ResultArray(t *testing.T) {
	resp := map[string]interface{}{
		"result": []interface{}{
			map[string]interface{}{"id": "456"},
		},
	}
	if got := cmd.ExtractStringFromResp(resp, "id"); got != "456" {
		t.Errorf("result array = %q, want 456", got)
	}
}

func TestExtractStringFromResp_ResultMap(t *testing.T) {
	resp := map[string]interface{}{
		"result": map[string]interface{}{"id": "789"},
	}
	if got := cmd.ExtractStringFromResp(resp, "id"); got != "789" {
		t.Errorf("result map = %q, want 789", got)
	}
}

func TestExtractStringFromResp_NotFound(t *testing.T) {
	if got := cmd.ExtractStringFromResp(map[string]interface{}{}, "id"); got != "" {
		t.Errorf("missing key = %q, want empty", got)
	}
}

func TestExtractStringFromResp_EmptyResult(t *testing.T) {
	resp := map[string]interface{}{"result": []interface{}{}}
	if got := cmd.ExtractStringFromResp(resp, "id"); got != "" {
		t.Errorf("empty array = %q, want empty", got)
	}
}

func TestExtractStringFromResp_NonStringValue(t *testing.T) {
	resp := map[string]interface{}{"id": 12345}
	if got := cmd.ExtractStringFromResp(resp, "id"); got != "" {
		t.Errorf("non-string = %q, want empty", got)
	}
}

// ---- firstNonEmpty ----

func TestFirstNonEmpty_Inline(t *testing.T) {
	got, err := cmd.FirstNonEmpty("--body", `{"a":1}`, "--body-file", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != `{"a":1}` {
		t.Errorf("got %q, want inline", got)
	}
}

func TestFirstNonEmpty_BothSet(t *testing.T) {
	_, err := cmd.FirstNonEmpty("--body", "val", "--body-file", "file.json")
	if err == nil {
		t.Fatal("should error when both set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error = %s", err)
	}
}

func TestFirstNonEmpty_File(t *testing.T) {
	f := filepath.Join(t.TempDir(), "test.json")
	os.WriteFile(f, []byte(`  {"b":2}  `), 0o644)
	got, err := cmd.FirstNonEmpty("--body", "", "--body-file", f)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != `{"b":2}` {
		t.Errorf("got %q, want trimmed", got)
	}
}

func TestFirstNonEmpty_FileMissing(t *testing.T) {
	_, err := cmd.FirstNonEmpty("--body", "", "--body-file", "/tmp/nonexistent_codearts.json")
	if err == nil {
		t.Fatal("should error")
	}
	if !strings.Contains(err.Error(), "nonexistent_codearts.json") {
		t.Errorf("error should contain path, got: %s", err)
	}
}

func TestFirstNonEmpty_FileEmpty(t *testing.T) {
	f := filepath.Join(t.TempDir(), "empty.json")
	os.WriteFile(f, []byte("   "), 0o644)
	_, err := cmd.FirstNonEmpty("--body", "", "--body-file", f)
	if err == nil {
		t.Fatal("should error")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error = %s", err)
	}
}

func TestFirstNonEmpty_BothEmpty(t *testing.T) {
	got, err := cmd.FirstNonEmpty("--body", "", "--body-file", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
