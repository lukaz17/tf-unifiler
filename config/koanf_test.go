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

package config

import (
	"testing"

	"github.com/tforceaio/tf-unifiler-go/filesys"
)

func TestFileConfig(t *testing.T) {
	prepareTests()

	cfg, err := BuildConfig(true, "../../.tests/config/unifiler.yml")
	if err != nil {
		t.Error("Error get config from file", err)
	}
	if cfg.Path.FFMpegPath != "/usr/bin/ffmpeg" {
		t.Errorf("Wrong FFMpegPath. Expected '%s' Actual '%s'", "/usr/bin/ffmpeg", cfg.Path.FFMpegPath)
	}
	if cfg.Path.ImageMagickPath != "magick" {
		t.Errorf("Wrong ImageMagickPath. Expected '%s' Actual '%s'", "magick", cfg.Path.ImageMagickPath)
	}
	if cfg.Path.MediaInfoPath != "/opt/mediainfo/bin/mediainfo" {
		t.Errorf("Wrong MediaInfoPath. Expected '%s' Actual '%s'", "/opt/mediainfo/bin/mediainfo", cfg.Path.MediaInfoPath)
	}
	if cfg.Path.X264Path != "x264" {
		t.Errorf("Wrong X264Path. Expected '%s' Actual '%s'", "x264", cfg.Path.X264Path)
	}
	if cfg.Path.X265Path != "/usr/bin/x265" {
		t.Errorf("Wrong X265Path. Expected '%s' Actual '%s'", "/usr/bin/x265", cfg.Path.X265Path)
	}
}

func prepareTests() {
	contents := []string{
		"paths:",
		"  ffmpeg: /usr/bin/ffmpeg",
		"  mediainfo: /opt/mediainfo/bin/mediainfo",
		"  x265: /usr/bin/x265",
	}

	if !filesys.IsExist("../../.tests/config") {
		filesys.CreateDirectoryRecursive("../../.tests/config")
	}
	if !filesys.IsExist("../../.tests/config/unifiler.yml") {
		filesys.WriteLines("../../.tests/config/unifiler.yml", contents)
	}
}
