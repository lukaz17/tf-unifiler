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

//go:build windows

package server

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// lockOffset is where the one-byte locked region starts, far past the pid text: Windows locks are mandatory, so locking the pid would stop Owner from reading it.
const lockOffset = 1 << 30

// Try to take an exclusive lock on f without blocking, and report false when another handle holds it.
// Windows locks belong to a handle, not a process, so a second handle in this same process contends too.
func tryLock(f *os.File) (bool, error) {
	region := windows.Overlapped{Offset: lockOffset}
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &region)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, windows.ERROR_LOCK_VIOLATION):
		return false, nil
	}
	return false, err
}

// Free the lock taken on f by tryLock.
func unlock(f *os.File) error {
	region := windows.Overlapped{Offset: lockOffset}
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &region)
}
