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
	"slices"
	"testing"
)

func TestRedactArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"masks value given as next arg", []string{"file", "pack", "-o", "a.7z", "--password", "secret"}, []string{"file", "pack", "-o", "a.7z", "--password", "******"}},
		{"masks inline value", []string{"file", "pack", "--password=secret"}, []string{"file", "pack", "--password=******"}},
		{"masks every occurrence of a repeated flag", []string{"file", "pack", "--password", "a", "--password=b"}, []string{"file", "pack", "--password", "******", "--password=******"}},
		{"masks value after combined bool shorthands", []string{"file", "pack", "-ms", "--password", "secret"}, []string{"file", "pack", "-ms", "--password", "******"}},
		{"masks value when flags precede the subcommand", []string{"file", "-i", "a", "pack", "--password", "secret"}, []string{"file", "-i", "a", "pack", "--password", "******"}},
		{"keeps another flag's value that looks like the sensitive flag", []string{"file", "pack", "-o", "--password", "a"}, []string{"file", "pack", "-o", "--password", "a"}},
		{"keeps combined shorthands' value that looks like the sensitive flag", []string{"file", "pack", "-mo", "--password", "a"}, []string{"file", "pack", "-mo", "--password", "a"}},
		{"keeps args after the terminator", []string{"file", "pack", "--", "--password", "secret"}, []string{"file", "pack", "--", "--password", "secret"}},
		{"keeps sensitive flag without value", []string{"file", "pack", "--password"}, []string{"file", "pack", "--password"}},
		{"keeps value of a non-sensitive flag", []string{"file", "rename", "-p", "crc32"}, []string{"file", "rename", "-p", "crc32"}},
		{"keeps args of an unknown command", []string{"nope", "--password", "secret"}, []string{"nope", "--password", "secret"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactArgs(tt.args)

			if !slices.Equal(got, tt.want) {
				t.Errorf("RedactArgs(%q) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestRedactArgsKeepsInputUntouched(t *testing.T) {
	args := []string{"file", "pack", "--password", "secret"}

	RedactArgs(args)

	if args[3] != "secret" {
		t.Errorf("input args were modified: %q", args)
	}
}
