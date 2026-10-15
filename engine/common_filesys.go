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

package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/tforceaio/tf-unifiler/crypto/hasher"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/tui"
)

// fileHashResult pairs a filesystem entry with its computed hashes.
type fileHashResult struct {
	Entry  *filesys.FsEntry
	Hashes []*hasher.HashResult
}

// List and hash for all files using the specified algorithms.
func listAndHashFiles(inputs []string, algorithms []string, recursive bool, notifier diag.Notifier) ([]*fileHashResult, error) {
	contents, err := filesys.List(inputs, recursive)
	if err != nil {
		return nil, err
	}

	if n, ok := notifier.(*tui.BubbleteaNotifier); ok {
		ps := tui.RunProcessStatus(n)
		defer ps.Stop()
		total := uint64(len(contents))
		for _, c := range contents {
			if c.IsDir {
				total--
			}
		}
		n.SetTotal(total)
	}

	var results []*fileHashResult
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		hashes, err := hasher.Hash(c.RelativePath, algorithms)
		if err != nil {
			return nil, fmt.Errorf("failed to hash %s: %w", c.RelativePath, err)
		}
		results = append(results, &fileHashResult{
			Entry:  c,
			Hashes: hashes,
		})
	}
	return results, nil
}

// Set file attributes that satisfy the following conditions:
// - Archive attribute is set to true.
// - Hidden attribute is set to false.
// - Read-only attribute is set to false.
// - Access time, Creation time, and Modified time are all set to February 2, 2020, 12:00:00 UTC.
func normalizeAttributes(path string) error {
	if err := filesys.SetAttribute(path, 1, -1, -1, -1); err != nil {
		return err
	}
	time := time.Date(2020, 2, 2, 12, 0, 0, 0, time.UTC)
	if err := filesys.SetTime(path, time, time, time); err != nil {
		return err
	}

	return nil
}

// Write a JSON file with the following path pattern: <dir>/<prefix><timestamp>.json
func writeJSON(dir, prefix string, data any) (string, error) {
	currentTimestamp := time.Now().UnixMilli()
	filePath := filesys.Join(dir, prefix+strconv.FormatInt(currentTimestamp, 10)+".json")
	content, err := json.Marshal(data)
	if err != nil {
		return filePath, err
	}
	err = filesys.WriteLines(filePath, []string{string(content)})
	return filePath, err
}
