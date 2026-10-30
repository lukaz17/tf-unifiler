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

// RootConfig contains all available configurations for the application.
type RootConfig struct {
	ConfigDir  string        `koanf:"-"`
	ConfigFile string        `koanf:"-"`
	IsPortable bool          `koanf:"-"`
	Logger     *LoggerConfig `koanf:"log"`
	Path       *PathConfig   `koanf:"paths"`
	Server     *ServerConfig `koanf:"server"`
}

// HTTPConfig contains configurations for HTTP server connection.
type HTTPConfig struct {
	Port int `koanf:"port"`
}

// LoggerConfig contains configurations for logging.
type LoggerConfig struct {
	Level string `koanf:"level"`
}

// PathConfig contains configurations for external dependencies location.
type PathConfig struct {
	FFMpegPath      string `koanf:"ffmpeg"`
	ImageMagickPath string `koanf:"imagemagick"`
	MediaInfoPath   string `koanf:"mediainfo"`
	RarPath         string `koanf:"rar"`
	X7zPath         string `koanf:"x7z"`
	X264Path        string `koanf:"x264"`
	X265Path        string `koanf:"x265"`
}

// ServerConfig contains configurations for HTTP server.
type ServerConfig struct {
	Name        string      `koanf:"name"`
	HTTP        *HTTPConfig `koanf:"http"`
	Autostart   bool        `koanf:"autostart"`
	Concurrency int         `koanf:"concurrency"`
	Backlog     int         `koanf:"backlog"`
}
