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

package app

import (
	"testing"

	"github.com/tforceaio/tf-unifiler/engine"
)

func TestRootCommandWiring(t *testing.T) {
	root := buildRootCmd()

	want := map[string]bool{
		"checksum": false,
		"config":   false,
		"file":     false,
		"metadata": false,
		"mirror":   false,
		"video":    false,
		"server":   false,
	}
	for _, c := range root.Commands() {
		want[c.Name()] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("root command missing subcommand %q", name)
		}
	}

	if root.PersistentFlags().Lookup(queueFlagName) == nil {
		t.Errorf("root command missing persistent --%s flag", queueFlagName)
	}
}

func TestServerCommandResolution(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPath string
		wantErr  bool
	}{
		{"server alone runs the shorthand", []string{"server"}, "unifiler server", false},
		{"server start runs start", []string{"server", "start"}, "unifiler server start", false},
		{"rejects an unknown server subcommand", []string{"server", "bogus"}, "", true},
		{"rejects an argument to server start", []string{"server", "start", "bogus"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := engine.DryRun(buildRootCmd(), tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DryRun(%q): expected error, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("DryRun(%q): unexpected error: %v", tt.args, err)
			}
			if cmd == nil || cmd.RunE == nil {
				t.Fatalf("DryRun(%q): expected a runnable command, got %v", tt.args, cmd)
			}
			if got := cmd.CommandPath(); got != tt.wantPath {
				t.Errorf("DryRun(%q) command = %q, want %q", tt.args, got, tt.wantPath)
			}
		})
	}
}

func TestServerIsNotRunnableAsTask(t *testing.T) {
	c := engine.NewController(false)
	defer c.Close()

	for _, args := range [][]string{{"server"}, {"server", "start"}} {
		var sink discardWriter
		if err := engine.RunTask(c, args, nil, &sink); err == nil {
			t.Errorf("expected error running %q as a task, got nil", args)
		}
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
