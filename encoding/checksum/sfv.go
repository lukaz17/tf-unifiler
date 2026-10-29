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
	"fmt"
	"io"
	"strings"
)

type sfvParser struct {
	parserBase
}

func newSfvParser(r io.Reader) *sfvParser {
	p := &sfvParser{}
	p.s = newScanner(r)
	return p
}

func (p *sfvParser) Parse() ([]*ChecksumItem, error) {
	items := []*ChecksumItem{}

	for {
		item := &ChecksumItem{}
		if tok, lit := p.scan(); tok == EOF {
			break
		} else if tok == LF {
			continue
		} else if tok == CR {
			if next, _ := p.scan(); next != LF {
				p.unscan()
			}
			continue
		} else if tok == SEMICOLON {
			for {
				t, _ := p.scan()
				if t == CR {
					if next, _ := p.scan(); next != LF {
						p.unscan()
					}
					break
				} else if t == LF || t == EOF {
					if t == EOF {
						p.unscan()
					}
					break
				}
			}
			continue
		} else if tok == WORD {
			p.unscan()
		} else {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
		}

		pathSlice := []string{}
		var lastTok token
		for {
			if tok, lit := p.scan(); tok == SPACE || tok == WORD {
				lastTok = tok
				pathSlice = append(pathSlice, lit)
			} else if tok == CR || tok == LF || tok == EOF {
				p.unscan()
				break
			} else {
				return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
			}
		}
		if len(pathSlice) == 0 {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", p.buf.lit)
		}
		if lastTok == SPACE {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", " ")
		}

		item.Hash = pathSlice[len(pathSlice)-1]
		pathSlice = pathSlice[:len(pathSlice)-1]
		for len(pathSlice) > 0 && (pathSlice[len(pathSlice)-1] == " " || pathSlice[len(pathSlice)-1] == "\t") {
			pathSlice = pathSlice[:len(pathSlice)-1]
		}
		if len(pathSlice) == 0 {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "hash", "")
		}
		item.Path = strings.Join(pathSlice, "")

		if tok, lit := p.scan(); tok == CR {
			if tok, lit = p.scan(); tok == LF {
				items = append(items, item)
			} else {
				return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "endline", lit)
			}
		} else if tok == LF {
			items = append(items, item)
		} else if tok == EOF {
			items = append(items, item)
			break
		} else {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "endline", lit)
		}
	}

	return items, nil
}

func ParseCRC32(r io.Reader) ([]*ChecksumItem, error) {
	return parseWithHashLength(newSfvParser(r).Parse, "CRC32", 8)
}
