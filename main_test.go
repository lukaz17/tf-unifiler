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
	"testing"

	"github.com/tforceaio/tf-unifiler/engine"
)

// The CLI root is the engine's non-serve tree plus the serve command owned by this package.
func TestRootCommandWiring(t *testing.T) {
	root := engine.BuildRootCmd()
	root.AddCommand(ServeCmd())

	want := map[string]bool{
		"checksum": false,
		"config":   false,
		"file":     false,
		"metadata": false,
		"mirror":   false,
		"video":    false,
		"serve":    false,
	}
	for _, c := range root.Commands() {
		want[c.Name()] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("root command missing subcommand %q", name)
		}
	}
}

// RunTask dispatches through the engine tree, which has no serve command, so a serve task must fail rather than start a nested server.
func TestServeIsNotRunnableAsTask(t *testing.T) {
	c := engine.NewController(false)
	defer c.Close()

	var sink discardWriter
	if err := engine.RunTask(c, []string{"serve"}, nil, &sink); err == nil {
		t.Fatal("expected error running 'serve' as a task, got nil")
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
