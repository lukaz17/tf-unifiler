// Copyright (C) 2024  T-Force I/O
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

package filesys

import (
	"os"
	"path/filepath"
)

func CreateDirectory(dPath string) error {
	return os.Mkdir(dPath, 0755)
}

func CreateDirectoryRecursive(dPath string) error {
	return os.MkdirAll(dPath, 0755)
}

func IsDirectory(dPath string) (bool, error) {
	fileInfo, err := os.Lstat(dPath)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func IsDirectoryUnsafe(dPath string) bool {
	isDir, err := IsDirectory(dPath)
	if err != nil {
		panic(err)
	}
	return isDir
}

func IsDirectoryEmpty(dPath string) (bool, error) {
	entries, err := os.ReadDir(dPath)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func IsDirectoryExist(fPath string) bool {
	fileInfo, err := os.Stat(fPath)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func listDirectory(dPath string) (FsEntries, error) {
	entries, err := os.ReadDir(dPath)
	if err != nil {
		return FsEntries{}, err
	}
	contents := make(FsEntries, len(entries))
	for i, e := range entries {
		relativePath := filepath.Join(dPath, e.Name())
		absolutePath, err := GetAbsPath(relativePath)
		if err != nil {
			return FsEntries{}, err
		}
		content := &FsEntry{
			AbsolutePath: absolutePath,
			RelativePath: relativePath,
			Name:         e.Name(),
			IsDir:        e.IsDir(),
		}
		contents[i] = content
	}
	return contents, nil
}

func listEntries(entries []*FsEntry, maxDepth int, depth int) (FsEntries, error) {
	contents := FsEntries{}
	for _, e := range entries {
		contents = append(contents, e)
		if (depth >= maxDepth && maxDepth >= 0) || !e.IsDir {
			continue
		}
		subEntries, err := listDirectory(e.RelativePath)
		if err != nil {
			return FsEntries{}, err
		}
		subContents, err := listEntries(subEntries, maxDepth, depth+1)
		if err != nil {
			return FsEntries{}, err
		}
		contents = append(contents, subContents...)
	}
	return contents, nil
}
