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

package engine

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
)

func TestLogCommandErrorNil(t *testing.T) {
	var logs bytes.Buffer
	if err := LogCommandError(zerolog.New(&logs), nil); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if logs.Len() != 0 {
		t.Errorf("expected no log for a nil error, got %q", logs.String())
	}
}

func TestLogCommandErrorKeepsCause(t *testing.T) {
	cause := errors.New("boom")
	err := LogCommandError(zerolog.Nop(), cause)

	if !errors.Is(err, cause) {
		t.Errorf("expected %v to wrap %v", err, cause)
	}
	if err.Error() != cause.Error() {
		t.Errorf("expected message %q, got %q", cause.Error(), err.Error())
	}
}

func TestPrintError(t *testing.T) {
	cmd := &cobra.Command{Use: "unifiler"}
	tests := []struct {
		name    string
		err     func(logger zerolog.Logger) error
		level   zerolog.Level // level of the failing command's logger.
		want    string
		wantLog bool // whether the command's logger shows the error.
	}{
		{
			name:    "logged failure is not printed again",
			err:     func(l zerolog.Logger) error { return LogCommandError(l, errors.New("boom")) },
			level:   zerolog.InfoLevel,
			want:    "",
			wantLog: true,
		},
		{
			name:    "failure hidden by the log level is printed without usage",
			err:     func(l zerolog.Logger) error { return LogCommandError(l, errors.New("boom")) },
			level:   zerolog.FatalLevel,
			want:    "Error: boom\n",
			wantLog: false,
		},
		{
			name:    "nil error prints nothing",
			err:     func(zerolog.Logger) error { return nil },
			level:   zerolog.InfoLevel,
			want:    "",
			wantLog: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs, out bytes.Buffer
			err := tt.err(zerolog.New(&logs).Level(tt.level))

			PrintError(&out, cmd, err)

			if got := out.String(); got != tt.want {
				t.Errorf("expected output %q, got %q", tt.want, got)
			}
			if got := strings.Count(logs.String(), "\n"); got != opx.Ternary(tt.wantLog, 1, 0) {
				t.Errorf("expected logged=%t, got logs %q", tt.wantLog, logs.String())
			}
		})
	}
}

func TestPrintErrorUsageOfRejectedFlag(t *testing.T) {
	cmd, err := executeRootCmd("config", "export", "--bogus")
	if err == nil {
		t.Fatal("expected error for an unknown flag, got nil")
	}

	var out bytes.Buffer
	PrintError(&out, cmd, err)

	want := "Error: unknown flag: --bogus\n" + cmd.UsageString() + "\n"
	if got := out.String(); got != want {
		t.Errorf("expected output %q, got %q", want, got)
	}
	if !strings.Contains(out.String(), "-h, --help") {
		t.Errorf("expected the usage to list the help flag, got %q", out.String())
	}
}

func TestPrintErrorHintOfUnknownCommand(t *testing.T) {
	cmd, err := executeRootCmd("bogus")
	if err == nil {
		t.Fatal("expected error for an unknown command, got nil")
	}

	var out bytes.Buffer
	PrintError(&out, cmd, err)

	want := "Error: " + err.Error() + "\nRun 'unifiler --help' for usage.\n"
	if got := out.String(); got != want {
		t.Errorf("expected output %q, got %q", want, got)
	}
}

func TestRootCmdLeavesErrorsToCaller(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"failing command body", []string{"fail"}},
		{"unknown flag", []string{"fail", "--bogus"}},
		{"unknown command", []string{"bogus"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := BuildRootCmd()
			root.AddCommand(&cobra.Command{
				Use: "fail",
				RunE: func(cmd *cobra.Command, args []string) error {
					return LogCommandError(zerolog.Nop(), errors.New("boom"))
				},
			})
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(tt.args)

			if err := root.Execute(); err == nil {
				t.Fatalf("Execute(%q): expected error, got nil", tt.args)
			}
			if out.Len() != 0 {
				t.Errorf("Execute(%q): expected Cobra to print nothing, got %q", tt.args, out.String())
			}
		})
	}
}

// Execute args on a new root command, and return the command Cobra stopped at with its error.
func executeRootCmd(args ...string) (*cobra.Command, error) {
	root := BuildRootCmd()
	root.SetArgs(args)
	return root.ExecuteC()
}
