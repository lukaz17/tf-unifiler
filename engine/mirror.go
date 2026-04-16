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
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/filesys"
	"github.com/tforceaio/tf-unifiler-go/parser"
)

// Struct FileMirrorMapping stores old and new filename after mirroring for rollback.
type FileMirrorMapping struct {
	Source string `json:"s,omitempty"`
	Hash   string `json:"h,omitempty"`
}

// MirrorModule handles user requests related to file centralization feature.
type MirrorModule struct {
	logger zerolog.Logger
}

// Return new MirrorModule.
func NewMirrorModule(c *Controller, cmdName string) *MirrorModule {
	return &MirrorModule{
		logger: c.CommandLogger("mirror", cmdName),
	}
}

// Create file structure in targetDir using a checksumFile.
func (m *MirrorModule) Export(workspaceDir, checksumFile, targetDir string) error {
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
	}
	if checksumFile == "" {
		return errors.New("checksum file is not set")
	} else if !filesys.IsFileExist(checksumFile) {
		return errors.New("checksum file is not found")
	}
	m.logger.Info().
		Str("cache", workspaceDir).
		Str("checksum", checksumFile).
		Str("root", targetDir).
		Msgf("Start exporting files structure.")

	workspaceRoot := MirrorWorkspaceRoot(workspaceDir)
	checksumReader, err := os.OpenFile(checksumFile, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	items, err := parser.ParseSha256(checksumReader)
	if err != nil {
		return err
	}
	defer checksumReader.Close()

	missingItems := []string{}
	for _, l := range items {
		cachePath := path.Join(workspaceRoot, l.Hash)
		if !filesys.IsFileExist(cachePath) {
			missingItems = append(missingItems, l.Hash)
		}
	}
	if len(missingItems) > 0 {
		m.logger.Warn().
			Strs("hashes", missingItems).
			Msg("Items are not found in workspace.")
		return errors.New("missing items in workspace")
	}

	if targetDir == "" {
		m.logger.Warn().Msg("Target path is not specified, it will be derived from checksum file path instead.")
	}
	targetRoot := targetDir
	if targetRoot == "" {
		checksumPath, _ := filesys.GetAbsPath(checksumFile)
		targetRoot, _ = path.Split(checksumPath)
	} else {
		targetRoot, _ = filesys.GetAbsPath(targetDir)
	}
	if filesys.IsFileExist(targetRoot) {
		return errors.New("a file with same name with target root existed")
	}
	for _, l := range items {
		cachePath := path.Join(workspaceRoot, l.Hash)
		targetPath := opx.Ternary(filesys.IsAbsPath(l.Path), l.Path, path.Join(targetRoot, l.Path))
		err := filesys.CreateHardlink(cachePath, targetPath)
		if err != nil {
			m.logger.Info().
				Str("src", cachePath).
				Str("dest", targetPath).
				Msg("Failed to create hardlink.")
			return err
		} else {
			m.logger.Info().
				Str("hash", l.Hash).
				Str("src", cachePath).
				Str("dest", targetPath).
				Msg("Exported file.")
		}
	}
	return nil
}

// Scan and calculate SHA-256 hashes for inputs (files/folders),
// then create hardlink to workspaceDir.
func (m *MirrorModule) Scan(workspaceDir string, inputs []string) error {
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
	}
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	m.logger.Info().
		Str("cache", workspaceDir).
		Strs("inputs", inputs).
		Msg("Start scanning files")

	workspaceRoot := MirrorWorkspaceRoot(workspaceDir)
	contents, err := filesys.List(inputs, true)
	if err != nil {
		return err
	}

	hResults := []*hasher.HashResult{}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResult, err := hasher.HashSha256(c.RelativePath)
		if err != nil {
			m.logger.Info().
				Str("path", c.RelativePath).
				Msg("Failed to compute hash.")
			return err
		}
		m.logger.Info().
			Str("algo", "sha256").
			Str("path", c.RelativePath).
			Int("size", fhResult.Size).
			Msg("Hashed file.")
		fhResult.Path = c.AbsolutePath
		hResults = append(hResults, fhResult)
	}

	mappings := []*FileMirrorMapping{}
	for _, e := range hResults {
		mapping := &FileMirrorMapping{
			Source: e.Path,
			Hash:   hex.EncodeToString(e.Hash),
		}
		mappings = append(mappings, mapping)
	}
	for _, r := range hResults {
		name := hex.EncodeToString(r.Hash)
		cachePath := path.Join(workspaceRoot, name)
		if filesys.IsFileExist(cachePath) {
			m.logger.Info().
				Str("src", r.Path).
				Str("cache", cachePath).
				Msg("Skipped. File is already cached.")
		} else {
			err := filesys.CreateHardlink(r.Path, cachePath)
			if err != nil {
				m.logger.Info().
					Str("src", r.Path).
					Str("dest", cachePath).
					Msg("Failed to create hardlink.")
				return err
			}
			m.logger.Info().
				Str("src", r.Path).
				Str("target", cachePath).
				Msg("Created cache file.")
		}
	}

	currentTimestamp := time.Now().UnixMilli()
	rollbackFilePath := filesys.Join(workspaceRoot, "mirror-"+strconv.FormatInt(currentTimestamp, 10)+".json")
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

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *MirrorModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msgf("Unexpected error has occurred. Program will exit.")
	}
}

// Return directory path to store Mirror module's ouputs inside Unifiler workspace.
func MirrorWorkspaceRoot(workspaceDir string) string {
	return filepath.Join(workspaceDir, ".unifiler", "mirror")
}

// Define Cobra Command for Mirror module.
func MirrorCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mirror",
		Short: "Centralize and save disk space by utilizing hard link feature in supported file system.",
	}

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Create file structure from checksum file.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMirrorFlags(cmd)
			m := NewMirrorModule(c, "export")
			m.logError(m.Export(flags.WorkspaceDir, flags.ChecksumFile, flags.Output))
		},
	}
	exportCmd.Flags().StringP("checksum", "i", "", "Checksum file path. Only supported SHA-256.")
	exportCmd.Flags().StringP("output", "o", "", "Directory where the files will be exported.")
	exportCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
	rootCmd.AddCommand(exportCmd)

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan and compute hashes files/directories then create hardlink to workspace.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMirrorFlags(cmd)
			m := NewMirrorModule(c, "export")
			m.logError(m.Scan(flags.WorkspaceDir, flags.Inputs))
		},
	}
	scanCmd.Flags().StringSliceP("inputs", "i", []string{}, "Files/Directories to import.")
	scanCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
	rootCmd.AddCommand(scanCmd)

	return rootCmd
}

// Struct MirrorFlags contains all flags used by Mirror module.
type MirrorFlags struct {
	ChecksumFile string
	Inputs       []string
	Output       string
	WorkspaceDir string
}

// Extract all flags from a Cobra Command.
func ParseMirrorFlags(cmd *cobra.Command) *MirrorFlags {
	checksumFile, _ := cmd.Flags().GetString("checksum")
	inputs, _ := cmd.Flags().GetStringSlice("inputs")
	output, _ := cmd.Flags().GetString("output")
	workspaceDir, _ := cmd.Flags().GetString("workspace")

	return &MirrorFlags{
		ChecksumFile: checksumFile,
		Inputs:       inputs,
		Output:       output,
		WorkspaceDir: workspaceDir,
	}
}
