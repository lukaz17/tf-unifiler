// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

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
		n.SetTotal(uint64(len(contents)))
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
