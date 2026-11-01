// Copyright (C) 2025  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// helperEnv makes this test binary act as the external app run by Run.
const helperEnv = "UNIFILER_EXEC_HELPER"

// helperArgs re-runs this test binary as the external app, executing TestHelperProcess only.
type helperArgs struct{}

func (helperArgs) Compile() []string {
	return []string{"-test.run=^TestHelperProcess$"}
}

// Act as the external app for Run tests: echo stdin to stdout and write a marker to stderr.
func TestHelperProcess(t *testing.T) {
	if os.Getenv(helperEnv) != "1" {
		return
	}
	in, _ := io.ReadAll(os.Stdin)
	fmt.Fprintf(os.Stdout, "stdout:%s\n", in)
	fmt.Fprintln(os.Stderr, "stderr")
	os.Exit(0)
}

func TestRunStreamsToOut(t *testing.T) {
	t.Setenv(helperEnv, "1")
	tests := []struct {
		name string
		in   io.Reader
		want []string
	}{
		{"input is fed to the app", strings.NewReader("ping"), []string{"stdout:ping\n", "stderr\n"}},
		{"no input reads nothing", nil, []string{"stdout:\n", "stderr\n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := Run(os.Args[0], helperArgs{}, tt.in, &out)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if got != "" {
				t.Errorf("Run returned %q; output streamed to out must not be returned too", got)
			}
			for _, want := range tt.want {
				if !strings.Contains(out.String(), want) {
					t.Errorf("out does not contain %q; got:\n%s", want, out.String())
				}
			}
		})
	}
}

func TestRunCapturesStdoutWithoutOut(t *testing.T) {
	t.Setenv(helperEnv, "1")

	got, err := Run(os.Args[0], helperArgs{}, strings.NewReader("ping"), nil)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got != "stdout:ping\n" {
		t.Errorf("Run returned %q, want the app's stdout %q", got, "stdout:ping\n")
	}
}
