// Copyright (C) 2024 T-Force I/O
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

package checksum

import (
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	var tests = []struct {
		name string
		s    string
		tok  token
		lit  string
	}{
		{"EOF", "", EOF, string(eof)},
		{"Whitespace", " xyz", SPACE, " "},
		{"Long Whitespace", "   xyz", SPACE, "   "},
		{"Tab", "\txyz", SPACE, "\t"},
		{"Multi Whitespace", " \t xyz", SPACE, " \t "},
		{"Carriage Return", "\r", CR, "\r"},
		{"Carriage Return", "\r\n", CR, "\r"},
		{"Line Feed", "\n", LF, "\n"},
		{"Asterisk", "****", ASTERISK, "*"},
		{"Percent", "!@#$%^&()\\/. qwert", WORD, "!@#$%^&()\\/."},
		{"ASCII", "abcdef0123456789 qwert", WORD, "abcdef0123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.s))
			tok, lit := s.Scan()
			if tt.tok != tok {
				t.Errorf("Token mismatch. Expected %d Acutual %d", tt.tok, tok)
			} else if tt.lit != lit {
				t.Errorf("Literal mismatch. Expected '%s' Actual '%s'", tt.lit, lit)
			}
		})
	}
}
