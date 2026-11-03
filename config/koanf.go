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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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

// Init configuration for the application. A failed load is not kept, so the next call loads again.
func InitKoanf(useFS bool) (*RootConfig, error) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if cfg != nil {
		return cfg, nil
	}

	isPortable := !useFS || IsPortable()
	configFile, err := configFilePath(isPortable)
	if err != nil {
		return nil, err
	}

	loaded, err := buildConfig(useFS, configFile)
	if err != nil {
		return nil, err
	}

	loaded.ConfigDir = filepath.Dir(configFile)
	loaded.ConfigFile = configFile
	loaded.Portable = isPortable
	cfg = loaded
	return cfg, nil
}

// Get default configuration values.
func DefaultConfig() (*koanf.Koanf, error) {
	k := koanf.New(".")

	err := k.Load(
		structs.Provider(RootConfig{
			Debug: false,
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
				Name: "tf-unifiler",
				HTTP: &HTTPConfig{
					Hostname: "127.0.0.1",
					Port:     17017,
				},
				Autostart:   true,
				APIKey:      "",
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

// Load the YAML config file f with its keys lowercased, so they match settings case-insensitively.
func LoadFile(f string) (*koanf.Koanf, error) {
	parsed := koanf.New(".")
	err := parsed.Load(file.Provider(f), yaml.Parser())
	if errors.Is(err, fs.ErrNotExist) {
		return koanf.New("."), nil
	}
	if err != nil {
		return nil, fmt.Errorf("config file %s: %w", f, err)
	}
	loaded, err := normalizeKeys(parsed)
	if err != nil {
		return nil, fmt.Errorf("config file %s: %w", f, err)
	}
	return loaded, nil
}

// Return key in the lowercase form settings are stored under, such as "server.http.port" for "Server.HTTP.Port".
// Return error if the key refers to non-existence conifg.
func SanitizeKey(key string) (string, error) {
	defaults, err := DefaultConfig()
	if err != nil {
		return "", err
	}
	lower := strings.ToLower(key)
	if !slices.Contains(defaults.Keys(), lower) {
		return "", fmt.Errorf("unknown setting %q", key)
	}
	return lower, nil
}

// Build configurations for the application with the following priority:
// Environment variables -> YAML configuration file -> default values.
func buildConfig(useFS bool, f string) (*RootConfig, error) {
	k, err := DefaultConfig()
	if err != nil {
		return nil, err
	}

	if useFS {
		if _, err := configFromYaml(k, f); err != nil {
			return nil, err
		}
	}
	applied, ignored, err := configFromEnv(k)
	if err != nil {
		return nil, err
	}

	var config RootConfig
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("decoding settings: %w", err)
	}
	config.AppliedEnvVars = applied
	config.IgnoredEnvVars = ignored
	return &config, nil
}

// Return the path of the config file: next to the executable in portable mode, otherwise in the user's config folder.
func configFilePath(isPortable bool) (string, error) {
	if isPortable {
		exec, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("locating executable: %w", err)
		}
		return filepath.Join(filepath.Dir(exec), configFileName), nil
	}
	switch runtime.GOOS {
	case "linux", "darwin":
		// Not os.UserConfigDir, which follows XDG_CONFIG_HOME and is ~/Library/Application Support on macOS.
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locating config folder: %w", err)
		}
		return filepath.Join(home, ".config", "tfunifiler", configFileName), nil
	case "windows":
		appData, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("locating config folder: %w", err)
		}
		return filepath.Join(appData, "TFunifiler", configFileName), nil
	}
	return configFileName, nil
}

// Get configuration values from environment variables.
// Return the names of the variables applied, then of those naming no setting; the latter are skipped, so a stray variable does not stop the program.
func configFromEnv(k *koanf.Koanf) ([]string, []string, error) {
	names := make(map[string]string) // setting key -> variable name, to report applied and skipped variables.
	vars := koanf.New(".")
	err := vars.Load(env.Provider("TFUNIFILER_", ".", func(s string) string {
		key := strings.Replace(
			strings.ToLower(
				strings.TrimPrefix(s, "TFUNIFILER_")), "_", ".", -1)
		names[key] = s
		return key
	}), nil)
	if err != nil {
		return nil, nil, err
	}

	defaults, err := DefaultConfig()
	if err != nil {
		return nil, nil, err
	}
	var applied, ignored []string
	for _, key := range vars.Keys() {
		if defaults.Exists(key) {
			applied = append(applied, names[key])
			continue
		}
		ignored = append(ignored, names[key])
		vars.Delete(key)
	}
	return applied, ignored, k.Merge(vars)
}

// Get configuration values from YAML file.
func configFromYaml(k *koanf.Koanf, f string) (*koanf.Koanf, error) {
	loaded, err := LoadFile(f)
	if err != nil {
		return k, err
	}
	return k, k.Merge(loaded)
}

// Return whether a file existed in file system.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Return the values of src under lowercase keys.
func normalizeKeys(src *koanf.Koanf) (*koanf.Koanf, error) {
	defaults, err := DefaultConfig()
	if err != nil {
		return nil, err
	}

	var unknown []string
	seen := make(map[string]string) // lowercase key -> key as written.
	out := koanf.New(".")
	for _, key := range src.Keys() {
		lower := strings.ToLower(key)
		if !defaults.Exists(lower) {
			unknown = append(unknown, key)
			continue
		}
		val := src.Get(key)
		if val == nil {
			continue
		}
		if prev, ok := seen[lower]; ok {
			return nil, fmt.Errorf("keys %q and %q name the same setting", prev, key)
		}
		seen[lower] = key
		if err := out.Set(lower, val); err != nil {
			return nil, err
		}
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unknown keys: %s", strings.Join(unknown, ", "))
	}
	return out, nil
}
