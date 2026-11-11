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
	return LogCommandError(m.logger, err)
}

// Define Cobra Command for File module.
func FileCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "file",
		Short: "Work with files in general.",
	}
	MarkCommandQueueable(rootCmd)
	rootCmd.PersistentFlags().StringArrayP("inputs", "i", []string{}, "Files or folders to process, in addition to the arguments; repeat for several.")

	hashCmd := &cobra.Command{
		Use:   "hash <input>...",
		Short: "Print the CRC32, MD5, SHA-1, SHA-256 and SHA-512 of files.",
		Long: `Print the size and the CRC32, MD5, SHA-1, SHA-256 and SHA-512 hashes of every
file in the inputs, going through folders recursively.

To save hashes to a file, use "unifiler checksum create".`,
		Example: `  unifiler file hash a.iso ./photos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "hash")
			return m.logError(m.Hash(flags.Inputs))
		},
	}
	rootCmd.AddCommand(hashCmd)

	packCmd := &cobra.Command{
		Use:   "pack <input>...",
		Short: "Compress files and folders into 7z or RAR archives.",
		Long: `Compress the inputs into the archive --output, or with --separate, each input
into its own archive, <input name>.<format>, in the folder --output. In a
terminal, a form shows the options to review before packing.

Needs 7-Zip (7z) or RAR (rar) on PATH, or their location in the paths.x7z or
paths.rar setting.

Warning: --move and --normalize change the original files. --move deletes the
inputs once they are packed. --normalize, before packing, sets their dates to
2020-02-02 12:00 UTC and, on Windows, sets their Archive attribute and clears
their Read-only, Hidden and System attributes.`,
		Example: `  unifiler file pack ./photos -o photos.7z -l ultra
  unifiler file pack ./a ./b -o ./archives --separate -f rar`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "pack")
			return m.logError(m.Pack(flags.Inputs, flags.Output, flags.Format, flags.Level, flags.Solid, flags.DictSize, flags.Password, flags.Threads, flags.Move, flags.Normalize, flags.Separate))
		},
	}
	packCmd.Flags().IntP("dict", "d", 0, "Dictionary size in MB; 0 uses the archiver's default.")
	packCmd.Flags().StringP("format", "f", "7z", "Archive format: 7z or rar.")
	packCmd.Flags().StringP("level", "l", "normal", "Compression level: none, fast, normal, high or ultra, or 0 to 4.")
	packCmd.Flags().BoolP("move", "m", false, "Delete the inputs once they are packed.")
	packCmd.Flags().BoolP("normalize", "n", false, "Reset the dates and attributes of the inputs before packing; changes the original files.")
	packCmd.Flags().StringP("output", "o", "", "Archive to create, or with --separate, folder to create the archives in. Required.")
	packCmd.Flags().String("password", "", "Password to encrypt the archive with.")
	MarkFlagSensitive(packCmd, "password")
	packCmd.Flags().BoolP("separate", "e", false, "Pack each input into its own archive, named after it, in the --output folder.")
	packCmd.Flags().BoolP("solid", "s", false, "Use solid compression: a better ratio, but slower to extract single files.")
	packCmd.Flags().IntP("threads", "t", 0, "Number of compression threads; 0 lets the archiver decide.")
	rootCmd.AddCommand(packCmd)

	renameCmd := &cobra.Command{
		Use:   "rename <input>...",
		Short: "Rename many files at once with a naming preset.",
		Long: `Rename many files at once, following the naming preset given with --preset.
Folders given as inputs are ignored, and a file whose new name is already taken
is skipped. Before renaming, the old and new names are saved to
unifiler-file-rename-<timestamp>.json in the current folder.

Presets:
  crc32, md4, md5, sha1, sha256, sha512
    Name each file after the hash of its content: <prefix><hash><extension>,
    where <prefix> is the preset name in hex followed by "_", such as
    736861323536_ for sha256.`,
		Example: `  unifiler file rename a.jpg b.jpg -p sha256`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseFileFlags(cmd, args)
			m := NewFileModule(c, "rename")
			return m.logError(m.Rename(flags.Inputs, flags.Preset))
		},
	}
	renameCmd.Flags().StringP("preset", "p", "", "Naming preset: crc32, md4, md5, sha1, sha256 or sha512. Required.")
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
