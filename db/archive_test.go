// Copyright (C) 2025 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"testing"

	"github.com/google/uuid"
)

func TestSaveArchives(t *testing.T) {
	archives := testingArchiveData()
	tests := []struct {
		group    string
		archives []*testingArchive
		expected int64
	}{
		{"new_archive", []*testingArchive{archives[2]}, 3},
		{"duplicated_archive_name", []*testingArchive{archives[0], archives[1]}, 2},
		{"duplicated_archive_name", []*testingArchive{archives[1], archives[2], archives[3]}, 4},
		{"same_name", []*testingArchive{archives[3], archives[3], archives[3]}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			ctx := getAndReseedArchiveDB("SaveArchives")
			var archives []*Archive
			for _, a := range tt.archives {
				archives = append(archives, a.Archive())
			}
			ctx.SaveArchives(archives)
			count, err := ctx.Count(&Archive{}, nil, nil)
			if err != nil {
				t.Error(err)
			}
			if count != tt.expected {
				t.Errorf("wrong number of records. expected %v actual %v", tt.expected, count)
			}
		})
	}
}

type testingArchive struct {
	ID   uuid.UUID
	Name string
}

func (a *testingArchive) Archive() *Archive {
	return &Archive{
		ID:   a.ID,
		Name: a.Name,
	}
}

func testingArchiveData() []*testingArchive {
	return []*testingArchive{
		{uuid.MustParse("0195b2a1-0001-7000-8000-000000000001"), "backups"},
		{uuid.MustParse("0195b2a1-0001-7000-8000-000000000002"), "documents"},
		{uuid.MustParse("0195b2a1-0001-7000-8000-000000000003"), "photos"},
		{uuid.MustParse("0195b2a1-0001-7000-8000-000000000004"), "videos"},
	}
}

func getAndReseedArchiveDB(function string) *DbContext {
	ctx := getTestDB("Archive", function)
	archives := testingArchiveData()
	tx := ctx.db.Begin()
	tx.Where("1 = 1").Delete(&Archive{})
	for _, a := range archives[0:2] {
		tx.Create(a.Archive())
	}
	tx.Commit()
	return ctx
}
