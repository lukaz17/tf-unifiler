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
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// lockFileName is the file in the config dir that enforces a single running server.
const lockFileName = "unifiler.server.lock"

// ErrLockHeld reports that a live server instance already owns the lock.
var ErrLockHeld = errors.New("tf-unifiler server lock held by a running instance")

// Lock guards the server singleton with an OS file lock, which the OS frees when its owner exits, however it exits.
// The locked file records the owner's pid for display and stays in place once released.
type Lock struct {
	path string
	file *os.File // open while this Lock holds the OS lock.
}

// Return a Lock backed by the lock file in dir.
func NewLock(dir string) *Lock {
	return &Lock{path: filepath.Join(dir, lockFileName)}
}

// Acquire claims the lock for the current process, or returns ErrLockHeld when another holds it.
func (l *Lock) Acquire() error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return fmt.Errorf("creating lock directory: %w", err)
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("opening lock: %w", err)
	}
	locked, err := tryLock(f)
	if err != nil {
		return errors.Join(fmt.Errorf("locking: %w", err), f.Close())
	}
	if !locked {
		return errors.Join(fmt.Errorf("%w (pid %d)", ErrLockHeld, readPID(f)), f.Close())
	}
	if err := writePID(f); err != nil {
		return errors.Join(err, unlock(f), f.Close())
	}
	l.file = f
	return nil
}

// Release frees the lock held by this Lock; safe to call when it holds none.
func (l *Lock) Release() error {
	if l.file == nil {
		return nil
	}
	f := l.file
	l.file = nil
	if err := errors.Join(unlock(f), f.Close()); err != nil {
		return fmt.Errorf("releasing lock: %w", err)
	}
	return nil
}

// Report the pid recorded by the process holding the lock and whether any process holds it.
// It briefly takes the lock to test it, so a server acquiring at that instant gets ErrLockHeld.
// The pid is 0 when the lock is free, or when its owner has not recorded a valid pid yet.
func (l *Lock) Owner() (int, bool) {
	// Read-write because NFS emulates flock with a byte-range lock, which it refuses to make exclusive on a read-only open.
	f, err := os.OpenFile(l.path, os.O_RDWR, 0)
	if err != nil {
		return 0, false
	}
	defer f.Close()
	locked, err := tryLock(f)
	if err != nil {
		return 0, false
	}
	if locked {
		// A failed unlock changes nothing: the deferred close frees the lock too, though Windows may do it late.
		_ = unlock(f)
		return 0, false
	}
	return readPID(f), true
}

// Return the pid recorded in the lock file f, or 0 when it holds no valid pid.
func readPID(f *os.File) int {
	b, err := io.ReadAll(f)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

// Replace the content of the locked file f with the current pid.
func writePID(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("clearing lock: %w", err)
	}
	if _, err := f.WriteAt(fmt.Appendf(nil, "%d\n", os.Getpid()), 0); err != nil {
		return fmt.Errorf("writing pid to lock: %w", err)
	}
	return nil
}
