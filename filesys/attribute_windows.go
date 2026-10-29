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

package filesys

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

// Set multiple attributes (Archive, ReadOnly, Hidden, System) of a file or folder at once.
// archive, readOnly, hidden, system: +1 to enable, -1 to disable, 0 to keep existing value.
func SetAttribute(path string, archive, readOnly, hidden, system int) error {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return err
	}

	applyAttr := func(flag uint32, val int) {
		if val > 0 {
			attrs |= flag
		} else if val < 0 {
			attrs &^= flag
		}
	}

	applyAttr(windows.FILE_ATTRIBUTE_ARCHIVE, archive)
	applyAttr(windows.FILE_ATTRIBUTE_READONLY, readOnly)
	applyAttr(windows.FILE_ATTRIBUTE_HIDDEN, hidden)
	applyAttr(windows.FILE_ATTRIBUTE_SYSTEM, system)

	return windows.SetFileAttributes(ptr, attrs)
}

// Set Archive attribute of a file or folder.
func SetArchiveAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_ARCHIVE, enable)
}

// Set Created time, Access time, Modified time of a file or folder.
func SetTime(path string, ctime, atime, mtime time.Time) error {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	handle, err := windows.CreateFile(
		pathPtr,
		windows.FILE_WRITE_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer windows.CloseHandle(handle)
	ctimew := windows.NsecToFiletime(ctime.UnixNano())
	atimew := windows.NsecToFiletime(atime.UnixNano())
	mtimew := windows.NsecToFiletime(mtime.UnixNano())
	return windows.SetFileTime(handle, &ctimew, &atimew, &mtimew)
}

// Set Created time of a file or folder.
func SetCreatedTime(path string, ctime time.Time) error {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	handle, err := windows.CreateFile(
		pathPtr,
		windows.FILE_WRITE_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer windows.CloseHandle(handle)
	var ctimew, atimew, mtimew windows.Filetime
	err = windows.GetFileTime(handle, &ctimew, &atimew, &mtimew)
	if err != nil {
		return err
	}
	ctimew = windows.NsecToFiletime(ctime.UnixNano())
	return windows.SetFileTime(handle, &ctimew, &atimew, &mtimew)
}

// Set Hidden attribute of a file or folder.
func SetHiddenAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_HIDDEN, enable)
}

// Set Read Only attribute of a file or folder.
func SetReadOnlyAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_READONLY, enable)
}

// Set System attribute of a file or folder.
func SetSystemAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_SYSTEM, enable)
}

// Set attribute value for file or folder on Windows.
func setFileAttribute(path string, flag uint32, enable bool) error {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return err
	}
	if enable {
		attrs |= flag
	} else {
		attrs &^= flag
	}
	return windows.SetFileAttributes(ptr, attrs)
}
