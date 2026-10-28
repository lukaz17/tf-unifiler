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
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
)

// ChecksumModule handles user requests related checksum file creation and verification.
type ChecksumModule struct {
	logger   zerolog.Logger
	notifier diag.Notifier
}

// Return new ChecksumModule.
func NewChecksumModule(c *Controller, cmdName string) *ChecksumModule {
	return &ChecksumModule{
		logger:   c.CommandLogger("checksum", cmdName),
		notifier: c.Notifier,
	}
}

// Create checksum file(s) for inputs using 1 or many algorithms.
func (m *ChecksumModule) Create(inputs []string, outDir string, algorithms []string, outputName string) error {
	if len(algorithms) == 0 {
		return errors.New("hash algorithm is not specified")
	}
	if err := validateRequiredString(outDir, "outdir"); err != nil {
		return err
	}

	m.logger.Info().
		Strs("algos", algorithms).
		Strs("files", filesys.NormalizePaths(inputs, true)).
		Str("output", filesys.NormalizePath(outDir, true)).
		Msg("Start computing hashes.")

	fhResults, err := listAndHashFiles(inputs, algorithms, true, m.notifier)
	if err != nil {
		return err
	}

	for _, r := range fhResults {
		m.logger.Info().
			Strs("algos", algorithms).
			Str("file", filesys.NormalizePath(r.Entry.RelativePath, true)).
			Int("size", r.Hashes[0].Size).
			Msg("Hashed file.")
	}

	outDirAbs, _ := filepath.Abs(outDir)
	baseName := "checksum"
	isSingleFileInput := func(inps []string) bool {
		if len(inps) != 1 {
			return false
		}
		isFile, err := filesys.IsFile(inps[0])
		if err != nil {
			return false
		}
		return isFile
	}
	if outputName != "" {
		baseName = outputName
	} else if name := filepath.Base(outDir); name != "" && name != "." && name != string(filepath.Separator) {
		baseName = name
	} else if isSingleFileInput(inputs) {
		if name := strings.TrimSuffix(filepath.Base(inputs[0]), filepath.Ext(inputs[0])); name != "" && name != "." {
			baseName = name
		}
	}

	for i, a := range algorithms {
		fContents := []string{}
		for _, r := range fhResults {
			h := r.Hashes[i]
			relPath, _ := filepath.Rel(outDirAbs, r.Entry.AbsolutePath)
			relPath = filesys.NormalizePath(relPath, true)
			var line string
			if a == "crc32" {
				line = fmt.Sprintf("%s %s", relPath, hex.EncodeToString(h.Hash))
			} else {
				line = fmt.Sprintf("%s *%s", hex.EncodeToString(h.Hash), relPath)
			}
			fContents = append(fContents, line)
		}

		ext := a
		if a == "crc32" {
			ext = "sfv"
		}
		oPath := filepath.Join(outDir, fmt.Sprintf("[%s].%s", baseName, ext))
		err := filesys.WriteLines(oPath, fContents)
		if err != nil {
			return err
		}
		m.logger.Info().
			Int("lineCount", len(fContents)).
			Str("path", filesys.NormalizePath(oPath, true)).
			Msg("Written checksum file.")
	}
	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *ChecksumModule) logError(err error) error {
	logProgramError(m.logger, err)
	return err
}

// Define Cobra Command for Checksum module.
func ChecksumCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "checksum",
		Short: "Checksum file processing.",
	}

	createCmd := &cobra.Command{
		Use:   "create <input>...",
		Short: "Create checksum file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := InitApp()
			defer c.Close()
			flags := ParseChecksumFlags(cmd, args)
			m := NewChecksumModule(c, "create")
			return m.logError(m.Create(flags.Inputs, flags.Output, flags.Algorithms, flags.OutputName))
		},
	}
	createCmd.Flags().StringSliceP("algorithm", "a", []string{"sha1"}, "Hash algorithms to use, comma-separated list supported. Supported algorithms: crc32, md4, md5, ripemd160, sha1, sha224, sha256, sha384, sha512.")
	createCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files/Directories to create checksum.")
	createCmd.Flags().StringP("outdir", "o", "", "Directory to store the calculated checksum file(s).")
	createCmd.Flags().StringP("title", "t", "", "Output file name. This will override program auto-naming scheme.")
	rootCmd.AddCommand(createCmd)

	return rootCmd
}

// ChecksumFlags contains all flags used by Checksum module.
type ChecksumFlags struct {
	Algorithms []string
	Inputs     []string
	Output     string
	OutputName string
}

// Extract all flags from a Cobra Command.
func ParseChecksumFlags(cmd *cobra.Command, args []string) *ChecksumFlags {
	algorithms, _ := cmd.Flags().GetStringSlice("algorithm")
	inputs, _ := cmd.Flags().GetStringArray("inputs")
	outdir, _ := cmd.Flags().GetString("outdir")
	output, _ := cmd.Flags().GetString("output")
	if output == "" && outdir != "" {
		output = outdir
	}
	outputName, _ := cmd.Flags().GetString("title")
	inputs = append(args, inputs...)

	return &ChecksumFlags{
		Algorithms: algorithms,
		Inputs:     inputs,
		Output:     output,
		OutputName: outputName,
	}
}
