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
	"fmt"
	"io"
	"regexp"
	"strings"
)

type parser struct {
	s   *scanner
	buf struct {
		tok token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

func newParser(r io.Reader) *parser {
	return &parser{s: NewScanner(r)}
}

func (p *parser) Parse() ([]*ChecksumItem, error) {
	items := []*ChecksumItem{}

	for {
		item := &ChecksumItem{}
		if tok, lit := p.scan(); tok == EOF {
			break
		} else if tok == WORD {
			item.Hash = lit
		} else {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "hash", lit)
		}

		if tok, lit := p.scan(); tok != SPACE {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "whitespace", lit)
		}

		pathSlice := []string{}
		if tok, lit := p.scan(); tok == ASTERISK {
			item.BinaryMode = true
		} else if tok == WORD || tok == SPACE {
			p.unscan()
		} else {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "whitespace", lit)
		}

		if tok, lit := p.scan(); tok == SPACE {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
		} else {
			p.unscan()
		}

		var lastTok token
		for {
			if tok, lit := p.scan(); tok == SPACE || tok == WORD {
				lastTok = tok
				pathSlice = append(pathSlice, lit)
			} else if tok == CR || tok == LF || tok == EOF {
				p.unscan()
				item.Path = strings.Join(pathSlice, "")
				break
			} else {
				return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
			}
		}
		if len(pathSlice) == 0 {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", p.buf.lit)
		}
		if lastTok == SPACE {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", " ")
		}

		if tok, lit := p.scan(); tok == CR {
			if tok, lit = p.scan(); tok == LF {
				items = append(items, item)
			} else {
				return []*ChecksumItem{}, fmt.Errorf("invalid token. expcted %s actual '%s'", "endline", lit)
			}
		} else if tok == LF {
			items = append(items, item)
		} else if tok == EOF {
			items = append(items, item)
			break
		} else {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expcted %s actual '%s'", "endline", lit)
		}
	}

	return items, nil
}

func (p *parser) scan() (token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	tok, lit := p.s.Scan()
	p.buf.tok, p.buf.lit = tok, lit
	return tok, lit
}

func (p *parser) unscan() {
	p.buf.n = 1
}

func ParseSha256(r io.Reader) ([]*ChecksumItem, error) {
	parser := newParser(r)
	items, err := parser.Parse()
	if err != nil {
		return items, err
	}
	hashCheckRegex := regexp.MustCompile("^[0-9A-Fa-f]{64}$")
	for _, l := range items {
		if !hashCheckRegex.MatchString(l.Hash) {
			return []*ChecksumItem{}, fmt.Errorf("invalid SHA-256 hash '%s'", l.Hash)
		}
	}
	return items, err
}
