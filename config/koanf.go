// Copyright (C) 2024 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/tforceaio/tf-unifiler/filesys"
)

// Singleton instance of configuration.
var cfg *RootConfig

// Initialize RootConfiguration follow the sequence:
// Default values -> YML file (f) -> Environment variable.
// The latter will override the former.
// YML file will only be used if useFS is true.
func BuildConfig(useFS bool, f string) (*RootConfig, error) {
	k := defaultConfig()
	if useFS && filesys.IsFileExist(f) {
		k, _ = configFromYaml(k, f)
	}
	k, _ = configFromEnv(k)

	var config RootConfig
	err := k.Unmarshal("", &config)
	return &config, err
}

// Entrypoint for creating RootConfiguration instance using Koanf.
// YML file will only be used if useFS is true.
func InitKoanf(useFS bool) (*RootConfig, error) {
	if cfg != nil {
		return cfg, nil
	}
	isPortable := !useFS || IsPortable()
	configFile := "unifiler.yml"
	if isPortable {
		exec, _ := os.Executable()
		exec, _ = filesys.GetAbsPath(exec)
		configFile = path.Join(path.Dir(exec), "unifiler.yml")
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		configFile = path.Join(home, ".config", "unifiler", "unifiler.yml")
	} else if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		configFile = path.Join(appData, "Unifiler", "unifiler.yml")
	}
	var err error
	cfg, err = BuildConfig(useFS, configFile)
	if err != nil {
		return cfg, err
	}

	cfg.ConfigDir = path.Dir(configFile)
	cfg.ConfigFile = configFile
	cfg.IsPortable = isPortable
	return cfg, nil
}

// Detect whether the app is running in portable mode.
func IsPortable() bool {
	exec, _ := os.Executable()
	exec, _ = filesys.GetAbsPath(exec)
	portableFile := path.Join(path.Dir(exec), "unifiler.portable")
	return filesys.IsFileExist(portableFile)
}

// Returns default values for RootConfig
func defaultConfig() *koanf.Koanf {
	var k = koanf.New(".")

	k.Load(
		structs.Provider(RootConfig{
			Path: &PathConfig{
				FFMpegPath:      "ffmpeg",
				ImageMagickPath: "magick",
				MediaInfoPath:   "mediainfo",
				X264Path:        "x264",
				X265Path:        "x265",
			},
		}, "koanf"),
		nil,
	)

	return k
}

// Override existing values in Koanf instance with value from environtment variables.
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

// Override existing values in Koanf instance with value from YAML file.
func configFromYaml(k *koanf.Koanf, f string) (*koanf.Koanf, error) {
	err := k.Load(file.Provider(f), yaml.Parser())
	if err != nil {
		return k, err
	}
	return k, nil
}
