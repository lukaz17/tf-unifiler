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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler/config"
)

func TestConfigSetMasksAPIKeyInLog(t *testing.T) {
	const key = "sk-unifiler-sentinel"
	path := filepath.Join(t.TempDir(), "unifiler.yaml")
	base := &Controller{Config: &config.RootConfig{ConfigFile: path}}

	var out bytes.Buffer
	if err := RunTask(base, []string{"config", "set", config.KeyServerAPIKey, key}, context.Background(), &out); err != nil {
		t.Fatalf("config set: %v", err)
	}
	if strings.Contains(out.String(), key) || !strings.Contains(out.String(), redactedValue) {
		t.Errorf("expected the logged value masked as %q; got:\n%s", redactedValue, out.String())
	}
	if data, err := os.ReadFile(path); err != nil || !strings.Contains(string(data), key) {
		t.Errorf("expected the key saved to %s, got %q (%v)", path, data, err)
	}
}

func TestConfigSetRestrictsFileToOwner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows ignores Unix permission bits except the owner write bit")
	}
	path := filepath.Join(t.TempDir(), "unifiler.yaml")
	if err := os.WriteFile(path, []byte("debug: false\n"), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	base := &Controller{Config: &config.RootConfig{ConfigFile: path}}

	if err := RunTask(base, []string{"config", "set", config.KeyServerAPIKey, "sk-unifiler-sentinel"}, context.Background(), &bytes.Buffer{}); err != nil {
		t.Fatalf("config set: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("expected mode 0600, got %#o", got)
	}
}

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

func TestConfigSetRejectsUnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unifiler.yaml")
	base := &Controller{Config: &config.RootConfig{ConfigFile: path}}

	for _, key := range []string{"logger.level", "server"} {
		if err := RunTask(base, []string{"config", "set", key, "debug"}, context.Background(), &bytes.Buffer{}); err == nil {
			t.Errorf("config set %s: expected error, got nil", key)
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s not written, got %v", path, err)
	}
}

func TestConfigSetReplacesKeyOfOtherCase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unifiler.yaml")
	if err := os.WriteFile(path, []byte("Debug: false\n"), 0600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	base := &Controller{Config: &config.RootConfig{ConfigFile: path}}

	if err := RunTask(base, []string{"config", "set", "DEBUG", "true"}, context.Background(), &bytes.Buffer{}); err != nil {
		t.Fatalf("config set: %v", err)
	}
	k, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("loading %s after config set: %v", path, err)
	}
	if got := k.String("debug"); got != "true" {
		t.Errorf("expected debug set to true, got %q", got)
	}
}
