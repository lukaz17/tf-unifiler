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

package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler/engine"
)

func TestStripQueueFlag(t *testing.T) {
	cases := []struct {
		name string
		args []string
		rest []string
	}{
		{"absent", []string{"file", "hash", "-i", "a.txt"}, []string{"file", "hash", "-i", "a.txt"}},
		{"bare at end", []string{"file", "hash", "--queue"}, []string{"file", "hash"}},
		{"bare in middle", []string{"file", "--queue", "hash"}, []string{"file", "hash"}},
		{"bare leading", []string{"--queue", "file", "hash"}, []string{"file", "hash"}},
		// A bool flag does not consume the next token: "true" stays a positional and is forwarded.
		{"switch consumes no value", []string{"file", "hash", "--queue", "true"}, []string{"file", "hash", "true"}},
		{"value true", []string{"file", "--queue=true"}, []string{"file"}},
		{"value false", []string{"file", "--queue=false"}, []string{"file"}},
		{"value 1", []string{"--queue=1"}, []string{}},
		{"value 0", []string{"--queue=0"}, []string{}},
		{"multiple occurrences", []string{"--queue", "file", "--queue=true", "hash"}, []string{"file", "hash"}},
		// The first -- ends flag parsing: a later --queue is a literal positional and -- is preserved.
		{"double dash keeps queue literal", []string{"--", "--queue"}, []string{"--", "--queue"}},
		{"queue before dash, literal after", []string{"--queue", "file", "rename", "--", "--weird"}, []string{"file", "rename", "--", "--weird"}},
		{"empty", nil, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rest := stripQueueFlag(tc.args)
			if !reflect.DeepEqual(rest, tc.rest) {
				t.Errorf("stripQueueFlag(%v) = %v, want %v", tc.args, rest, tc.rest)
			}
		})
	}
}

func TestIsDeferred(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		queued  bool
		wantErr bool
	}{
		{"valid local", []string{"file", "hash", "-i", "a.txt"}, false, false},
		{"valid queued", []string{"file", "hash", "-i", "a.txt", "--queue"}, true, false},
		{"queued with equals", []string{"file", "--queue=true", "hash", "-i", "a.txt"}, true, false},
		{"queue false runs locally", []string{"file", "hash", "--queue=false"}, false, false},
		// Bad syntax must be reported here, not forwarded.
		{"unknown flag", []string{"file", "hash", "--badflag", "--queue"}, false, true},
		{"unknown command", []string{"bogus", "--queue"}, false, true},
		{"bad flag value", []string{"file", "pack", "--dict", "notanint", "--queue"}, false, true},
		// Help/version short-circuit before the hook: run locally, never forwarded.
		{"help wins over queue", []string{"file", "hash", "--queue", "--help"}, false, false},
		{"version", []string{"--version"}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queued, err := isDeferred(buildRootCmd(), tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("isDeferred(%v): expected error, got nil", tc.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("isDeferred(%v): unexpected error: %v", tc.args, err)
			}
			if queued != tc.queued {
				t.Errorf("isDeferred(%v) queued = %v, want %v", tc.args, queued, tc.queued)
			}
		})
	}
}

// Confirm that a validated, queued command, once its --queue flag is stripped,
// still parses cleanly under the server's exact command tree (engine.BuildRootCmd).
func TestStrippedArgsParseOnServerTree(t *testing.T) {
	inputs := [][]string{
		{"file", "hash", "--queue"},
		{"file", "hash", "-i", "a.txt", "--queue"},
		{"file", "--queue", "hash", "-i", "a.txt"},
		{"--queue", "checksum", "create", "-a", "sha1", "x.txt"},
		{"file", "pack", "--queue=true", "-o", "out.7z", "a.txt"},
	}
	for _, args := range inputs {
		args := args
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			queued, err := isDeferred(buildRootCmd(), args)
			if err != nil {
				t.Fatalf("isDeferred(%v): unexpected error: %v", args, err)
			}
			if !queued {
				t.Fatalf("isDeferred(%v): expected queued=true", args)
			}

			rest := stripQueueFlag(args)
			server := engine.BuildRootCmd()
			cmd, flags, err := server.Find(rest)
			if err != nil {
				t.Fatalf("server tree rejected %v: %v", rest, err)
			}
			if err := cmd.ParseFlags(flags); err != nil {
				t.Fatalf("server tree failed to parse %v: %v", rest, err)
			}
		})
	}
}
