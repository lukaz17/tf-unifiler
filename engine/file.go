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

package engine

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/core/compression"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/internal/nullable"
	"github.com/tforceaio/tf-unifiler/tui"
)

// FileRenameMapping stores old and new filename after renaming for rollback.
type FileRenameMapping struct {
	Source string `json:"s,omitempty"`
	Target string `json:"t,omitempty"`
}

// FileModule handles user requests related to batch processing files.
type FileModule struct {
	cfg         *config.RootConfig
	interactive bool
	notifier    diag.Notifier
	logger      zerolog.Logger
	in          io.Reader
	out         io.Writer
}

// Return new FileModule.
func NewFileModule(c *Controller, cmdName string) *FileModule {
	return &FileModule{
		cfg:         c.Config,
		interactive: c.Interactive,
		notifier:    c.Notifier,
		logger:      c.CommandLogger("file", cmdName),
		in:          c.In,
		out:         c.Out,
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
			Int64("size", r.Hashes[0].Size).
			Msg("Hashed file.")
	}

	return nil
}

// Multi-pack inputs (files/folders) into an archives.
func (m *FileModule) Pack(inputs []string, output string, format string, level string, solid bool, dictSize int, password string, threads int, move bool, normalize bool, separate bool) error {
	if err := validateInputs(inputs); err != nil {
		return err
	}
	if output == "" {
		return errors.New("output path is required")
	}
	if err := validateArchiveFormat(format); err != nil {
		return err
	}
	compressLevel, err := compression.ParseLevel(level)
	if err != nil {
		return err
	}

	m.logger.Info().
		Strs("inputs", filesys.NormalizePaths(inputs, true)).
		Str("output", filesys.NormalizePath(output, true)).
		Str("compress", level).
		Str("format", format).
		Bool("move", move).
		Bool("normalize", normalize).
		Bool("separateInputs", separate).
		Bool("solid", solid).
		Int("dictSize", dictSize).
		Int("threads", threads).
		Msg("Packing parameters.")
	fmtInputs := strings.Builder{}
	for _, input := range inputs {
		fmtInputs.WriteString(fmt.Sprintf("\n - %s", input))
	}
	fmt.Fprintf(m.out, "\nInputs:%s\n", fmtInputs.String())
	fmt.Fprintf(m.out, "\nOutput:\n - %s\n", output)

	if m.interactive {
		fmt.Fprintln(m.out, "")
		opts, err := tui.RunCompressOptions(&tui.CompressOptionsValue{
			ArchiveType:    format,
			CompressLevel:  compressLevel,
			SolidMode:      solid,
			DictionarySize: dictSize,
			ThreadNumber:   threads,
			MoveToArchive:  move,
			NormalizeAttrs: normalize,
			SeparateInputs: separate,
		})
		if err != nil {
			return err
		}
		format = opts.ArchiveType
		level = compression.LevelCodes[opts.CompressLevel]
		solid = opts.SolidMode
		dictSize = opts.DictionarySize
		threads = opts.ThreadNumber
		move = opts.MoveToArchive
		normalize = opts.NormalizeAttrs
		separate = opts.SeparateInputs
	}

	m.logger.Info().
		Strs("inputs", filesys.NormalizePaths(inputs, true)).
		Str("output", filesys.NormalizePath(output, true)).
		Str("compress", level).
		Str("format", format).
		Bool("move", move).
		Bool("normalize", normalize).
		Bool("separateInputs", separate).
		Bool("solid", solid).
		Int("dictSize", dictSize).
		Int("threads", threads).
		Msg("Start packing files.")

	var solidMode nullable.Bool
	if solid {
		solidMode = nullable.FromBool(solid)
	}
	var threadNum nullable.Int
	if threads > 0 {
		threadNum = nullable.FromInt(threads)
	}
	_, err = packFiles(inputs, output, format, compressLevel, solidMode, dictSize, password, threadNum, move, normalize, separate, m.cfg.Path, m.logger, m.notifier, m.in, m.out, false)
	if err != nil {
		return err
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

	switch preset {
	case "crc32", "md4", "md5", "sha1", "sha256", "sha512":
		return m.renameByHash(inputs, preset, hex.EncodeToString([]byte(preset))+"_")
	default:
		return errors.New("preset is invalid")
	}
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
			Int64("size", r.Hashes[0].Size).
			Msg("Hashed file.")
	}

	mappings := []*FileRenameMapping{}
	for _, r := range fhResults {
		parent := filepath.Dir(r.Entry.RelativePath)
		ext := filepath.Ext(r.Entry.RelativePath)
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
			m.logger.Err(err).
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
func (m *FileModule) logError(err error) error {
	logProgramError(m.logger, err)
	return err
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
		RunE: func(cmd *cobra.Command, args []string) error {
			c := InitAppWithBanner(cmd.Context())
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "hash")
			return m.logError(m.Hash(flags.Inputs))
		},
	}
	rootCmd.AddCommand(hashCmd)

	packCmd := &cobra.Command{
		Use:   "pack <input>...",
		Short: "Pack files into an archive.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := InitAppWithBanner(cmd.Context())
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "pack")
			return m.logError(m.Pack(flags.Inputs, flags.Output, flags.Format, flags.Level, flags.Solid, flags.DictSize, flags.Password, flags.Threads, flags.Move, flags.Normalize, flags.Separate))
		},
	}
	packCmd.Flags().IntP("dict", "d", 0, "Dictionary size in MB for compression (0 = use default).")
	packCmd.Flags().StringP("format", "f", "7z", "Archive format: rar or 7z.")
	packCmd.Flags().StringP("level", "l", "normal", "Compression level: none (0), fast (1), normal (2), high (3), ultra (4)")
	packCmd.Flags().BoolP("move", "m", false, "Move files into archive after compression.")
	packCmd.Flags().BoolP("normalize", "n", false, "Normalize file attributes before compression.")
	packCmd.Flags().StringP("output", "o", "", "Output archive file path.")
	packCmd.Flags().String("password", "", "Password to protect the archive.")
	MarkFlagSensitive(packCmd, "password")
	packCmd.Flags().BoolP("separate", "e", false, "Compress each input into a separate archive.")
	packCmd.Flags().BoolP("solid", "s", false, "Enable solid mode for better compression ratio.")
	packCmd.Flags().IntP("threads", "t", 0, "Number of threads to use for compression.")
	rootCmd.AddCommand(packCmd)

	renameCmd := &cobra.Command{
		Use:   "rename <input>...",
		Short: "Multi-rename files and directories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := InitAppWithBanner(cmd.Context())
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "rename")
			return m.logError(m.Rename(flags.Inputs, flags.Preset))
		},
	}
	renameCmd.Flags().StringP("preset", "p", "", "Name of pre-defined settings for renaming.")
	rootCmd.AddCommand(renameCmd)

	return rootCmd
}

// FileFlags contains all flags used by File module.
type FileFlags struct {
	DictSize  int
	Format    string
	Inputs    []string
	Level     string
	Move      bool
	Normalize bool
	Output    string
	Password  string
	Preset    string
	Separate  bool
	Solid     bool
	Threads   int
}

// Extract all flags from a Cobra Command.
func ParseFileFlags(cmd *cobra.Command, args []string) *FileFlags {
	dictSize, _ := cmd.Flags().GetInt("dict")
	format, _ := cmd.Flags().GetString("format")
	inputs, _ := cmd.Flags().GetStringArray("inputs")
	level, _ := cmd.Flags().GetString("level")
	move, _ := cmd.Flags().GetBool("move")
	normalize, _ := cmd.Flags().GetBool("normalize")
	output, _ := cmd.Flags().GetString("output")
	password, _ := cmd.Flags().GetString("password")
	preset, _ := cmd.Flags().GetString("preset")
	separate, _ := cmd.Flags().GetBool("separate")
	solid, _ := cmd.Flags().GetBool("solid")
	threads, _ := cmd.Flags().GetInt("threads")
	inputs = append(args, inputs...)

	return &FileFlags{
		DictSize:  dictSize,
		Format:    format,
		Inputs:    inputs,
		Level:     level,
		Move:      move,
		Normalize: normalize,
		Output:    output,
		Password:  password,
		Preset:    preset,
		Separate:  separate,
		Solid:     solid,
		Threads:   threads,
	}
}
