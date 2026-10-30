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

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg, err := buildConfig(true, filepath.Join(t.TempDir(), "unifiler.yaml"))
	if err != nil {
		t.Error("Error get default config", err)
	}
	if cfg.Debug {
		t.Error("Wrong Debug. Expected default 'false' Actual 'true'")
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("Wrong Logger.Level. Expected '%s' Actual '%s'", "info", cfg.Logger.Level)
	}
	if cfg.Path.FFMpegPath != "ffmpeg" {
		t.Errorf("Wrong FFMpegPath. Expected '%s' Actual '%s'", "ffmpeg", cfg.Path.FFMpegPath)
	}
	if cfg.Path.ImageMagickPath != "magick" {
		t.Errorf("Wrong ImageMagickPath. Expected '%s' Actual '%s'", "magick", cfg.Path.ImageMagickPath)
	}
	if !cfg.Server.Autostart {
		t.Error("Wrong Server.Autostart. Expected default 'true' Actual 'false'")
	}
}

func TestFileConfig(t *testing.T) {
	contents := []string{
		"debug: true",
		"paths:",
		"  ffmpeg: /usr/bin/ffmpeg",
		"  imagemagick: magick",
		"  mediainfo: /opt/mediainfo/bin/mediainfo",
		"  x264: x264",
		"  x265: /usr/bin/x265",
		"log:",
		"  level: debug",
	}
	f := filepath.Join(t.TempDir(), "unifiler.yaml")
	if err := os.WriteFile(f, []byte(strings.Join(contents, "\n")), 0664); err != nil {
		t.Fatal(err)
	}

	cfg, err := buildConfig(true, f)
	if err != nil {
		t.Error("Error get config from file", err)
	}
	if !cfg.Debug {
		t.Error("Wrong Debug. Expected 'true' Actual 'false'")
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
	if cfg.Logger.Level != "debug" {
		t.Errorf("Wrong Logger.Level. Expected '%s' Actual '%s'", "debug", cfg.Logger.Level)
	}
}

func TestEnvConfig(t *testing.T) {
	t.Setenv("TFUNIFILER_DEBUG", "true")
	t.Setenv("TFUNIFILER_LOG_LEVEL", "warn")
	cfg, err := buildConfig(true, filepath.Join(t.TempDir(), "unifiler.yaml"))
	if err != nil {
		t.Error("Error get config from env", err)
	}
	if !cfg.Debug {
		t.Error("Wrong Debug. Expected 'true' Actual 'false'")
	}
	if cfg.Logger.Level != "warn" {
		t.Errorf("Wrong Logger.Level. Expected '%s' Actual '%s'", "warn", cfg.Logger.Level)
	}
}
