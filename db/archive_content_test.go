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

func TestArchiveContentPathHash(t *testing.T) {
	tests := []struct {
		name      string
		dir1      string
		file1     string
		ext1      string
		dir2      string
		file2     string
		ext2      string
		wantEqual bool
	}{
		{"same_path", "docs/reports", "annual", ".pdf", "docs/reports", "annual", ".pdf", true},
		{"different_extension", "docs", "file", ".pdf", "docs", "file", ".txt", false},
		{"different_name", "docs", "file1", ".pdf", "docs", "file2", ".pdf", false},
		{"different_directory", "docs/a", "file", ".pdf", "docs/b", "file", ".pdf", false},
		{"no_extension", "docs", "README", "", "docs", "README", "", true},
		{"extension_vs_no_extension", "docs", "file", "", "docs", "file", ".txt", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := archiveContentPathHash(tt.dir1, tt.file1, tt.ext1)
			h2 := archiveContentPathHash(tt.dir2, tt.file2, tt.ext2)
			if tt.wantEqual && h1 != h2 {
				t.Errorf("expected equal hashes but got different: %x vs %x", h1, h2)
			}
			if !tt.wantEqual && h1 == h2 {
				t.Errorf("expected different hashes but got equal: %x", h1)
			}
		})
	}
}

func TestSaveArchiveContents(t *testing.T) {
	docs, photos := testingArchiveContentData()
	tests := []struct {
		group    string
		contents []*testingArchiveContent
		expected int64
	}{
		{"new_archive_content", []*testingArchiveContent{docs[5], docs[6]}, 19},
		{"duplicated_archive_content", []*testingArchiveContent{docs[0], docs[1]}, 17},
		{"duplicated_archive_content", []*testingArchiveContent{photos[9], photos[10], photos[15]}, 18},
		{"same_archive_content", []*testingArchiveContent{photos[9], photos[15], photos[15]}, 18},
	}
	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			ctx := getAndReseedArchiveContentDB("SaveArchiveContents")
			var acs []*ArchiveContent
			for _, a := range tt.contents {
				acs = append(acs, a.ArchiveContent())
			}
			ctx.SaveArchiveContents(acs)
			count, err := ctx.Count(&ArchiveContent{}, nil, nil)
			if err != nil {
				t.Error(err)
			}
			if count != tt.expected {
				t.Errorf("wrong number of records. expected %v actual %v", tt.expected, count)
			}
		})
	}
}

type testingArchiveContent struct {
	ArchiveID uuid.UUID
	Directory string
	Name      string
	Extension string
	HashID    uuid.UUID
}

func (a *testingArchiveContent) ArchiveContent() *ArchiveContent {
	return NewArchiveContent(a.ArchiveID, a.Directory, a.Name, a.Extension, a.HashID)
}

func testingArchiveContentData() (docs, photos []*testingArchiveContent) {
	docsArchiveID := uuid.MustParse("0195b2a1-0001-7000-8000-000000000001")
	photosArchiveID := uuid.MustParse("0195b2a1-0001-7000-8000-000000000003")

	docs = []*testingArchiveContent{
		{docsArchiveID, "documents/project", "report", ".pdf", uuid.MustParse("0195b2a1-0002-7000-8000-000000000001")},
		{docsArchiveID, "documents/project", "budget", ".xlsx", uuid.MustParse("0195b2a1-0002-7000-8000-000000000002")},
		{docsArchiveID, "documents/project", "plan", ".docx", uuid.MustParse("0195b2a1-0002-7000-8000-000000000003")},
		{docsArchiveID, "documents/personal", "cv", ".pdf", uuid.MustParse("0195b2a1-0002-7000-8000-000000000004")},
		{docsArchiveID, "documents/personal", "notes", ".txt", uuid.MustParse("0195b2a1-0002-7000-8000-000000000005")},
		{docsArchiveID, "documents/work", "contract", ".pdf", uuid.MustParse("0195b2a1-0002-7000-8000-000000000006")},
		{docsArchiveID, "documents/work", "invoice", ".docx", uuid.MustParse("0195b2a1-0002-7000-8000-000000000007")},
	}

	photos = []*testingArchiveContent{
		{photosArchiveID, "photos/2023/beach", "img001", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000001")},
		{photosArchiveID, "photos/2023/beach", "img002", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000002")},
		{photosArchiveID, "photos/2023/beach", "img003", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000003")},
		{photosArchiveID, "photos/2023/mountain", "img001", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000004")},
		{photosArchiveID, "photos/2023/mountain", "img002", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000005")},
		{photosArchiveID, "photos/2023/mountain", "img003", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000006")},
		{photosArchiveID, "photos/2023/mountain", "img004", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000007")},
		{photosArchiveID, "photos/2024/summer", "img001", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000008")},
		{photosArchiveID, "photos/2024/summer", "img002", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000009")},
		{photosArchiveID, "photos/2024/summer", "img003", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-00000000000a")},
		{photosArchiveID, "photos/2024/summer", "img004", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-00000000000b")},
		{photosArchiveID, "photos/2024/winter", "img001", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-00000000000c")},
		{photosArchiveID, "photos/2024/winter", "img002", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-00000000000d")},
		{photosArchiveID, "photos/2024/winter", "img003", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-00000000000e")},
		{photosArchiveID, "photos/2024/winter", "img004", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-00000000000f")},
		{photosArchiveID, "photos/2024/winter", "img005", ".jpg", uuid.MustParse("0195b2a1-0003-7000-8000-000000000010")},
	}
	return
}

func getAndReseedArchiveContentDB(function string) *DbContext {
	ctx := getTestDB("ArchiveContent", function)
	docs, photos := testingArchiveContentData()
	var seed []*testingArchiveContent
	seed = append(seed, docs[0:5]...)
	seed = append(seed, photos[0:12]...)
	tx := ctx.db.Begin()
	tx.Where("1 = 1").Delete(&ArchiveContent{})
	for _, m := range seed {
		tx.Create(m.ArchiveContent())
	}
	tx.Commit()
	return ctx
}
