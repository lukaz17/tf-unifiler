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
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/encoding/checksum"
	"github.com/tforceaio/tf-unifiler/filesys"
)

// FileMirrorMapping stores old and new filename after mirroring for rollback.
type FileMirrorMapping struct {
	Source string `json:"s,omitempty"`
	Hash   string `json:"h,omitempty"`
}

// MirrorModule handles user requests related to centralized file vaulting.
type MirrorModule struct {
	logger   zerolog.Logger
	notifier diag.Notifier
}

// Return new MirrorModule.
func NewMirrorModule(c *Controller, cmdName string) *MirrorModule {
	return &MirrorModule{
		logger:   c.CommandLogger("mirror", cmdName),
		notifier: c.Notifier,
	}
}

// Create file structure in targetDir using a checksumFile.
func (m *MirrorModule) Export(workspaceDir, checksumFile, targetDir string) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}
	if err := validateInput(checksumFile, "checksum file"); err != nil {
		return err
	}
	m.logger.Info().
		Str("cache", filesys.NormalizePath(workspaceDir, true)).
		Str("checksum", filesys.NormalizePath(checksumFile, true)).
		Str("root", filesys.NormalizePath(targetDir, true)).
		Msgf("Start exporting files structure.")

	workspaceRoot := MirrorWorkspaceRoot(workspaceDir)
	checksumReader, err := os.OpenFile(checksumFile, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	items, err := checksum.ParseSha256(checksumReader)
	if err != nil {
		return err
	}
	defer checksumReader.Close()

	missingItems := []string{}
	for _, l := range items {
		cachePath := filepath.Join(workspaceRoot, l.Hash)
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
		targetRoot, _ = filepath.Split(checksumPath)
	} else {
		targetRoot, _ = filesys.GetAbsPath(targetDir)
	}
	if filesys.IsFileExist(targetRoot) {
		return errors.New("a file with same name with target root existed")
	}
	for _, l := range items {
		cachePath := filepath.Join(workspaceRoot, l.Hash)
		targetPath := opx.Ternary(filesys.IsAbsPath(l.Path), l.Path, filepath.Join(targetRoot, l.Path))
		err := filesys.CreateHardlink(cachePath, targetPath)
		if err != nil {
			m.logger.Info().
				Str("src", filesys.NormalizePath(cachePath, true)).
				Str("dest", filesys.NormalizePath(targetPath, true)).
				Msg("Failed to create hardlink.")
			return err
		} else {
			m.logger.Info().
				Str("hash", l.Hash).
				Str("src", filesys.NormalizePath(cachePath, true)).
				Str("dest", filesys.NormalizePath(targetPath, true)).
				Msg("Exported file.")
		}
	}
	return nil
}

// Scan and calculate SHA-256 hashes for inputs (files/folders),
// then create hardlink to workspaceDir.
func (m *MirrorModule) Scan(workspaceDir string, inputs []string) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}
	if err := validateInputs(inputs); err != nil {
		return err
	}
	m.logger.Info().
		Str("cache", filesys.NormalizePath(workspaceDir, true)).
		Strs("inputs", filesys.NormalizePaths(inputs, true)).
		Msg("Start scanning files")

	workspaceRoot := MirrorWorkspaceRoot(workspaceDir)
	fhResults, err := listAndHashFiles(inputs, []string{"sha256"}, true, m.notifier)
	if err != nil {
		return err
	}

	for _, r := range fhResults {
		m.logger.Info().
			Str("algo", "sha256").
			Str("path", filesys.NormalizePath(r.Entry.RelativePath, true)).
			Int64("size", r.Hashes[0].Size).
			Msg("Hashed file.")
	}

	mappings := []*FileMirrorMapping{}
	for _, r := range fhResults {
		mapping := &FileMirrorMapping{
			Source: r.Entry.AbsolutePath,
			Hash:   hex.EncodeToString(r.Hashes[0].Hash),
		}
		mappings = append(mappings, mapping)
	}
	for _, r := range fhResults {
		name := hex.EncodeToString(r.Hashes[0].Hash)
		cachePath := filepath.Join(workspaceRoot, name)
		if filesys.IsFileExist(cachePath) {
			m.logger.Info().
				Str("src", filesys.NormalizePath(r.Entry.AbsolutePath, true)).
				Str("cache", filesys.NormalizePath(cachePath, true)).
				Msg("Skipped. File is already cached.")
		} else {
			err := filesys.CreateHardlink(r.Entry.AbsolutePath, cachePath)
			if err != nil {
				m.logger.Info().
					Str("src", filesys.NormalizePath(r.Entry.AbsolutePath, true)).
					Str("dest", filesys.NormalizePath(cachePath, true)).
					Msg("Failed to create hardlink.")
				return err
			}
			m.logger.Info().
				Str("src", filesys.NormalizePath(r.Entry.AbsolutePath, true)).
				Str("target", filesys.NormalizePath(cachePath, true)).
				Msg("Created cache file.")
		}
	}

	rollbackFilePath, err := writeJSON(workspaceRoot, "mirror-", mappings)
	if err != nil {
		m.logger.Info().
			Str("path", filesys.NormalizePath(rollbackFilePath, true)).
			Msg("Failed to write rollback file.")
		return err
	}
	m.logger.Info().
		Str("path", filesys.NormalizePath(rollbackFilePath, true)).
		Msg("Written rollback file.")

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *MirrorModule) logError(err error) error {
	return LogCommandError(m.logger, err)
}

// Return directory path to store Mirror module's outputs inside Unifiler workspace.
func MirrorWorkspaceRoot(workspaceDir string) string {
	return filepath.Join(workspaceDir, "mirror")
}

// Define Cobra Command for Mirror module.
func MirrorCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mirror",
		Short: "Manage a content-addressable store of files.",
		Long: `Manage a content-addressable store, the mirror folder of the workspace given
with -w, which keeps each distinct file once, named by the SHA-256 of its
content. Files are added to the store, and folders rebuilt from it, as hard
links, so every copy of a file shares the same data on disk.

The workspace folder must already exist; "unifiler metadata" can share it.

Hard links cannot cross drives, so the store and the files must be on the same
drive. Editing a linked file in place changes every copy of it, the one in the
store included.`,
	}
	MarkCommandQueueable(rootCmd)
	rootCmd.PersistentFlags().StringP("workspace", "w", "", "Workspace folder, whose mirror folder is the store; it must exist. Required.")

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Rebuild a folder tree from the store using a checksum file.",
		Long: `Link each file listed in the --checksum file, such as one made by
"unifiler checksum create -a sha256", from the store into --output, at the
path the checksum file lists for it. Nothing is linked when the store lacks a
listed hash.`,
		Example: `  unifiler mirror export -w ./ws -i "./photos/[photos].sha256" -o ./restored`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseMirrorFlags(cmd, []string{})
			m := NewMirrorModule(c, "export")
			return m.logError(m.Export(flags.WorkspaceDir, flags.ChecksumFile, flags.Output))
		},
	}
	exportCmd.Flags().StringP("checksum", "i", "", "SHA-256 checksum file listing the files to rebuild. Required.")
	exportCmd.Flags().StringP("output", "o", "", "Folder to rebuild the files in; the checksum file's folder when empty.")
	rootCmd.AddCommand(exportCmd)

	scanCmd := &cobra.Command{
		Use:   "scan <input>...",
		Short: "Add files and folders to the store.",
		Long: `Hash every file in the inputs, going through folders recursively, and link it
into the store as <sha256>, unless the store already has that content. Each
file's path and SHA-256 are saved to mirror-<timestamp>.json in the store.`,
		Example: `  unifiler mirror scan ./photos -w ./ws`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseMirrorFlags(cmd, args)
			m := NewMirrorModule(c, "scan")
			return m.logError(m.Scan(flags.WorkspaceDir, flags.Inputs))
		},
	}
	scanCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files or folders to add, in addition to the arguments; repeat for several.")
	rootCmd.AddCommand(scanCmd)

	return rootCmd
}

// MirrorFlags contains all flags used by Mirror module.
type MirrorFlags struct {
	ChecksumFile string
	Inputs       []string
	Output       string
	WorkspaceDir string
}

// Extract all flags from a Cobra Command.
func ParseMirrorFlags(cmd *cobra.Command, args []string) *MirrorFlags {
	checksumFile, _ := cmd.Flags().GetString("checksum")
	inputs, _ := cmd.Flags().GetStringArray("inputs")
	output, _ := cmd.Flags().GetString("output")
	workspaceDir, _ := cmd.Flags().GetString("workspace")
	inputs = append(args, inputs...)

	return &MirrorFlags{
		ChecksumFile: checksumFile,
		Inputs:       inputs,
		Output:       output,
		WorkspaceDir: workspaceDir,
	}
}
