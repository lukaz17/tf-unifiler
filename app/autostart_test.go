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
	"testing"
	"time"

	"github.com/tforceaio/tf-unifiler/server"
)

func TestEnsureServerSpawnsThenReachable(t *testing.T) {
	lock := server.NewLock(t.TempDir()) // no live owner
	calls := 0
	probe := func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("refused")
		}
		return nil
	}
	spawned := 0
	spawn := func() error { spawned++; return nil }

	if err := ensureServer(context.Background(), lock, probe, spawn, 2*time.Second, time.Millisecond); err != nil {
		t.Fatalf("ensureServer: %v", err)
	}
	if spawned != 1 {
		t.Errorf("expected exactly one spawn, got %d", spawned)
	}
}

func TestEnsureServerSkipsSpawnWhenLiveOwner(t *testing.T) {
	lock := server.NewLock(t.TempDir())
	if err := lock.Acquire(); err != nil { // this process becomes the live owner
		t.Fatalf("Acquire: %v", err)
	}
	t.Cleanup(func() { _ = lock.Release() })

	spawned := 0
	spawn := func() error { spawned++; return nil }
	probe := func(context.Context) error { return nil }

	if err := ensureServer(context.Background(), lock, probe, spawn, time.Second, time.Millisecond); err != nil {
		t.Fatalf("ensureServer: %v", err)
	}
	if spawned != 0 {
		t.Errorf("expected no spawn when a live owner holds the lock, got %d", spawned)
	}
}

func TestEnsureServerSpawnError(t *testing.T) {
	lock := server.NewLock(t.TempDir())
	probe := func(context.Context) error { return nil }
	spawn := func() error { return errors.New("boom") }

	if err := ensureServer(context.Background(), lock, probe, spawn, time.Second, time.Millisecond); err == nil {
		t.Fatal("expected error when spawn fails")
	}
}

func TestEnsureServerTimesOut(t *testing.T) {
	lock := server.NewLock(t.TempDir())
	probe := func(context.Context) error { return errors.New("never") }
	spawn := func() error { return nil }

	if err := ensureServer(context.Background(), lock, probe, spawn, 30*time.Millisecond, 5*time.Millisecond); err == nil {
		t.Fatal("expected timeout error when the server never becomes reachable")
	}
}
