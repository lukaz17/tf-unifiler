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
	"context"
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler/config"
)

func TestConfigWritesToTaskOutput(t *testing.T) {
	base := &Controller{Config: &config.RootConfig{Server: &config.ServerConfig{Name: "unifiler-output-sentinel"}}}
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"export to stdout", []string{"config", "export"}, "name: unifiler-output-sentinel"},
		{"get key", []string{"config", "get", "server.name"}, "unifiler-output-sentinel\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := RunTask(base, tt.args, context.Background(), &out); err != nil {
				t.Fatalf("RunTask(%v): %v", tt.args, err)
			}
			if !strings.Contains(out.String(), tt.want) {
				t.Errorf("task output does not contain %q, so the result was printed elsewhere; got:\n%s", tt.want, out.String())
			}
		})
	}
}
