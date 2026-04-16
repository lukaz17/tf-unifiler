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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/filesys"
)

// ChecksumModule handles user requests related checksum file creation and verification.
type ChecksumModule struct {
	logger zerolog.Logger
}

// Return new ChecksumModule.
func NewChecksumModule(c *Controller, cmdName string) *ChecksumModule {
	return &ChecksumModule{
		logger: c.CommandLogger("checksum", cmdName),
	}
}

// Create checksum file(s) for inputs using 1 or many algorithms.
func (m *ChecksumModule) Create(inputs []string, output string, algorithms []string) error {
	if len(algorithms) == 0 {
		return errors.New("hash algorithm is not specified")
	}

	m.logger.Info().
		Strs("algos", algorithms).
		Strs("files", inputs).
		Str("output", output).
		Msg("Start computing hashes.")

	contents, err := filesys.List(inputs, true)
	if err != nil {
		return err
	}

	hResults := []*hasher.HashResult{}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, algorithms)
		m.logger.Info().
			Strs("algos", algorithms).
			Str("file", c.RelativePath).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
		if err != nil {
			return err
		}
		hResults = append(hResults, fhResults...)
	}

	for _, a := range algorithms {
		fContents := []string{}
		for _, r := range hResults {
			if a == r.Algorithm {
				line := fmt.Sprintf("%s *%s", hex.EncodeToString(r.Hash), r.Path)
				fContents = append(fContents, line)
			}
		}

		outputInternal := opx.Ternary(output == "", "checksum", output)
		// substitute file extension. for more information: https://go.dev/play/p/0wZcne8ZC8G
		oPath := fmt.Sprintf("%s.%s", strings.TrimSuffix(outputInternal, filepath.Ext(outputInternal)), a)
		err := filesys.WriteLines(oPath, fContents)
		if err != nil {
			return err
		}
		m.logger.Info().
			Int("lineCount", len(fContents)).
			Str("path", oPath).
			Msg("Written checksum file.")
	}
	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *ChecksumModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

// Define Cobra Command for Checksum module.
func ChecksumCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "checksum",
		Short: "Create and verify checksum files.",
	}

	createCmd := &cobra.Command{
		Use:   "create <input>...",
		Short: "Create checksum file(s) using 1 or many hash algorithms.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseChecksumFlags(cmd, args)
			m := NewChecksumModule(c, "create")
			m.logError(m.Create(flags.Inputs, flags.Output, flags.Algorithms))
		},
	}
	createCmd.Flags().StringSliceP("algo", "a", []string{"sha1"}, "Hash algorithms to use, comma-separated list supported. Supported algorithms: md4, md5, ripemd160, sha1, sha224, sha256, sha384, sha512.")
	createCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files/Directories to create checksum.")
	createCmd.Flags().StringP("output", "o", "", "Directory to store the calculated checksum file(s).")
	createCmd.Flags().StringP("title", "t", "", "Output file name. This will override program smart naming scheme.")
	rootCmd.AddCommand(createCmd)

	return rootCmd
}

// Struct ChecksumFlags contains all flags used by Checksum module.
type ChecksumFlags struct {
	Algorithms []string
	Inputs     []string
	Output     string
	OutputName string
}

// Extract all flags from a Cobra Command.
func ParseChecksumFlags(cmd *cobra.Command, args []string) *ChecksumFlags {
	algorithms, _ := cmd.Flags().GetStringSlice("algo")
	inputs, _ := cmd.Flags().GetStringArray("inputs")
	output, _ := cmd.Flags().GetString("output")
	outputName, _ := cmd.Flags().GetString("title")
	inputs = append(args, inputs...)

	return &ChecksumFlags{
		Algorithms: algorithms,
		Inputs:     inputs,
		Output:     output,
		OutputName: outputName,
	}
}
