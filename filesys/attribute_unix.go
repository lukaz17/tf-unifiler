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

//go:build !windows

package filesys

import (
	"os"
	"time"
)

// Set multiple attributes (Archive, ReadOnly, Hidden, System) of a file or folder at once. Does nothing on Unix-like system.
func SetAttribute(path string, archive, readOnly, hidden, system int) error {
	return nil
}

// Set Archive attribute of a file or folder. Does nothing on Unix-like system.
func SetArchiveAttribute(path string, enable bool) error {
	return nil
}

// Set Created time of a file or folder. Does nothing on Unix-like system.
func SetCreatedTime(path string, ctime time.Time) error {
	return nil
}

// Set Hidden attribute of a file or folder. Does nothing on Unix-like system.
func SetHiddenAttribute(path string, enable bool) error {
	return nil
}

// Set Read Only attribute of a file or folder. Does nothing on Unix-like system.
func SetReadOnlyAttribute(path string, enable bool) error {
	return nil
}

// Set System attribute of a file or folder. Does nothing on Unix-like system.
func SetSystemAttribute(path string, enable bool) error {
	return nil
}

// Set Created time, Access time, Modified time of a file or folder.
func SetTime(path string, ctime, atime, mtime time.Time) error {
	return os.Chtimes(path, atime, mtime)
}
