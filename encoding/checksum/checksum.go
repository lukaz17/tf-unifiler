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
	"regexp"
	"strings"
)

type checksumParser struct {
	parserBase
}

func newChecksumParser(r io.Reader) *checksumParser {
	p := &checksumParser{}
	p.s = newScanner(r)
	return p
}

func (p *checksumParser) Parse() ([]*ChecksumItem, error) {
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
		} else if tok == WORD {
			item.Hash = lit
		} else {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "hash", lit)
		}

		if tok, lit := p.scan(); tok != SPACE {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "whitespace", lit)
		}

		if tok, lit := p.scan(); tok == ASTERISK {
			item.BinaryMode = true
		} else if tok == WORD || tok == SPACE {
			p.unscan()
		} else {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "whitespace", lit)
		}

		if tok, lit := p.scan(); tok == SPACE {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
		} else {
			p.unscan()
		}

		pathSlice := []string{}
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
				return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
			}
		}
		if len(pathSlice) == 0 {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", p.buf.lit)
		}
		if lastTok == SPACE {
			return nil, fmt.Errorf("invalid token. expected %s actual '%s'", "path", " ")
		}

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

// Parse a md5 checksum file
func ParseMd5(r io.Reader) ([]*ChecksumItem, error) {
	return parseWithHashLength(newChecksumParser(r).Parse, "MD5", 32)
}

// Parse a sha1 checksum file
func ParseSha1(r io.Reader) ([]*ChecksumItem, error) {
	return parseWithHashLength(newChecksumParser(r).Parse, "SHA-1", 40)
}

// Parse a sha256 checksum file
func ParseSha256(r io.Reader) ([]*ChecksumItem, error) {
	return parseWithHashLength(newChecksumParser(r).Parse, "SHA-256", 64)
}

// Parse a sha512 checksum file
func ParseSha512(r io.Reader) ([]*ChecksumItem, error) {
	return parseWithHashLength(newChecksumParser(r).Parse, "SHA-512", 128)
}

func parseWithHashLength(parse func() ([]*ChecksumItem, error), algo string, length int) ([]*ChecksumItem, error) {
	items, err := parse()
	if err != nil {
		return items, err
	}
	pattern := fmt.Sprintf("^[0-9A-Fa-f]{%d}$", length)
	hashCheckRegex := regexp.MustCompile(pattern)
	for _, l := range items {
		if !hashCheckRegex.MatchString(l.Hash) {
			return nil, fmt.Errorf("invalid %s hash '%s'", algo, l.Hash)
		}
	}
	return items, nil
}
