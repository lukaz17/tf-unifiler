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
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tforceaio/tf-unifiler/server"
	"github.com/tforceaio/tf-unifiler/task"
)

const (
	stopTestTimeout  = time.Second
	stopTestInterval = 5 * time.Millisecond
)

func TestCheckListenAuth(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		apiKey  string
		wantErr bool
	}{
		{"loopback without key is allowed", "127.0.0.1", "", false},
		{"localhost without key is allowed", "localhost", "", false},
		{"all interfaces without key is refused", "0.0.0.0", "", true},
		{"empty host without key is refused", "", "", true},
		{"LAN IP without key is refused", "192.168.1.5", "", true},
		{"LAN IP with key is allowed", "192.168.1.5", "s3cret", false},
		{"all interfaces with key is allowed", "0.0.0.0", "s3cret", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkListenAuth(tt.host, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkListenAuth(%q, key set: %t): expected error %t, got %v", tt.host, tt.apiKey != "", tt.wantErr, err)
			}
		})
	}
}

func TestStopServerStopsServer(t *testing.T) {
	ts := serveStoppable(t, "tf-unifiler", "", t.TempDir())

	if err := stopServer(context.Background(), server.NewClient(ts.URL, ""), "tf-unifiler", true, stopTestTimeout, stopTestInterval); err != nil {
		t.Fatalf("stopServer: %v", err)
	}
	if _, err := server.NewClient(ts.URL, "").Status(context.Background()); err == nil {
		t.Error("server still reachable after stopServer returned")
	}
}

func TestStopServerRefusesForeignService(t *testing.T) {
	ts := serveStoppable(t, "something-else", "", t.TempDir())

	if err := stopServer(context.Background(), server.NewClient(ts.URL, ""), "tf-unifiler", true, stopTestTimeout, stopTestInterval); err == nil {
		t.Fatal("expected error stopping a foreign service, got nil")
	}
	if _, err := server.NewClient(ts.URL, "").Status(context.Background()); err != nil {
		t.Errorf("foreign service was shut down: %v", err)
	}
}

func TestStopServerWrongAPIKey(t *testing.T) {
	ts := serveStoppable(t, "tf-unifiler", "s3cret", t.TempDir())

	err := stopServer(context.Background(), server.NewClient(ts.URL, "wrong"), "tf-unifiler", true, stopTestTimeout, stopTestInterval)
	if !errors.Is(err, server.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestStopServerUnreachable(t *testing.T) {
	ts := serveStoppable(t, "tf-unifiler", "", t.TempDir())
	ts.Close()

	if err := stopServer(context.Background(), server.NewClient(ts.URL, ""), "tf-unifiler", true, stopTestTimeout, stopTestInterval); err == nil {
		t.Fatal("expected error when no server is reachable, got nil")
	}
}

func TestStopServerWaitsForLockOwner(t *testing.T) {
	tests := []struct {
		name    string
		local   bool
		wantErr bool
	}{
		{"local server waits for the lock owner", true, true},
		{"remote server ignores the lock", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The lock's owner is this test process, which outlives the wait.
			dir := t.TempDir()
			lock := server.NewLock(dir)
			if err := lock.Acquire(); err != nil {
				t.Fatalf("Acquire: %v", err)
			}
			t.Cleanup(func() { _ = lock.Release() })
			ts := serveStoppable(t, "tf-unifiler", "", dir)

			err := stopServer(context.Background(), server.NewClient(ts.URL, ""), "tf-unifiler", tt.local, 50*time.Millisecond, stopTestInterval)
			if (err != nil) != tt.wantErr {
				t.Errorf("stopServer(local: %t): expected error %t, got %v", tt.local, tt.wantErr, err)
			}
		})
	}
}

func TestStopServerTimesOut(t *testing.T) {
	// A server that accepts the request but never goes away.
	ts := httptest.NewServer(server.New(":0", "", server.StatusInfo{Service: "tf-unifiler"}, idleTasks{}).Handler())
	t.Cleanup(ts.Close)

	if err := stopServer(context.Background(), server.NewClient(ts.URL, ""), "tf-unifiler", false, 50*time.Millisecond, stopTestInterval); err == nil {
		t.Fatal("expected timeout error when the server never stops, got nil")
	}
}

// idleTasks is a server.TaskService holding no tasks.
type idleTasks struct{}

func (idleTasks) Enqueue([]string) (task.Task, error) { return task.Task{}, task.ErrClosed }
func (idleTasks) Get(string) (task.Task, bool)        { return task.Task{}, false }
func (idleTasks) List() []task.Task                   { return nil }
func (idleTasks) Cancel(string) error                 { return task.ErrNotFound }
func (idleTasks) Clear() int                          { return 0 }
func (idleTasks) Stats() (int, int)                   { return 0, 0 }

// Serve service over httptest, guarded by apiKey and reporting configDir, and close it once a client requests a shutdown.
func serveStoppable(t *testing.T, service, apiKey, configDir string) *httptest.Server {
	t.Helper()
	s := server.New(":0", apiKey, server.StatusInfo{Service: service, ConfigDir: configDir}, idleTasks{})
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	go func() {
		<-s.ShutdownRequested()
		ts.Close()
	}()
	return ts
}
