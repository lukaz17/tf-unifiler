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
	"slices"
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
	if cfg.Server.HTTP.Hostname != "127.0.0.1" {
		t.Errorf("Wrong Server.HTTP.Hostname. Expected '%s' Actual '%s'", "127.0.0.1", cfg.Server.HTTP.Hostname)
	}
	if cfg.Server.APIKey != "" {
		t.Errorf("Wrong Server.APIKey. Expected '%s' Actual '%s'", "", cfg.Server.APIKey)
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
	t.Setenv("TFUNIFILER_SERVER_HTTP_HOSTNAME", "0.0.0.0")
	t.Setenv("TFUNIFILER_SERVER_APIKEY", "s3cret")
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
	if cfg.Server.HTTP.Hostname != "0.0.0.0" {
		t.Errorf("Wrong Server.HTTP.Hostname. Expected '%s' Actual '%s'", "0.0.0.0", cfg.Server.HTTP.Hostname)
	}
	if cfg.Server.APIKey != "s3cret" {
		t.Errorf("Wrong Server.APIKey. Expected '%s' Actual '%s'", "s3cret", cfg.Server.APIKey)
	}
	applied := []string{"TFUNIFILER_DEBUG", "TFUNIFILER_LOG_LEVEL", "TFUNIFILER_SERVER_APIKEY", "TFUNIFILER_SERVER_HTTP_HOSTNAME"}
	if !slices.Equal(cfg.AppliedEnvVars, applied) {
		t.Errorf("Wrong AppliedEnvVars. Expected %v Actual %v", applied, cfg.AppliedEnvVars)
	}
}

func TestFileConfigKeysIgnoreCase(t *testing.T) {
	f := writeConfigFile(t, "Debug: true\nLog:\n  Level: debug\nSERVER:\n  HTTP:\n    Port: 8080\n")
	cfg, err := buildConfig(true, f)
	if err != nil {
		t.Fatal("Error get config from file", err)
	}
	if !cfg.Debug {
		t.Error("Wrong Debug. Expected 'true' Actual 'false'")
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("Wrong Logger.Level. Expected '%s' Actual '%s'", "debug", cfg.Logger.Level)
	}
	if cfg.Server.HTTP.Port != 8080 {
		t.Errorf("Wrong Server.HTTP.Port. Expected '%d' Actual '%d'", 8080, cfg.Server.HTTP.Port)
	}
	if cfg.Server.HTTP.Hostname != "127.0.0.1" {
		t.Errorf("Wrong Server.HTTP.Hostname. Expected default '%s' Actual '%s'", "127.0.0.1", cfg.Server.HTTP.Hostname)
	}
}

func TestFileConfigEmptySectionKeepsDefaults(t *testing.T) {
	f := writeConfigFile(t, "server:\nlog:\n  level:\n")
	cfg, err := buildConfig(true, f)
	if err != nil {
		t.Fatal("Error get config from file", err)
	}
	if cfg.Server == nil || cfg.Server.HTTP.Port != 17017 {
		t.Errorf("Wrong Server. Expected defaults Actual %+v", cfg.Server)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("Wrong Logger.Level. Expected default '%s' Actual '%s'", "info", cfg.Logger.Level)
	}
}

func TestFileConfigErrors(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{"malformed YAML", "log:\n  level: debug\n bad: [\n", "yaml"},
		{"unknown key", "logger:\n  level: debug\npaths:\n  ffmpg: ffmpeg\n", "unknown keys: logger.level, paths.ffmpg"},
		{"key differing only in case", "debug: true\nDebug: false\n", `keys "Debug" and "debug" name the same setting`},
		{"invalid value", "server:\n  http:\n    port: abc\n", "server.http.port"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := writeConfigFile(t, tt.contents)
			cfg, err := buildConfig(true, f)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Expected error containing %q Actual %v", tt.want, err)
			}
			if cfg != nil {
				t.Errorf("Expected no config Actual %+v", cfg)
			}
		})
	}
}

func TestFileConfigUnreadable(t *testing.T) {
	// A folder in place of the file fails to read for a reason other than being missing.
	f := filepath.Join(t.TempDir(), "unifiler.yaml")
	if err := os.Mkdir(f, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := buildConfig(true, f); err == nil {
		t.Error("Expected error reading a folder as config file Actual nil")
	}
}

func TestEnvConfigIgnoresUnknownVariable(t *testing.T) {
	t.Setenv("TFUNIFILER_LOGGER_LEVEL", "debug")
	t.Setenv("TFUNIFILER_SERVER_HTTP_PROT", "8080")
	t.Setenv("TFUNIFILER_LOG_LEVEL", "warn")
	cfg, err := buildConfig(true, filepath.Join(t.TempDir(), "unifiler.yaml"))
	if err != nil {
		t.Fatal("Error get config from env", err)
	}
	want := []string{"TFUNIFILER_LOGGER_LEVEL", "TFUNIFILER_SERVER_HTTP_PROT"}
	if !slices.Equal(cfg.IgnoredEnvVars, want) {
		t.Errorf("Wrong IgnoredEnvVars. Expected %v Actual %v", want, cfg.IgnoredEnvVars)
	}
	if applied := []string{"TFUNIFILER_LOG_LEVEL"}; !slices.Equal(cfg.AppliedEnvVars, applied) {
		t.Errorf("Wrong AppliedEnvVars. Expected %v Actual %v", applied, cfg.AppliedEnvVars)
	}
	if cfg.Logger.Level != "warn" {
		t.Errorf("Wrong Logger.Level. Expected '%s' Actual '%s'", "warn", cfg.Logger.Level)
	}
	if cfg.Server.HTTP.Port != 17017 {
		t.Errorf("Wrong Server.HTTP.Port. Expected default '%d' Actual '%d'", 17017, cfg.Server.HTTP.Port)
	}
}

func TestInitKoanfRetriesAfterFailure(t *testing.T) {
	cfg = nil
	t.Cleanup(func() { cfg = nil })

	t.Setenv("TFUNIFILER_SERVER_HTTP_PORT", "abc")
	for i := range 2 {
		if loaded, err := InitKoanf(false); err == nil || loaded != nil {
			t.Fatalf("Call %d: expected error and no config Actual %+v, %v", i+1, loaded, err)
		}
	}

	t.Setenv("TFUNIFILER_SERVER_HTTP_PORT", "8080")
	loaded, err := InitKoanf(false)
	if err != nil {
		t.Fatal("Error init config after fixing it", err)
	}
	if loaded.Server.HTTP.Port != 8080 {
		t.Errorf("Wrong Server.HTTP.Port. Expected '%d' Actual '%d'", 8080, loaded.Server.HTTP.Port)
	}
}

func TestIsPortableFindsMarkerNextToExecutable(t *testing.T) {
	exec, err := os.Executable()
	if err != nil {
		t.Fatalf("locating test binary: %v", err)
	}
	marker := filepath.Join(filepath.Dir(exec), "unifiler.portable")
	tests := []struct {
		name      string
		hasMarker bool
	}{
		{"no marker", false},
		{"marker next to executable", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hasMarker {
				if err := os.WriteFile(marker, nil, 0644); err != nil {
					t.Fatalf("creating marker: %v", err)
				}
				t.Cleanup(func() {
					if err := os.Remove(marker); err != nil {
						t.Errorf("removing marker: %v", err)
					}
				})
			}

			portable, err := IsPortable()

			if err != nil {
				t.Fatalf("IsPortable: %v", err)
			}
			if portable != tt.hasMarker {
				t.Errorf("IsPortable() = %t, want %t", portable, tt.hasMarker)
			}
		})
	}
}

func TestSanitizeKey(t *testing.T) {
	if got, err := SanitizeKey("Server.APIKey"); err != nil || got != KeyServerAPIKey {
		t.Errorf("SanitizeKey(%q) = %q, %v; want %q", "Server.APIKey", got, err, KeyServerAPIKey)
	}
	for _, key := range []string{"server", "logger.level"} {
		if got, err := SanitizeKey(key); err == nil {
			t.Errorf("SanitizeKey(%q) = %q; want error", key, got)
		}
	}
}

// Write contents to a config file in a temporary folder and return its path.
func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "unifiler.yaml")
	if err := os.WriteFile(f, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return f
}
