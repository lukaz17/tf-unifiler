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

package engine

import (
	"encoding/hex"
	"errors"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
)

// Struct FileRenameMapping stores old and new filename after renaming for rollback.
type FileRenameMapping struct {
	Source string `json:"s,omitempty"`
	Target string `json:"t,omitempty"`
}

// FileModule handles user requests related to batch processing files.
type FileModule struct {
	notifier diag.Notifier
	logger   zerolog.Logger
}

// Return new FileModule.
func NewFileModule(c *Controller, cmdName string) *FileModule {
	return &FileModule{
		notifier: c.Notifier,
		logger:   c.CommandLogger("file", cmdName),
	}
}

// Compute hashes using common algorithms (MD5, SHA-1, SHA-256, SHA-512) for inputs (files/folders),
// then print the result to console.
func (m *FileModule) Hash(inputs []string) error {
	if err := validateInputs(inputs); err != nil {
		return err
	}
	m.logger.Info().
		Strs("files", filesys.NormalizePaths(inputs, true)).
		Msg("Start hashing files.")

	algos := []string{"crc32", "md5", "sha1", "sha256", "sha512"}
	fhResults, err := listAndHashFiles(inputs, algos, true, m.notifier)
	if err != nil {
		return err
	}

	for _, r := range fhResults {
		m.logger.Info().
			Str("crc32", hex.EncodeToString(r.Hashes[0].Hash)).
			Str("md5", hex.EncodeToString(r.Hashes[1].Hash)).
			Str("path", filesys.NormalizePath(r.Entry.RelativePath, true)).
			Str("sha1", hex.EncodeToString(r.Hashes[2].Hash)).
			Str("sha256", hex.EncodeToString(r.Hashes[3].Hash)).
			Str("sha512", hex.EncodeToString(r.Hashes[4].Hash)).
			Int("size", r.Hashes[0].Size).
			Msg("Hashed file.")
	}

	return nil
}

// Multi-rename files. Input which is directories will be ignored.
func (m *FileModule) Rename(inputs []string, preset string) error {
	if err := validateInputs(inputs); err != nil {
		return err
	}
	m.logger.Info().
		Strs("inputs", filesys.NormalizePaths(inputs, true)).
		Str("preset", preset).
		Msg("Start renaming file.")

	if preset == "crc32" {
		return m.renameByHash(inputs, preset, "6372633332_")
	}
	if preset == "md4" {
		return m.renameByHash(inputs, preset, "6d6434_")
	}
	if preset == "md5" {
		return m.renameByHash(inputs, preset, "6d6435_")
	}
	if preset == "sha1" {
		return m.renameByHash(inputs, preset, "73686131_")
	}
	if preset == "sha256" {
		return m.renameByHash(inputs, preset, "736861323536_")
	}
	if preset == "sha512" {
		return m.renameByHash(inputs, preset, "736861353132_")
	}

	return errors.New("preset is invalid")
}

// Rename files using hashes of their contents.
func (m *FileModule) renameByHash(inputs []string, algo string, prefix string) error {
	fhResults, err := listAndHashFiles(inputs, []string{algo}, false, m.notifier)
	if err != nil {
		return err
	}

	for _, r := range fhResults {
		m.logger.Info().
			Str("algo", algo).
			Str("path", filesys.NormalizePath(r.Entry.RelativePath, true)).
			Int("size", r.Hashes[0].Size).
			Msg("Hashed file.")
	}

	mappings := []*FileRenameMapping{}
	for _, r := range fhResults {
		parent := path.Dir(r.Entry.RelativePath)
		ext := path.Ext(r.Entry.RelativePath)
		targetName := prefix + hex.EncodeToString(r.Hashes[0].Hash) + ext
		target := opx.Ternary(parent == ".", targetName, filesys.Join(parent, targetName))
		mapping := &FileRenameMapping{
			Source: r.Entry.RelativePath,
			Target: target,
		}
		mappings = append(mappings, mapping)
	}

	rollbackFilePath, err := writeJSON(".", "unifiler-file-rename-", mappings)
	if err != nil {
		m.logger.Info().
			Str("path", filesys.NormalizePath(rollbackFilePath, true)).
			Msg("Failed to write rollback file.")
		return err
	}
	m.logger.Info().
		Str("path", filesys.NormalizePath(rollbackFilePath, true)).
		Msg("Written rollback file.")

	for _, e := range mappings {
		if e.Source == e.Target {
			m.logger.Info().
				Str("src", filesys.NormalizePath(e.Source, true)).
				Str("dest", filesys.NormalizePath(e.Target, true)).
				Msg("Skipped. File is already renamed.")
			continue
		}
		if filesys.IsFileExist(e.Target) {
			m.logger.Info().
				Str("src", filesys.NormalizePath(e.Source, true)).
				Str("dest", filesys.NormalizePath(e.Target, true)).
				Msg("Skipped. Target file is existed.")
			continue
		}
		err := os.Rename(e.Source, e.Target)
		if err != nil {
			m.logger.Err(err)
			m.logger.Info().
				Str("src", filesys.NormalizePath(e.Source, true)).
				Str("dest", filesys.NormalizePath(e.Target, true)).
				Msg("Failed to rename file.")
			continue
		}
		m.logger.Info().
			Str("src", filesys.NormalizePath(e.Source, true)).
			Str("target", filesys.NormalizePath(e.Target, true)).
			Msg("Renamed file.")
	}

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *FileModule) logError(err error) {
	logProgramError(m.logger, err)
}

// Define Cobra Command for File module.
func FileCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "file",
		Short: "Batch processing file.",
	}
	rootCmd.PersistentFlags().StringArrayP("inputs", "i", []string{}, "Files/Directories to process.")

	hashCmd := &cobra.Command{
		Use:   "hash <input>...",
		Short: "Compute and print hashes for files.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "hash")
			m.logError(m.Hash(flags.Inputs))
		},
	}
	rootCmd.AddCommand(hashCmd)

	renameCmd := &cobra.Command{
		Use:   "rename <input>...",
		Short: "Multi-rename files and directories.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "rename")
			m.logError(m.Rename(flags.Inputs, flags.Preset))
		},
	}
	renameCmd.Flags().StringP("preset", "p", "", "Name of pre-defined settings for renaming.")
	rootCmd.AddCommand(renameCmd)

	return rootCmd
}

// Struct FileFlags contains all flags used by File module.
type FileFlags struct {
	Inputs []string
	Preset string
}

// Extract all flags from a Cobra Command.
func ParseFileFlags(cmd *cobra.Command, args []string) *FileFlags {
	inputs, _ := cmd.Flags().GetStringArray("inputs")
	preset, _ := cmd.Flags().GetString("preset")
	inputs = append(args, inputs...)

	return &FileFlags{
		Inputs: inputs,
		Preset: preset,
	}
}
