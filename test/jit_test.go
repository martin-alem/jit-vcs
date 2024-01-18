package test

import (
	"bytes"
	"jit/cmd"
	"jit/pkg/util"
	"os"
	"strings"
	"testing"
)

func TestJitVersionFlag(t *testing.T) {
	// Define test cases
	var tests = []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "version flag",
			args:     []string{"jit", "-v"},
			expected: util.JitVersion,
		},
		{
			name:     "long version flag",
			args:     []string{"jit", "--version"},
			expected: util.JitVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Backup and defer restoration of os.Args and os.Stdout
			oldArgs := os.Args
			oldStdout := os.Stdout
			defer func() {
				os.Args = oldArgs
				os.Stdout = oldStdout
			}()

			// Set up the arguments for the test
			os.Args = tc.args

			// Capture the stdout output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the Jit function
			go func() {
				cmd.Jit()
				_ = w.Close()
			}()

			// Read the output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)

			// Check the output
			if got := buf.String(); !strings.Contains(got, tc.expected) {
				t.Errorf("Jit() = %q, want %q", got, tc.expected)
			}

		})
	}
}
