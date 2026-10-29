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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

var (
	cfg   *RootConfig
	cfgMu sync.Mutex
)

const configFileName = "unifiler.yaml"

// Init configuration for the application.
func InitKoanf(useFS bool) (*RootConfig, error) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if cfg != nil {
		return cfg, nil
	}

	isPortable := !useFS || IsPortable()
	configFile := configFileName

	if isPortable {
		exec, _ := os.Executable()
		configFile = filepath.Join(filepath.Dir(exec), configFileName)
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		configFile = filepath.Join(home, ".config", "tfunifiler", configFileName)
	} else if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		configFile = filepath.Join(appData, "TFunifiler", configFileName)
	}

	var err error
	cfg, err = buildConfig(useFS, configFile)
	if err != nil {
		return cfg, err
	}

	cfg.ConfigDir = filepath.Dir(configFile)
	cfg.ConfigFile = configFile
	cfg.IsPortable = isPortable
	return cfg, nil
}

// Get default configuration values.
func DefaultConfig() (*koanf.Koanf, error) {
	k := koanf.New(".")

	err := k.Load(
		structs.Provider(RootConfig{
			Logger: &LoggerConfig{
				Level: "info",
			},
			Path: &PathConfig{
				FFMpegPath:      "ffmpeg",
				ImageMagickPath: "magick",
				MediaInfoPath:   "mediainfo",
				RarPath:         "rar",
				X7zPath:         "7z",
				X264Path:        "x264",
				X265Path:        "x265",
			},
			Server: &ServerConfig{
				HTTP: &HTTPConfig{
					Port: 8080,
				},
				Concurrency: 1,
				Backlog:     256,
			},
		}, "koanf"),
		nil,
	)
	if err != nil {
		return nil, err
	}

	return k, nil
}

// Check if the application is in portable mode.
func IsPortable() bool {
	exec, _ := os.Executable()
	portableFile := filepath.Join(filepath.Dir(exec), "unifiler.portable")
	return fileExists(portableFile)
}

// Build configurations for the application with the following priority:
// Environment variables -> YAML configuration file -> default values.
func buildConfig(useFS bool, f string) (*RootConfig, error) {
	k, err := DefaultConfig()
	if err != nil {
		return &RootConfig{}, err
	}

	var loadErr error
	if useFS && fileExists(f) {
		if _, err := configFromYaml(k, f); err != nil {
			loadErr = fmt.Errorf("loading config file %s: %w", f, err)
		}
	}
	if _, err := configFromEnv(k); err != nil && loadErr == nil {
		loadErr = fmt.Errorf("loading environment config: %w", err)
	}

	var config RootConfig
	if err := k.Unmarshal("", &config); err != nil {
		return &config, err
	}
	return &config, loadErr
}

// Get configuration values from environment variables.
func configFromEnv(k *koanf.Koanf) (*koanf.Koanf, error) {
	err := k.Load(env.Provider("TFUNIFILER_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(
				strings.TrimPrefix(s, "TFUNIFILER_")), "_", ".", -1)
	}), nil)
	if err != nil {
		return k, err
	}
	return k, nil
}

// Get configuration values from YAML file.
func configFromYaml(k *koanf.Koanf, f string) (*koanf.Koanf, error) {
	err := k.Load(file.Provider(f), yaml.Parser())
	if err != nil {
		return k, err
	}
	return k, nil
}

// Return whether a file existed in file system.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
