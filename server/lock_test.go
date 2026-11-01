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
	"fmt"
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
	t.Cleanup(func() { _ = lock.Release() })

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
	owner := NewLock(dir)
	if err := owner.Acquire(); err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	t.Cleanup(func() { _ = owner.Release() })

	err := NewLock(dir).Acquire()
	if !errors.Is(err, ErrLockHeld) {
		t.Fatalf("expected ErrLockHeld, got %v", err)
	}
}

func TestLockAcquireRefusesOwnerBeforePIDWritten(t *testing.T) {
	dir := t.TempDir()
	owner := NewLock(dir)
	if err := owner.Acquire(); err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	t.Cleanup(func() { _ = owner.Release() })
	// The owner holds the lock but has not written its pid yet.
	if err := os.Truncate(filepath.Join(dir, lockFileName), 0); err != nil {
		t.Fatalf("emptying lock file: %v", err)
	}

	err := NewLock(dir).Acquire()
	if !errors.Is(err, ErrLockHeld) {
		t.Fatalf("expected ErrLockHeld, got %v", err)
	}
}

func TestLockAcquireReclaimsPIDOfProcessWithoutLock(t *testing.T) {
	dir := t.TempDir()
	// A crashed owner's pid reused by a live process holding no lock: this test process.
	if err := os.WriteFile(filepath.Join(dir, lockFileName), fmt.Appendf(nil, "%d\n", os.Getpid()), 0o600); err != nil {
		t.Fatalf("seeding stale lock: %v", err)
	}
	lock := NewLock(dir)

	err := lock.Acquire()
	t.Cleanup(func() { _ = lock.Release() })

	if err != nil {
		t.Fatalf("Acquire over a reused pid: %v", err)
	}
}

func TestLockOwnerIgnoresPIDOfProcessWithoutLock(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, lockFileName), fmt.Appendf(nil, "%d\n", os.Getpid()), 0o600); err != nil {
		t.Fatalf("seeding stale lock: %v", err)
	}

	pid, alive := NewLock(dir).Owner()

	if pid != 0 || alive {
		t.Errorf("expected no owner, got pid %d alive=%v", pid, alive)
	}
}

func TestLockAcquireReclaimsStaleLock(t *testing.T) {
	dir := t.TempDir()
	// A lock file no process holds is reclaimable whatever it contains.
	if err := os.WriteFile(filepath.Join(dir, lockFileName), []byte("not-a-pid\n"), 0o600); err != nil {
		t.Fatalf("seeding stale lock: %v", err)
	}

	lock := NewLock(dir)
	if err := lock.Acquire(); err != nil {
		t.Fatalf("Acquire over stale lock: %v", err)
	}
	t.Cleanup(func() { _ = lock.Release() })
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
	next := NewLock(dir)
	if err := next.Acquire(); err != nil {
		t.Errorf("expected reacquire after Release, got %v", err)
	}
	t.Cleanup(func() { _ = next.Release() })
}

func TestLockReleaseMissingIsNoError(t *testing.T) {
	if err := NewLock(t.TempDir()).Release(); err != nil {
		t.Errorf("Release of absent lock should be nil, got %v", err)
	}
}

func TestLockReleaseByLoserKeepsWinnerLock(t *testing.T) {
	dir := t.TempDir()
	winner := NewLock(dir)
	if err := winner.Acquire(); err != nil {
		t.Fatalf("winner Acquire: %v", err)
	}
	t.Cleanup(func() { _ = winner.Release() })
	loser := NewLock(dir)
	if err := loser.Acquire(); !errors.Is(err, ErrLockHeld) {
		t.Fatalf("expected loser to get ErrLockHeld, got %v", err)
	}

	if err := loser.Release(); err != nil {
		t.Fatalf("loser Release: %v", err)
	}

	if pid, alive := winner.Owner(); pid != os.Getpid() || !alive {
		t.Errorf("expected winner pid %d to still own the lock, got %d alive=%v", os.Getpid(), pid, alive)
	}
}
