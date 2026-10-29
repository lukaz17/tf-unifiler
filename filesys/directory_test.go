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
	"path/filepath"
	"reflect"
	"testing"
)

func TestListEntries(t *testing.T) {
	prepareTests()

	tests := []struct {
		name     string
		files    []string
		maxDepth int
		results  []string
	}{
		{"multiple entries", []string{"../.tests/module/http", "../.tests/module/fmt"}, -1, []string{
			filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/http/1-get.md"), filepath.FromSlash("../.tests/module/http/2-post.md"), filepath.FromSlash("../.tests/module/http/3-put.md"), filepath.FromSlash("../.tests/module/http/4-delete.md"),
			filepath.FromSlash("../.tests/module/fmt"), filepath.FromSlash("../.tests/module/fmt/1-printf.md"), filepath.FromSlash("../.tests/module/fmt/2-errorf.md"),
		}},
		{"depth level 1", []string{"../.tests/module", "../.tests/basic"}, 1, []string{
			filepath.FromSlash("../.tests/module"), filepath.FromSlash("../.tests/module/fmt"), filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/readme.md"),
			filepath.FromSlash("../.tests/basic"), filepath.FromSlash("../.tests/basic/1-helloworld.md"),
		}},
		{"depth level 2", []string{"../.tests/module", "../.tests/basic"}, 2, []string{
			filepath.FromSlash("../.tests/module"), filepath.FromSlash("../.tests/module/fmt"), filepath.FromSlash("../.tests/module/fmt/1-printf.md"), filepath.FromSlash("../.tests/module/fmt/2-errorf.md"),
			filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/http/1-get.md"), filepath.FromSlash("../.tests/module/http/2-post.md"), filepath.FromSlash("../.tests/module/http/3-put.md"), filepath.FromSlash("../.tests/module/http/4-delete.md"),
			filepath.FromSlash("../.tests/module/readme.md"),
			filepath.FromSlash("../.tests/basic"), filepath.FromSlash("../.tests/basic/1-helloworld.md"),
		}},
		{"recursive", []string{"../.tests"}, -1, []string{
			filepath.FromSlash("../.tests"), filepath.FromSlash("../.tests/basic"), filepath.FromSlash("../.tests/basic/1-helloworld.md"),
			filepath.FromSlash("../.tests/module"), filepath.FromSlash("../.tests/module/fmt"), filepath.FromSlash("../.tests/module/fmt/1-printf.md"), filepath.FromSlash("../.tests/module/fmt/2-errorf.md"),
			filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/http/1-get.md"), filepath.FromSlash("../.tests/module/http/2-post.md"), filepath.FromSlash("../.tests/module/http/3-put.md"), filepath.FromSlash("../.tests/module/http/4-delete.md"),
			filepath.FromSlash("../.tests/module/readme.md"),
			filepath.FromSlash("../.tests/readme.md"),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := make([]*FsEntry, len(tt.files))
			for i, f := range tt.files {
				entries[i], _ = CreateEntry(f)
			}
			contents, err := listEntries(entries, tt.maxDepth, 0)
			fPaths := contents.GetPaths()
			if !reflect.DeepEqual(fPaths, tt.results) {
				t.Error(err)
				t.Errorf("Wrong file listing. Expected '%s' Actual '%s'", tt.results, fPaths)
			}
		})
	}
}
