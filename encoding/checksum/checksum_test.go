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

package checksum

import (
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler/xlib"
)

func TestParseMd5Error(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		err     string
	}{
		{"wrong hash size (too short)", "ca47868bca0d531a275f20e99eb04b go.mod", "invalid MD5 hash 'ca47868bca0d531a275f20e99eb04b'"},
		{"wrong hash size (too long)", "ca47868bca0d531a275f20e99eb04ba100 go.mod", "invalid MD5 hash 'ca47868bca0d531a275f20e99eb04ba100'"},
		{"invalid character", "ca47868bca0d531a275f20e99eb04bXY go.mod", "invalid MD5 hash 'ca47868bca0d531a275f20e99eb04bXY'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMd5(strings.NewReader(tt.content))
			errs := xlib.ErrString(err)
			if errs != tt.err {
				t.Errorf("wrong error. Expected %q. Actual %q.", tt.err, errs)
			}
		})
	}
}

func TestParseSha1Error(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		err     string
	}{
		{"wrong hash size (too short)", "b91806c5f58c62d69dda437d20ee86eed0044f go.mod", "invalid SHA-1 hash 'b91806c5f58c62d69dda437d20ee86eed0044f'"},
		{"wrong hash size (too long)", "b91806c5f58c62d69dda437d20ee86eed0044f50b7 go.mod", "invalid SHA-1 hash 'b91806c5f58c62d69dda437d20ee86eed0044f50b7'"},
		{"invalid character", "b91806c5f58c62d69dda437d20ee86eed0044fXY go.mod", "invalid SHA-1 hash 'b91806c5f58c62d69dda437d20ee86eed0044fXY'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSha1(strings.NewReader(tt.content))
			errs := xlib.ErrString(err)
			if errs != tt.err {
				t.Errorf("wrong error. Expected %q. Actual %q.", tt.err, errs)
			}
		})
	}
}

func TestParseSha512Error(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		err     string
	}{
		{"wrong hash size (sha256 length)", "b91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bca go.mod", "invalid SHA-512 hash 'b91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bca'"},
		{"invalid character", "b91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bcab91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bXY go.mod", "invalid SHA-512 hash 'b91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bcab91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bXY'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSha512(strings.NewReader(tt.content))
			errs := xlib.ErrString(err)
			if errs != tt.err {
				t.Errorf("wrong error. Expected %q. Actual %q.", tt.err, errs)
			}
		})
	}
}

func TestParseSha256Error(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		err     string
	}{
		{"wrong hash size (too short)", "ca47868bca0d531a275f20e99eb04ba1 go.mod", "invalid SHA-256 hash 'ca47868bca0d531a275f20e99eb04ba1'"},
		{"wrong hash size (too long)", "eb3ff11bb7ec6a6cfc5f0ffb391af473628a443b95f5188d9b147ed9bd9862b36d8c22d796b185d16422b7be9663b4a9b9f68f6dfda1f189c9b8343492acb2a3 go.mod", "invalid SHA-256 hash 'eb3ff11bb7ec6a6cfc5f0ffb391af473628a443b95f5188d9b147ed9bd9862b36d8c22d796b185d16422b7be9663b4a9b9f68f6dfda1f189c9b8343492acb2a3'"},
		{"space in hash", "b91806c5f58c62d69dda437d20ee 86eed0044f50b780a7795d5014ed04ad1bca go.mod", "invalid SHA-256 hash 'b91806c5f58c62d69dda437d20ee'"},
		{"invalid character", "b91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bXY go.mod", "invalid SHA-256 hash 'b91806c5f58c62d69dda437d20ee86eed0044f50b780a7795d5014ed04ad1bXY'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSha256(strings.NewReader(tt.content))
			errs := xlib.ErrString(err)
			if errs != tt.err {
				t.Errorf("wrong error. Expected %q. Actual %q.", tt.err, errs)
			}
		})
	}
}

func TestParserItemCount(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		count   int
	}{
		{"0 item", "", 0},
		{"1 item", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod", 1},
		{"2 items", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\nca47868bca0d531a275f20e99eb04ba1 go.sum", 2},
		{"CRLF", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\r\nca47868bca0d531a275f20e99eb04ba1 go.sum", 2},
		{"Trailing CRLF", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\r\nca47868bca0d531a275f20e99eb04ba1 go.sum\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newChecksumParser(strings.NewReader(tt.content))
			items, _ := p.Parse()
			if len(items) != tt.count {
				t.Errorf("Invalid number of items. Expected %d. Actual %d.", tt.count, len(items))
			}
		})
	}
}

func TestParserSyntax(t *testing.T) {
	var tests = []struct {
		name       string
		content    string
		path       string
		binaryMode bool
		hash       string
	}{
		{"Text mode", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod", "go.mod", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Binary mode", "a3c51dd48bf7fabbbd354bd4e16b0ec1 *go.mod", "go.mod", true, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Multiple space", "a3c51dd48bf7fabbbd354bd4e16b0ec1    go.mod", "go.mod", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Path with multiple segment", "a3c51dd48bf7fabbbd354bd4e16b0ec1 filesystem/file.go", "filesystem/file.go", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Path with space #1", "a3c51dd48bf7fabbbd354bd4e16b0ec1 file system/file.go", "file system/file.go", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Path with space #2", "a3c51dd48bf7fabbbd354bd4e16b0ec1 *file system/file.go", "file system/file.go", true, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newChecksumParser(strings.NewReader(tt.content))
			items, _ := p.Parse()
			item := items[0]
			if item.Path != tt.path {
				t.Errorf("Wrong path. Expected '%s'. Actual '%s'.", tt.path, item.Path)
			}
			if item.BinaryMode != tt.binaryMode {
				t.Errorf("Wrong mode. Expected %t. Actual %t.", tt.binaryMode, item.BinaryMode)
			}
			if item.Hash != tt.hash {
				t.Errorf("Wrong hash. Expected '%s'. Actual '%s'.", tt.hash, item.Hash)
			}
		})
	}
}

func TestParserError(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		err     string
	}{
		{"hash only (no newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1", "invalid token. expected whitespace actual '\x00'"},
		{"hash only (with newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1\n", "invalid token. expected whitespace actual '\n'"},
		{"missing path (double space, EOF)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  ", "invalid token. expected whitespace actual '\x00'"},
		{"missing path (double space, newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  \n", "invalid token. expected whitespace actual '\n'"},
		{"missing path (asterisk, EOF)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *", "invalid token. expected path actual '\x00'"},
		{"missing path (asterisk, newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *\n", "invalid token. expected path actual '\n'"},
		{"path only (no newline)", "go.mod", "invalid token. expected whitespace actual '\x00'"},
		{"path only (with newline)", "go.mod\n", "invalid token. expected whitespace actual '\n'"},
		{"missing hash (leading space)", " go.mod", "invalid token. expected hash actual ' '"},
		{"path only (asterisk, no newline)", "*go.mod", "invalid token. expected hash actual '*'"},
		{"path only (asterisk, with newline)", "*go.mod\n", "invalid token. expected hash actual '*'"},
		{"missing hash (leading space and asterisk)", " *go.mod\n", "invalid token. expected hash actual ' '"},
		{"space before path (no newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  * go.mod", "invalid token. expected path actual ' '"},
		{"space before path (with newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  * go.mod\n", "invalid token. expected path actual ' '"},
		{"space after path (no newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *go.mod ", "invalid token. expected path actual ' '"},
		{"space after path (with newline)", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *go.mod \n", "invalid token. expected path actual ' '"},
		{"space before hash (no newline)", " a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod", "invalid token. expected hash actual ' '"},
		{"space before hash (with newline)", " a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\n", "invalid token. expected hash actual ' '"},
		{"space in hash (text mode, EOF)", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 go.mod", ""},
		{"space in hash (text mode, newline)", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 go.mod\n", ""},
		{"space in hash (binary mode, EOF)", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 *go.mod", "invalid token. expected path actual '*'"},
		{"space in hash (binary mode, newline)", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 *go.mod\n", "invalid token. expected path actual '*'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newChecksumParser(strings.NewReader(tt.content))
			_, err := p.Parse()
			errs := xlib.ErrString(err)
			if errs != tt.err {
				t.Errorf("wrong error. Expected %q. Actual %q.", tt.err, errs)
			}
		})
	}
}
