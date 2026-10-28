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
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/config"
)

// ConfigModule handles configuration management commands.
type ConfigModule struct {
	cfg    *config.RootConfig
	logger zerolog.Logger
}

// Return new ConfigModule instance.
func NewConfigModule(c *Controller) *ConfigModule {
	return &ConfigModule{
		cfg:    c.Config,
		logger: c.ModuleLogger("config"),
	}
}

// Export active configuration to stdout or file.
func (m *ConfigModule) Export(outputPath string, diffOnly bool) error {
	m.logger.Info().
		Bool("diffOnly", diffOnly).
		Msg("Start export configuration.")

	k := koanf.New(".")
	k.Load(structs.Provider(m.cfg, "koanf"), nil)

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
	data, _ := k.Marshal(yaml.Parser())
	if outputPath == "" {
		fmt.Print(string(data))
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
	k.Load(structs.Provider(m.cfg, "koanf"), nil)

	if val := k.Get(key); val != nil {
		fmt.Printf("%v\n", val)
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

	k := koanf.New(".")
	if _, err := os.Stat(m.cfg.ConfigFile); err == nil {
		err = k.Load(file.Provider(m.cfg.ConfigFile), yaml.Parser())
		if err != nil {
			return err
		}
	}
	err := k.Set(key, value)
	if err != nil {
		return err
	}
	data, _ := k.Marshal(yaml.Parser())
	err = m.writeFile(m.cfg.ConfigFile, data)
	if err != nil {
		return err
	}

	m.logger.Info().Str("key", key).Str("value", value).
		Msg("Configuration updated successfully.")
	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *ConfigModule) logError(err error) error {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred.")
	}
	return err
}

// Write data to a file safely.
func (m *ConfigModule) writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// Define Cobra Command for Config module.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage application configuration.",
	}

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export active configuration to stdout or file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := InitApp()
			defer c.Close()
			m := NewConfigModule(c)
			f := m.ParseConfigFlags(cmd, args)
			return m.logError(m.Export(f.Output, f.Diff))
		},
	}
	exportCmd.Flags().BoolP("diff", "d", false, "Export diffrences from default configuration only.")
	exportCmd.Flags().StringP("output", "o", "", "Output file path. Empty will output ot stdout.")

	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get the active value of a configuration key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := NewController(true)
			defer c.Close()
			m := NewConfigModule(c)
			f := m.ParseConfigFlags(cmd, args)
			return m.logError(m.Get(f.Key))
		},
	}
	getCmd.Flags().StringP("key", "k", "", "Configuration key in dot-delimited format.")

	setCmd := &cobra.Command{
		Use:   "set <key> [value]",
		Short: "Set a configuration value in the active YAML file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := NewController(true)
			defer c.Close()
			m := NewConfigModule(c)
			f := m.ParseConfigFlags(cmd, args)
			return m.logError(m.Set(f.Key, f.Value))
		},
	}
	setCmd.Flags().StringP("key", "k", "", "Configuration key in dot-delimited format.")
	setCmd.Flags().StringP("value", "v", "", "Configuration value")

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
