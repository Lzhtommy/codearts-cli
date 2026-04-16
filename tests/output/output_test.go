package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Lzhtommy/codearts-cli/internal/output"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	output.PrintJSON(&buf, map[string]string{"key": "value"})
	got := buf.String()
	if !strings.Contains(got, `"key": "value"`) {
		t.Errorf("PrintJSON output = %q, want key:value", got)
	}
}

func TestPrintJSON_NoHTMLEscape(t *testing.T) {
	var buf bytes.Buffer
	output.PrintJSON(&buf, map[string]string{"url": "https://example.com?a=1&b=2"})
	if strings.Contains(buf.String(), `\u0026`) {
		t.Errorf("PrintJSON should not escape HTML: %s", buf.String())
	}
}

func TestSuccessf(t *testing.T) {
	var buf bytes.Buffer
	output.Successf(&buf, "done %d", 42)
	if buf.String() != "✓ done 42\n" {
		t.Errorf("Successf = %q", buf.String())
	}
}

func TestErrorf(t *testing.T) {
	var buf bytes.Buffer
	output.Errorf(&buf, "fail %s", "now")
	if buf.String() != "✗ fail now\n" {
		t.Errorf("Errorf = %q", buf.String())
	}
}

func TestDryRunf(t *testing.T) {
	var buf bytes.Buffer
	output.DryRunf(&buf, "preview %s", "request")
	if buf.String() != "[dry-run] preview request\n" {
		t.Errorf("DryRunf = %q", buf.String())
	}
}
