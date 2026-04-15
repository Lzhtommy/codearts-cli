// Package output provides consistent stdout/stderr formatting for codearts-cli.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PrintJSON writes v as indented JSON to w.
func PrintJSON(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "output: failed to marshal json: %v\n", err)
	}
}

// Successf writes an info-level message to stderr, keeping stdout clean for
// JSON consumers.
func Successf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, "✓ "+format+"\n", args...)
}

// Errorf writes an error-level message to stderr.
func Errorf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, "✗ "+format+"\n", args...)
}
