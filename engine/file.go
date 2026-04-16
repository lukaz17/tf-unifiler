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

package engine

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/filesys"
)

// Struct FileRenameMapping stores old and new filename after renaming for rollback.
type FileRenameMapping struct {
	Source string `json:"s,omitempty"`
	Target string `json:"t,omitempty"`
}

// FileModule handles user requests related to batch processing of files in general.
type FileModule struct {
	logger zerolog.Logger
}

// Return new FileModule.
func NewFileModule(c *Controller, cmdName string) *FileModule {
	return &FileModule{
		logger: c.CommandLogger("file", cmdName),
	}
}

// Compute hashes using common algorithms (MD5, SHA-1, SHA-256, SHA-512) for inputs (files/folders),
// then print the result to console.
func (m *FileModule) Hash(inputs []string) error {
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	m.logger.Info().
		Strs("files", inputs).
		Msg("Start hashing files.")

	contents, err := filesys.List(inputs, true)
	if err != nil {
		return err
	}

	algos := []string{"md5", "sha1", "sha256", "sha512"}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, algos)
		if err != nil {
			m.logger.Info().
				Str("path", c.RelativePath).
				Msg("Failed to compute hash.")
			return err
		}
		m.logger.Info().
			Str("md5", hex.EncodeToString(fhResults[0].Hash)).
			Str("path", c.RelativePath).
			Str("sha1", hex.EncodeToString(fhResults[1].Hash)).
			Str("sha256", hex.EncodeToString(fhResults[2].Hash)).
			Str("sha512", hex.EncodeToString(fhResults[3].Hash)).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
	}

	return nil
}

// Multi-rename files. Input which is directories will be ignored.
func (m *FileModule) Rename(inputs []string, preset string) error {
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	m.logger.Info().
		Strs("inputs", inputs).
		Str("preset", preset).
		Msg("Start renaming file.")

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
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}

	contents, err := filesys.List(inputs, false)
	if err != nil {
		return err
	}

	files := []*filesys.FsEntry{}
	for _, c := range contents {
		if !c.IsDir {
			files = append(files, c)
		}
	}
	hResults := []*hasher.HashResult{}
	for _, c := range files {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, []string{algo})
		if err != nil {
			m.logger.Info().
				Str("path", c.RelativePath).
				Msg("Failed to compute hash.")
			return err
		}
		m.logger.Info().
			Str("algo", algo).
			Str("path", c.RelativePath).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
		hResults = append(hResults, fhResults...)
	}

	mappings := []*FileRenameMapping{}
	for _, e := range hResults {
		parent := path.Dir(e.Path)
		ext := path.Ext(e.Path)
		targetName := prefix + hex.EncodeToString(e.Hash) + ext
		target := opx.Ternary(parent == ".", targetName, filesys.Join(parent, targetName))
		mapping := &FileRenameMapping{
			Source: e.Path,
			Target: target,
		}
		mappings = append(mappings, mapping)
	}

	currentTimestamp := time.Now().UnixMilli()
	rollbackFilePath := filesys.Join(".", "unifiler-file-rename-"+strconv.FormatInt(currentTimestamp, 10)+".json")
	fContent, _ := json.Marshal(mappings)
	fContents := []string{string(fContent)}
	err = filesys.WriteLines(rollbackFilePath, fContents)
	if err == nil {
		m.logger.Info().
			Int("lineCount", len(fContents)).
			Str("path", rollbackFilePath).
			Msg("Written rollback file.")
	} else {
		m.logger.Info().
			Str("path", rollbackFilePath).
			Msg("Failed to write rollback file.")
		return err
	}

	for _, e := range mappings {
		if e.Source == e.Target {
			m.logger.Info().
				Str("src", e.Source).
				Str("dest", e.Target).
				Msg("Skipped. File is already renamed.")
			continue
		}
		if filesys.IsFileExist(e.Target) {
			m.logger.Info().
				Str("src", e.Source).
				Str("dest", e.Target).
				Msg("Skipped. Target file is existed.")
			continue
		}
		err := os.Rename(e.Source, e.Target)
		if err != nil {
			m.logger.Err(err)
			m.logger.Info().
				Str("src", e.Source).
				Str("dest", e.Target).
				Msg("Failed to rename file.")
			continue
		}
		m.logger.Info().
			Str("src", e.Source).
			Str("target", e.Target).
			Msg("Renamed file.")
	}

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *FileModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

// Define Cobra Command for File module.
func FileCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "file",
		Short: "Batch file processing in general.",
	}

	hashCmd := &cobra.Command{
		Use:   "hash <input>...",
		Short: "Compute hashes for files using common algorithms (MD5, SHA-1, SHA-256, SHA-512).",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "hash")
			m.logError(m.Hash(flags.Inputs))
		},
	}
	hashCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files/Directories to hash.")
	rootCmd.AddCommand(hashCmd)

	renameCmd := &cobra.Command{
		Use:   "rename <input>...",
		Short: "Rename multiples file using pre-defined settings.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "rename")
			m.logError(m.Rename(flags.Inputs, flags.Preset))
		},
	}
	renameCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files to rename. Directories will be ignored.")
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
