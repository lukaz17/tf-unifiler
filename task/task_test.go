// Copyright (C) 2025  T-Force I/O
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

package task

import (
	"strings"
	"testing"
)

func TestBoundedBufferKeepsLatestBytes(t *testing.T) {
	const marker = "... output truncated ...\n"
	tests := []struct {
		name   string
		limit  int
		writes []string
		want   string
	}{
		{"output under the limit is kept whole", 8, []string{"abc", "def"}, "abcdef"},
		{"output at the limit is kept whole", 6, []string{"abc", "def"}, "abcdef"},
		{"output crossing the limit keeps the latest bytes", 4, []string{"abc", "def", "g"}, marker + "defg"},
		{"one write past the limit keeps its end", 4, []string{"ab", "cdefgh"}, marker + "efgh"},
		{"zero limit keeps everything", 0, []string{strings.Repeat("x", 100)}, strings.Repeat("x", 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &boundedBuffer{limit: tt.limit}
			for _, w := range tt.writes {
				n, err := b.Write([]byte(w))
				if n != len(w) || err != nil {
					t.Fatalf("write %q: expected (%d, nil), got (%d, %v)", w, len(w), n, err)
				}
			}

			got := b.String()

			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
