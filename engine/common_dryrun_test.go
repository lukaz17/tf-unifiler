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
	"testing"

	"github.com/spf13/cobra"
)

func TestDryRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPath string // command path of the command that would run, "" for none.
		wantErr  bool
	}{
		{"returns the command that would run", []string{"group", "leaf", "--name", "a", "--count", "1"}, "root group leaf", false},
		{"returns no command for help", []string{"group", "leaf", "--help"}, "", false},
		{"returns no command for a command group", []string{"group"}, "", false},
		{"rejects an unknown flag", []string{"group", "leaf", "--name", "a", "--bogus"}, "", true},
		{"rejects a bad flag value", []string{"group", "leaf", "--name", "a", "--count", "x"}, "", true},
		{"rejects an unknown command", []string{"bogus"}, "", true},
		{"rejects a missing required flag", []string{"group", "leaf"}, "", true},
		{"rejects a violated flag group", []string{"group", "leaf", "--name", "a", "--left", "--right"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, _ := newDryRunTree()

			cmd, err := DryRun(root, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("DryRun(%q): expected error, got command %v", tt.args, cmd)
				}
				return
			}
			if err != nil {
				t.Fatalf("DryRun(%q): unexpected error: %v", tt.args, err)
			}
			gotPath := ""
			if cmd != nil {
				gotPath = cmd.CommandPath()
			}
			if gotPath != tt.wantPath {
				t.Errorf("DryRun(%q) command = %q, want %q", tt.args, gotPath, tt.wantPath)
			}
		})
	}
}

func TestDryRunSkipsCommandBody(t *testing.T) {
	root, ran := newDryRunTree()

	if _, err := DryRun(root, []string{"group", "leaf", "--name", "a"}); err != nil {
		t.Fatalf("DryRun: unexpected error: %v", err)
	}

	if *ran {
		t.Error("DryRun ran the command body")
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"accepts a valid command", []string{"file", "pack", "-o", "a.7z", "--password", "secret", "a.txt"}, false},
		{"accepts flags before the subcommand", []string{"file", "-i", "a.txt", "hash"}, false},
		{"rejects empty args", nil, true},
		{"rejects an unknown flag", []string{"file", "pack", "--pasword", "secret"}, true},
		{"rejects an unknown command", []string{"bogus"}, true},
		{"rejects a bad flag value", []string{"file", "pack", "--dict", "notanint"}, true},
		{"rejects the client-only server command", []string{"server"}, true},
		{"rejects the client-only server start command", []string{"server", "start"}, true},
		{"rejects the client-only queue flag", []string{"file", "hash", "--queue"}, true},
		{"rejects help", []string{"file", "hash", "--help"}, true},
		{"rejects version", []string{"--version"}, true},
		{"rejects a command group", []string{"file"}, true},
		{"rejects Cobra's help command", []string{"help", "file"}, true},
		{"rejects Cobra's completion command", []string{"completion", "bash"}, true},
		{"rejects Cobra's shell completion request", []string{"__complete", "file", "pack", "--password", "secret"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArgs(tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("ValidateArgs(%q): expected error, got nil", tt.args)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateArgs(%q): unexpected error: %v", tt.args, err)
			}
		})
	}
}

// Return a root holding a "group" command with a runnable "leaf" that requires --name and allows only one of --left and --right,
// and a flag set when the leaf's body runs.
func newDryRunTree() (*cobra.Command, *bool) {
	ran := false
	root := &cobra.Command{Use: "root"}
	group := &cobra.Command{Use: "group"}
	leaf := &cobra.Command{
		Use: "leaf",
		RunE: func(*cobra.Command, []string) error {
			ran = true
			return nil
		},
	}
	leaf.Flags().Int("count", 0, "")
	leaf.Flags().String("name", "", "")
	leaf.Flags().Bool("left", false, "")
	leaf.Flags().Bool("right", false, "")
	if err := leaf.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	leaf.MarkFlagsMutuallyExclusive("left", "right")
	group.AddCommand(leaf)
	root.AddCommand(group)
	return root, &ran
}
