// Copyright (C) 2024  T-Force I/O
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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/config"
)

// ConfigModule handles configuration management commands.
type ConfigModule struct {
	cfg    *config.RootConfig
	logger zerolog.Logger
	out    io.Writer
}

// Return new ConfigModule instance.
func NewConfigModule(c *Controller) *ConfigModule {
	return &ConfigModule{
		cfg:    c.Config,
		logger: c.ModuleLogger("config"),
		out:    c.Out,
	}
}

// Export active configuration to stdout or file.
func (m *ConfigModule) Export(outputPath string, diffOnly bool) error {
	m.logger.Info().
		Bool("diffOnly", diffOnly).
		Msg("Start export configuration.")

	k := koanf.New(".")
	if err := k.Load(structs.Provider(m.cfg, "koanf"), nil); err != nil {
		return err
	}

	if diffOnly {
		defaults, err := config.DefaultConfig()
		if err != nil {
			return err
		}
		for key, defVal := range defaults.All() {
			if val := k.Get(key); val != nil && reflect.DeepEqual(val, defVal) {
				k.Delete(key)
			}
		}
	}
	data, err := k.Marshal(yaml.Parser())
	if err != nil {
		return err
	}
	if outputPath == "" {
		fmt.Fprint(m.out, string(data))
	} else {
		err := m.writeFile(outputPath, data)
		if err != nil {
			return err
		}
		m.logger.Info().Str("outputPath", outputPath).
			Msg("Configuration exported successfully.")
	}

	return nil
}

// Get value of key from active configuration.
func (m *ConfigModule) Get(key string) error {
	if err := validateRequiredString(key, "key"); err != nil {
		return err
	}

	k := koanf.New(".")
	if err := k.Load(structs.Provider(m.cfg, "koanf"), nil); err != nil {
		return err
	}

	if val := k.Get(key); val != nil {
		fmt.Fprintf(m.out, "%v\n", val)
	}
	return nil
}

// Write value of key to active YAML config file.
func (m *ConfigModule) Set(key, value string) error {
	if err := validateRequiredString(key, "key"); err != nil {
		return err
	}
	if err := validateRequiredString(value, "value"); err != nil {
		return err
	}
	// An unknown key would be written, then fail every later load of the file.
	key, err := config.SanitizeKey(key)
	if err != nil {
		return err
	}

	// Loaded with lowercase keys, so the file never holds two keys that differ only in case.
	k, err := config.LoadFile(m.cfg.ConfigFile)
	if err != nil {
		return err
	}
	err = k.Set(key, value)
	if err != nil {
		return err
	}
	data, err := k.Marshal(yaml.Parser())
	if err != nil {
		return err
	}
	err = m.writeFile(m.cfg.ConfigFile, data)
	if err != nil {
		return err
	}

	// The YAML keys are matched case-insensitively when the file is loaded.
	logged := opx.Ternary(strings.EqualFold(key, config.KeyServerAPIKey), redactedValue, value)
	m.logger.Info().Str("key", key).Str("value", logged).
		Msg("Configuration updated successfully.")
	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *ConfigModule) logError(err error) error {
	return LogCommandError(m.logger, err)
}

// Write data to a file safely.
func (m *ConfigModule) writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Configuration may hold the server's API key, so only the owner may read it.
	if err := os.Chmod(path, 0600); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to restrict permissions of file %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// Define Cobra Command for Config module.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage application settings.",
		Long: `Manage application settings, which come from unifiler.yaml in the config
folder. Environment variables override them, and a setting found in neither
keeps its default. Run "unifiler config export" to list every key.

The config folder, which also holds the log file logs/unifiler.log, is:
  - Windows: %APPDATA%\TFunifiler
  - Linux and macOS: ~/.config/tfunifiler
  - Portable mode, when a file named unifiler.portable is next to the
    executable: the executable's folder

An environment variable is named TFUNIFILER_ followed by the key in upper case,
with _ in place of ., such as TFUNIFILER_LOG_LEVEL=debug.`,
	}

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Print every setting in use as YAML, or save it to a file.",
		Long: `Print every setting in use, after environment variables are applied, as YAML,
or save it to the --output file.

The output includes server.apikey when an API key is set, so keep it private.`,
		Example: `  unifiler config export
  unifiler config export --diff -o settings.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			m := NewConfigModule(c)
			f := m.ParseConfigFlags(cmd, args)
			return m.logError(m.Export(f.Output, f.Diff))
		},
	}
	exportCmd.Flags().BoolP("diff", "d", false, "Only export the settings that differ from the defaults.")
	exportCmd.Flags().StringP("output", "o", "", "File to save the YAML to; printed to the terminal when empty.")

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value in use for a setting.",
		Long: `Print the value in use for a setting, from the config file, an environment
variable or the default. Keys are dot-separated, such as paths.x7z; nothing is
printed for an unknown key.`,
		Example: `  unifiler config get paths.x7z`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitApp(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			m := NewConfigModule(c)
			f := m.ParseConfigFlags(cmd, args)
			return m.logError(m.Get(f.Key))
		},
	}
	getCmd.Flags().StringP("key", "k", "", "Setting key, such as paths.x7z; same as the <key> argument.")

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Save a setting to the config file.",
		Long: `Save a setting to unifiler.yaml in the config folder; an unknown key is
rejected. An environment variable for the same key still overrides the file,
and a running server applies the change only after a restart.`,
		Example: `  unifiler config set paths.x7z /opt/7-zip/7zz
  unifiler config set server.concurrency 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitApp(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			m := NewConfigModule(c)
			f := m.ParseConfigFlags(cmd, args)
			return m.logError(m.Set(f.Key, f.Value))
		},
	}
	setCmd.Flags().StringP("key", "k", "", "Setting key, such as paths.x7z; same as the <key> argument.")
	setCmd.Flags().StringP("value", "v", "", "Value to save; same as the <value> argument.")

	cmd.AddCommand(exportCmd, getCmd, setCmd)

	return cmd
}

// ConfigFlags contains all flags used by Config module.
type ConfigFlags struct {
	Diff   bool
	Key    string
	Output string
	Value  string
}

// Extract all flags from a Cobra Command.
func (m *ConfigModule) ParseConfigFlags(cmd *cobra.Command, args []string) *ConfigFlags {
	diff, _ := cmd.Flags().GetBool("diff")
	key, _ := cmd.Flags().GetString("key")
	output, _ := cmd.Flags().GetString("output")
	value, _ := cmd.Flags().GetString("value")

	if len(args) >= 1 && key == "" {
		key = args[0]
	}
	if len(args) >= 2 && value == "" {
		value = args[1]
	}

	f := &ConfigFlags{
		Diff:   diff,
		Output: output,
		Key:    key,
		Value:  value,
	}
	return f
}
