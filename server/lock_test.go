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

package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLockAcquireWritesOwnPID(t *testing.T) {
	dir := t.TempDir()
	lock := NewLock(dir)

	if err := lock.Acquire(); err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	pid, alive := lock.Owner()
	if pid != os.Getpid() {
		t.Errorf("expected owner pid %d, got %d", os.Getpid(), pid)
	}
	if !alive {
		t.Error("expected current process to be reported alive")
	}
	if _, err := os.Stat(filepath.Join(dir, lockFileName)); err != nil {
		t.Errorf("expected lock file to exist: %v", err)
	}
}

func TestLockAcquireRefusesLiveOwner(t *testing.T) {
	dir := t.TempDir()
	if err := NewLock(dir).Acquire(); err != nil {
		t.Fatalf("first Acquire: %v", err)
	}

	err := NewLock(dir).Acquire()
	if !errors.Is(err, ErrLockHeld) {
		t.Fatalf("expected ErrLockHeld, got %v", err)
	}
}

func TestLockAcquireReclaimsStaleLock(t *testing.T) {
	dir := t.TempDir()
	// A non-numeric pidfile cannot name a live process, so it is reclaimable on any OS.
	if err := os.WriteFile(filepath.Join(dir, lockFileName), []byte("not-a-pid\n"), 0o600); err != nil {
		t.Fatalf("seeding stale lock: %v", err)
	}

	lock := NewLock(dir)
	if err := lock.Acquire(); err != nil {
		t.Fatalf("Acquire over stale lock: %v", err)
	}
	if pid, alive := lock.Owner(); pid != os.Getpid() || !alive {
		t.Errorf("expected reclaimed owner pid %d alive, got %d alive=%v", os.Getpid(), pid, alive)
	}
}

func TestLockReleaseAllowsReacquire(t *testing.T) {
	dir := t.TempDir()
	lock := NewLock(dir)
	if err := lock.Acquire(); err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	if _, alive := lock.Owner(); alive {
		t.Error("expected no live owner after Release")
	}
	if err := NewLock(dir).Acquire(); err != nil {
		t.Errorf("expected reacquire after Release, got %v", err)
	}
}

func TestLockReleaseMissingIsNoError(t *testing.T) {
	if err := NewLock(t.TempDir()).Release(); err != nil {
		t.Errorf("Release of absent lock should be nil, got %v", err)
	}
}

func TestProcessAliveCurrentProcess(t *testing.T) {
	if !processAlive(os.Getpid()) {
		t.Error("current process should be reported alive")
	}
}
