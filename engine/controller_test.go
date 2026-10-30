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
	"context"
	"io"
	"testing"

	"github.com/tforceaio/tf-unifiler/config"
)

func TestTaskScopedControllerNeverPrompts(t *testing.T) {
	tests := []struct {
		name        string
		interactive bool
	}{
		{"server started in a terminal", true},
		{"server started in the background", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &Controller{Config: &config.RootConfig{}, Interactive: tt.interactive}
			ctx := context.WithValue(context.Background(), taskRuntimeKey{}, &taskRuntimeValue{base: base, out: io.Discard})

			c := getTaskScopedController(ctx)
			if c == nil {
				t.Fatal("expected a task-scoped controller, got nil")
			}
			if c.Interactive {
				t.Error("task-scoped controller is interactive; queued commands would prompt on the server's terminal")
			}
			if m := NewFileModule(c, "pack"); m.interactive {
				t.Error("file module of a task is interactive; file pack would open the options picker on the server's terminal")
			}
		})
	}
}
