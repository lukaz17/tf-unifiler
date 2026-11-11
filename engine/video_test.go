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
	"bytes"
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/tforceaio/tf-unifiler/config"
)

func TestExtractFramesSeeksToTimeInFileName(t *testing.T) {
	t.Setenv(fakeVideoToolsEnv, "1")
	tests := []struct {
		name     string
		offset   string
		wantFile string
		wantSeek string
	}{
		{"fractional second", "1.5", "movie.mkv_0_00_01_500.jpg", "1.500"},
		{"whole second", "2", "movie.mkv_0_00_02_000.jpg", "2.000"},
		{"under a second", "0.25", "movie.mkv_0_00_00_250.jpg", "0.250"},
		{"millisecond inexact in binary", "1.001", "movie.mkv_0_00_01_001.jpg", "1.001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			video := filepath.Join(dir, "movie.mkv")
			if err := os.WriteFile(video, nil, 0644); err != nil {
				t.Fatalf("creating video: %v", err)
			}
			framesDir := filepath.Join(dir, "frames")
			base := &Controller{Config: &config.RootConfig{Path: &config.PathConfig{FFMpegPath: os.Args[0], MediaInfoPath: os.Args[0]}}}
			// --limit at --offset takes exactly one frame.
			args := []string{"video", "extract-frames", video, "-o", framesDir, "-q", "1", "--offset", tt.offset, "--limit", tt.offset}

			var out bytes.Buffer
			err := RunTask(base, args, context.Background(), &out)

			if err != nil {
				t.Fatalf("RunTask(%v): %v\n%s", args, err, out.String())
			}
			seek, err := os.ReadFile(filepath.Join(framesDir, tt.wantFile))
			if err != nil {
				t.Fatalf("reading frame: %v", err)
			}
			if string(seek) != tt.wantSeek {
				t.Errorf("frame %s was taken with -ss %s, want -ss %s", tt.wantFile, seek, tt.wantSeek)
			}
		})
	}
}

func TestConvertSecondToMsRoundsToNearest(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		wantMs  int64
	}{
		{"one millisecond past", 1.001, 1001},
		{"two milliseconds past", 2.002, 2002},
		{"under half a millisecond past", 1.0004, 1000},
		{"whole seconds", 2, 2000},
	}
	m := &VideoModule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := m.convertSecondToMs(tt.seconds)

			if ms.Int64() != tt.wantMs {
				t.Errorf("convertSecondToMs(%v) = %v ms, want %d ms", tt.seconds, ms, tt.wantMs)
			}
		})
	}
}

func TestDefaultScreenshotParameterByVideoLength(t *testing.T) {
	tests := []struct {
		name         string
		lengthMs     int64
		wantOffset   int64
		wantInterval int64
	}{
		{"empty video", 0, 1000, 2500},
		{"one minute", 60000, 1000, 2500},
		{"just under two minutes", 119999, 1000, 2500},
		{"two minutes", 120000, 1300, 4300},
		{"three hours", 10800000, 3400, 17100},
	}
	m := &VideoModule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, interval := m.defaultScreenshotParameter(big.NewInt(tt.lengthMs))

			if offset.Int64() != tt.wantOffset || interval.Int64() != tt.wantInterval {
				t.Errorf("offset/interval = %v/%v ms, want %d/%d ms", offset, interval, tt.wantOffset, tt.wantInterval)
			}
		})
	}
}

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
	if flags.Quality != 3 {
		t.Errorf("default quality = %d, want 3", flags.Quality)
	}
}
