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

package engine

import (
	"testing"
)

func TestParseVideoFlagsOfExtractFrames(t *testing.T) {
	cmd, _, err := VideoCmd().Find([]string{"extract-frames"})
	if err != nil {
		t.Fatalf("finding extract-frames: %v", err)
	}
	if err := cmd.ParseFlags([]string{"--interval", "30", "--offset", "60", "--limit", "90.5"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}
	flags := ParseVideoFlags(cmd, []string{"movie.mkv"})
	if flags.Interval != 30 || flags.Offset != 60 || flags.Limit != 90.5 {
		t.Errorf("interval/offset/limit = %v/%v/%v, want 30/60/90.5", flags.Interval, flags.Offset, flags.Limit)
	}
	if flags.Quality != 1 {
		t.Errorf("default quality = %d, want 1, FFmpeg's best JPEG quality", flags.Quality)
	}
}
