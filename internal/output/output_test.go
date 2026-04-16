package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	PrintJSON(&buf, map[string]string{"key": "value"})
	got := buf.String()
	if !strings.Contains(got, `"key": "value"`) {
		t.Errorf("PrintJSON output = %q, want key:value", got)
	}
	// Indented with 2 spaces.
	if !strings.Contains(got, "  ") {
		t.Error("PrintJSON should produce indented output")
	}
}

func TestPrintJSON_NoHTMLEscape(t *testing.T) {
	var buf bytes.Buffer
	PrintJSON(&buf, map[string]string{"url": "https://example.com?a=1&b=2"})
	got := buf.String()
	// Should NOT escape & to \u0026.
	if strings.Contains(got, `\u0026`) {
		t.Errorf("PrintJSON should not escape HTML, got: %s", got)
	}
}

func TestSuccessf(t *testing.T) {
	var buf bytes.Buffer
	Successf(&buf, "done %d", 42)
	got := buf.String()
	want := "✓ done 42\n"
	if got != want {
		t.Errorf("Successf = %q, want %q", got, want)
	}
}

func TestErrorf(t *testing.T) {
	var buf bytes.Buffer
	Errorf(&buf, "fail %s", "now")
	got := buf.String()
	want := "✗ fail now\n"
	if got != want {
		t.Errorf("Errorf = %q, want %q", got, want)
	}
}

func TestDryRunf(t *testing.T) {
	var buf bytes.Buffer
	DryRunf(&buf, "preview %s", "request")
	got := buf.String()
	want := "[dry-run] preview request\n"
	if got != want {
		t.Errorf("DryRunf = %q, want %q", got, want)
	}
}
