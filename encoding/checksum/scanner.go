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
	"bufio"
	"bytes"
	"io"
)

var eof = rune(0)

type scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) Scan() (tok token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isWord(ch) {
		s.unread()
		return s.scanWord()
	}

	switch ch {
	case '*':
		return ASTERISK, "*"
	case '\r':
		return CR, "\r"
	case '\n':
		return LF, "\n"
	case eof:
		return EOF, string(eof)
	}

	return INVALID, string(ch)
}

func (s *scanner) scanWhitespace() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); isWhitespace(ch) {
			buf.WriteRune(ch)
		} else {
			s.unread()
			break
		}
	}

	return SPACE, buf.String()
}

func (s *scanner) scanWord() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); isWord(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread()
			break
		}
	}

	return WORD, buf.String()
}

func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *scanner) unread() {
	_ = s.r.UnreadRune()
}

func isAsterisk(ch rune) bool {
	return ch == '*'
}

func isEndline(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

func isEndOfFile(ch rune) bool {
	return ch == eof
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func isWord(ch rune) bool {
	return !isAsterisk(ch) && !isEndline(ch) && !isEndOfFile(ch) && !isWhitespace(ch)
}
