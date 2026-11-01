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

func TestList(t *testing.T) {
	prepareTests()

	tests := []struct {
		name    string
		files   []string
		results []string
	}{
		{"files only", []string{"file.go"}, []string{"file.go"}},
		{"directories only", []string{"../.tests/module/http"}, []string{
			filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/http/1-get.md"), filepath.FromSlash("../.tests/module/http/2-post.md"), filepath.FromSlash("../.tests/module/http/3-put.md"), filepath.FromSlash("../.tests/module/http/4-delete.md"),
		}},
		{"file and directories", []string{"../.tests/module", "../.tests/readme.md"}, []string{
			filepath.FromSlash("../.tests/module"), filepath.FromSlash("../.tests/module/fmt"), filepath.FromSlash("../.tests/module/fmt/1-printf.md"), filepath.FromSlash("../.tests/module/fmt/2-errorf.md"),
			filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/http/1-get.md"), filepath.FromSlash("../.tests/module/http/2-post.md"), filepath.FromSlash("../.tests/module/http/3-put.md"), filepath.FromSlash("../.tests/module/http/4-delete.md"),
			filepath.FromSlash("../.tests/module/readme.md"),
			filepath.FromSlash("../.tests/readme.md"),
		}},
		{"two dots", []string{"../.tests"}, []string{
			filepath.FromSlash("../.tests"), filepath.FromSlash("../.tests/basic"), filepath.FromSlash("../.tests/basic/1-helloworld.md"),
			filepath.FromSlash("../.tests/module"), filepath.FromSlash("../.tests/module/fmt"), filepath.FromSlash("../.tests/module/fmt/1-printf.md"), filepath.FromSlash("../.tests/module/fmt/2-errorf.md"),
			filepath.FromSlash("../.tests/module/http"), filepath.FromSlash("../.tests/module/http/1-get.md"), filepath.FromSlash("../.tests/module/http/2-post.md"), filepath.FromSlash("../.tests/module/http/3-put.md"), filepath.FromSlash("../.tests/module/http/4-delete.md"),
			filepath.FromSlash("../.tests/module/readme.md"), filepath.FromSlash("../.tests/readme.md"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, err := List(tt.files, true)
			fPaths := contents.GetPaths()
			if !reflect.DeepEqual(fPaths, tt.results) {
				t.Error(err)
				t.Errorf("Wrong file listing. Expected '%s' Actual '%s'", tt.results, fPaths)
			}
		})
	}
}

func prepareTests() {
	dirs := []string{
		"../.tests",
		"../.tests/basic",
		"../.tests/module",
		"../.tests/module/fmt",
		"../.tests/module/http",
	}
	files := []string{
		"../.tests/readme.md",
		"../.tests/basic/1-helloworld.md",
		"../.tests/module/readme.md",
		"../.tests/module/fmt/1-printf.md",
		"../.tests/module/fmt/2-errorf.md",
		"../.tests/module/http/1-get.md",
		"../.tests/module/http/2-post.md",
		"../.tests/module/http/3-put.md",
		"../.tests/module/http/4-delete.md",
	}
	for _, d := range dirs {
		if !IsExist(d) {
			CreateDirectory(d)
		}
	}
	for _, f := range files {
		if !IsExist(f) {
			WriteLines(f, []string{})
		}
	}
}
