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
	"strconv"
	"strings"
)

// lockFileName is the pidfile in the config dir that enforces a single running server.
const lockFileName = "unifiler.server.lock"

// ErrLockHeld reports that a live server instance already owns the lock.
var ErrLockHeld = errors.New("tf-unifiler server lock held by a running instance")

// Lock is a pidfile guarding the server singleton; its content is the owning server's pid.
type Lock struct {
	path string
}

// Return a Lock backed by the pidfile in dir.
func NewLock(dir string) *Lock {
	return &Lock{path: filepath.Join(dir, lockFileName)}
}

// Acquire claims the lock for the current process, reclaiming a stale pidfile left by a dead owner.
func (l *Lock) Acquire() error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return fmt.Errorf("creating lock directory: %w", err)
	}
	// Two attempts settle a race between two starters: the loser re-reads the winner's now-live pid on the second pass and gets ErrLockHeld.
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			return l.writePID(f)
		}
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("creating lock: %w", err)
		}
		if pid, alive := l.Owner(); alive {
			return fmt.Errorf("%w (pid %d)", ErrLockHeld, pid)
		}
		if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing stale lock: %w", err)
		}
	}
	return ErrLockHeld
}

// Release removes the pidfile; safe to call when the lock is already gone.
func (l *Lock) Release() error {
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing lock: %w", err)
	}
	return nil
}

// Report the pid recorded in the pidfile and whether that process is still alive.
// A missing, empty, or malformed pidfile reports (0, false).
func (l *Lock) Owner() (int, bool) {
	b, err := os.ReadFile(l.path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, processAlive(pid)
}

// Record the current pid in the freshly created lock file f, cleaning up on failure.
func (l *Lock) writePID(f *os.File) error {
	if _, err := fmt.Fprintf(f, "%d\n", os.Getpid()); err != nil {
		_ = f.Close()
		_ = os.Remove(l.path)
		return fmt.Errorf("writing pid to lock: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(l.path)
		return fmt.Errorf("closing lock: %w", err)
	}
	return nil
}
