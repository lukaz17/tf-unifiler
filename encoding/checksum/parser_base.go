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

// parserBase holds the shared scanner and one-token look-ahead buffer used by
// all checksum format parsers.
type parserBase struct {
	s   *scanner
	buf struct {
		tok token
		lit string
		n   int
	}
}

// scan returns the next token, replaying the buffered token when unscan was
// called previously.
func (p *parserBase) scan() (token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	tok, lit := p.s.Scan()
	p.buf.tok, p.buf.lit = tok, lit
	return tok, lit
}

// unscan pushes the most recently scanned token back into the buffer so that
// the next call to scan returns it again.
func (p *parserBase) unscan() {
	p.buf.n = 1
}
