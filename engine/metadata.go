// Copyright (C) 2025 T-Force I/O
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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforce-io/tf-golib/opx/slicext"
	"github.com/tforce-io/tf-golib/strfmt"
	"github.com/tforceaio/tf-unifiler/core"
	"github.com/tforceaio/tf-unifiler/db"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
)

// MetadataModule handles user requests related to centralized file metadata.
type MetadataModule struct {
	logger   zerolog.Logger
	notifier diag.Notifier
}

// Return new MetadataModule.
func NewMetadataModule(c *Controller, cmdName string) *MetadataModule {
	return &MetadataModule{
		logger:   c.CommandLogger("metadata", cmdName),
		notifier: c.Notifier,
	}
}

// Index whole structure and compute hashes using common algorithms (CRC32, MD5, SHA-1, SHA-256, SHA-512)
// for input (folder) and add them to archive.
func (m *MetadataModule) Index(workspaceDir string, input string, archiveName string, update bool) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}
	if input == "" {
		return errors.New("input is not set")
	}

	// Resolve . and .. to an actual absolute path before any existence check.
	absInput, err := filepath.Abs(input)
	if err != nil {
		return fmt.Errorf("failed to resolve input path: %w", err)
	}

	if !filesys.IsDirectoryExist(absInput) {
		return errors.New("input is not found or is not a directory")
	}

	// Determine archive name: user-supplied value or the directory's base name.
	if archiveName == "" {
		// A root directory has no meaningful base name (Dir equals itself).
		if filepath.Dir(absInput) == absInput {
			return errors.New("cannot determine archive name from root directory, please specify --name")
		}
		archiveName = filepath.Base(absInput)
	}

	m.logger.Info().
		Str("input", absInput).
		Str("name", archiveName).
		Str("workspace", workspaceDir).
		Msg("Start indexing.")

	algos := []string{"crc32", "md5", "sha1", "sha256", "sha512"}
	fhResults, err := listAndHashFiles([]string{absInput}, algos, true, m.notifier)
	if err != nil {
		return err
	}

	hResults := []*core.FileMultiHash{}
	for _, r := range fhResults {
		// Compute path relative to the input directory so ArchiveContent paths are
		// portable and not tied to the machine's absolute directory layout.
		relPath, err := filepath.Rel(absInput, r.Entry.AbsolutePath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %s: %w", r.Entry.AbsolutePath, err)
		}
		m.logger.Info().
			Strs("algos", algos).
			Str("path", relPath).
			Int("size", r.Hashes[0].Size).
			Msg("Hashed file.")
		fileMultiHash := &core.FileMultiHash{
			Crc32:     r.Hashes[0].Hash,
			Md5:       r.Hashes[1].Hash,
			Sha1:      r.Hashes[2].Hash,
			Sha256:    r.Hashes[3].Hash,
			Sha512:    r.Hashes[4].Hash,
			Size:      uint32(r.Hashes[0].Size),
			Directory: filepath.Dir(relPath),
			FileName:  r.Entry.Name,
		}
		hResults = append(hResults, fileMultiHash)
	}

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}

	return m.saveIResults(ctx, archiveName, update, absInput, hResults)
}

// Compute hashes of inputs (files/folders) and refining their contents.
// All files in collections are used by default for matching, onlyObsoleted will use obsoleted files only.
// Invert will match non-existed files in database instead.
// Erase will delete the file directly instead of moving them.
func (m *MetadataModule) Refine(workspaceDir string, inputs, collections []string, onlyObsoleted, invert, erase bool) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}
	if err := validateInputs(inputs); err != nil {
		return err
	}
	m.logger.Info().
		Strs("collections", collections).
		Bool("erase", erase).
		Strs("files", inputs).
		Bool("invert", invert).
		Bool("onlyObsoleted", onlyObsoleted).
		Str("workspace", workspaceDir).
		Msg("Start refining file system.")

	algos := []string{"crc32", "md5", "sha1", "sha256", "sha512"}
	fhResults, err := listAndHashFiles(inputs, algos, true, m.notifier)
	if err != nil {
		return err
	}

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}

	for _, r := range fhResults {
		sha256 := hex.EncodeToString(r.Hashes[3].Hash)
		m.logger.Info().
			Str("crc32", hex.EncodeToString(r.Hashes[0].Hash)).
			Str("md5", hex.EncodeToString(r.Hashes[1].Hash)).
			Str("path", r.Entry.RelativePath).
			Str("sha1", hex.EncodeToString(r.Hashes[2].Hash)).
			Str("sha256", sha256).
			Int("size", r.Hashes[0].Size).
			Msg("Hashed file.")
		metadatas, err := ctx.GetHashesInSets(collections, []string{sha256}, onlyObsoleted)
		if err != nil {
			return err
		}
		noMetadata := len(metadatas) == 0
		if invert == noMetadata {
			newFile := strfmt.NewPathFromStr(r.Entry.AbsolutePath)
			intDir := opx.Ternary(invert, ".extra", ".backup")
			newFile.Parents = append(newFile.Parents, intDir)
			if erase {
				err = os.Remove(r.Entry.AbsolutePath)
			} else {
				err = filesys.CreateDirectoryRecursive(newFile.ParentPath())
				if err != nil {
					return err
				}
				err = os.Rename(r.Entry.AbsolutePath, newFile.FullPath())
			}
			if err != nil {
				return err
			}
			if erase {
				m.logger.Info().
					Str("path", r.Entry.RelativePath).
					Msg("Deleted file.")
			} else {
				m.logger.Info().
					Str("src", r.Entry.RelativePath).
					Str("dest", newFile.FullPath()).
					Msg("Moved file.")
			}
		}
	}

	return nil
}

// Scan and compute hashes using common algorithms (CRC32, MD5, SHA-1, SHA-256, SHA-512) for inputs (files/folders)
// and add them to collection.
// Mark them as obseleted if delete is true.
func (m *MetadataModule) Scan(workspaceDir string, inputs, collections []string, delete bool) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}
	if err := validateInputs(inputs); err != nil {
		return err
	}
	if len(collections) == 0 {
		return errors.New("collections is empty")
	}
	m.logger.Info().
		Strs("collections", collections).
		Bool("delete", delete).
		Strs("files", inputs).
		Str("workspace", workspaceDir).
		Msg("Start scanning files metadata.")

	algos := []string{"crc32", "md5", "sha1", "sha256", "sha512"}
	fhResults, err := listAndHashFiles(inputs, algos, true, m.notifier)
	if err != nil {
		return err
	}

	hResults := []*core.FileMultiHash{}
	for _, r := range fhResults {
		m.logger.Info().
			Strs("algos", algos).
			Str("path", r.Entry.RelativePath).
			Int("size", r.Hashes[0].Size).
			Msg("Hashed file.")
		fileMultiHash := &core.FileMultiHash{
			Crc32:     r.Hashes[0].Hash,
			Md5:       r.Hashes[1].Hash,
			Sha1:      r.Hashes[2].Hash,
			Sha256:    r.Hashes[3].Hash,
			Sha512:    r.Hashes[4].Hash,
			Size:      uint32(r.Hashes[0].Size),
			Directory: filepath.Dir(r.Entry.RelativePath),
			FileName:  r.Entry.Name,
		}
		hResults = append(hResults, fileMultiHash)
	}

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}
	err = m.saveHResults(ctx, hResults, delete, collections)
	if err != nil {
		return err
	}

	return nil
}

// Query Hash data.
func (m *MetadataModule) QueryHash(workspaceDir string, collections, sha256s []string, obsoleted bool) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}

	hashes, err := ctx.GetHashesInSets(collections, sha256s, obsoleted)
	if err != nil {
		return err
	}

	fmt.Println("RESULTS")
	for i, h := range hashes {
		fmt.Println(i+1, h.Sha256, "/", h.Md5, "/", h.Sha1, "/", h.Description)
	}

	return nil
}

// Query Session data.
func (m *MetadataModule) QuerySession(workspaceDir string, sessionID string) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}

	if sessionID == "" {
		sessions, err := ctx.GetSessions()
		if err != nil {
			return err
		}
		fmt.Println("Latest sessions: ")
		for _, s := range sessions {
			fmt.Printf("%s %v\n", s.ID, s.Time)
		}
		return nil
	}

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return err
	}
	session, err := ctx.GetSession(sid)
	if err != nil {
		return err
	}
	if session == nil {
		fmt.Println("Session not found.")
		return nil
	}
	sessionChanges, err := ctx.CountSessionChanges(sid)
	if err != nil {
		return err
	}

	fmt.Println("DETAILS")
	fmt.Println("Time: ", session.Time)
	fmt.Println("-----------------")
	fmt.Println("CHANGES")
	fmt.Println("Hash:    ", sessionChanges.Hash)
	fmt.Println("Mapping: ", sessionChanges.Mapping)
	fmt.Println("Set:     ", sessionChanges.Set)
	fmt.Println("SetHash: ", sessionChanges.SetHash)

	return nil
}

// Query Set data.
func (m *MetadataModule) QuerySet(workspaceDir, setName string) error {
	if err := validateWorkspace(workspaceDir); err != nil {
		return err
	}

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}

	if setName == "" {
		sets, err := ctx.GetSetsByNames([]string{})
		if err != nil {
			return err
		}
		fmt.Println("ALL SETS: ")
		for _, s := range sets {
			fmt.Printf("%s %v\n", s.ID, s.Name)
		}
		return nil
	}

	sets, err := ctx.GetSetsByNames([]string{setName})
	if err != nil {
		return err
	}
	if len(sets) == 0 {
		fmt.Println("Set not found.")
		return nil
	}
	hashes, err := ctx.GetHashesInSets([]string{setName}, []string{}, false)
	if err != nil {
		return err
	}

	fmt.Println("DETAILS")
	fmt.Println("ID: ", sets[0].ID)
	fmt.Println("-----------------")
	fmt.Println("RESULTS")
	for i, h := range hashes {
		fmt.Println(i+1, h.Sha256, "/", h.Md5, "/", h.Sha1, "/", h.Description)
	}

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *MetadataModule) logError(err error) {
	logProgramError(m.logger, err)
}

// Save indexing results to metadata database along with their respective archives.
func (m *MetadataModule) saveIResults(ctx *db.DbContext, archiveName string, update bool, directory string, hResults []*core.FileMultiHash) error {
	sessionID, err := uuid.NewV7()
	if err != nil {
		m.logger.Info().Msg("Failed to generate SessionID.")
		return err
	}
	// Save Session
	session := db.NewSession(sessionID, time.Now().UTC())
	err = ctx.SaveSessions([]*db.Session{session})
	if err != nil {
		m.logger.Info().Msg("Failed to save Session.")
		return err
	}
	// Save Hashes
	hashes := make([]*db.Hash, len(hResults))
	for i, res := range hResults {
		hashes[i] = db.NewHash(res, false)
		hashes[i].SessionID = sessionID
	}
	err = ctx.SaveHashes(hashes)
	if err != nil {
		m.logger.Info().Msg("Failed to save Hashes.")
		return err
	}
	// Reload Hashes
	sha256s := make([]string, len(hResults))
	for i, res := range hResults {
		sha256s[i] = res.Sha256.HexStr()
	}
	hashes, err = ctx.GetHashesBySha256s(sha256s)
	if err != nil {
		m.logger.Info().Msg("Failed to reload Hashes.")
		return err
	}
	hashesMap := map[string]db.Bytes32{}
	for _, hash := range hashes {
		hashesMap[hash.Sha256] = hash.ID
	}
	// Save Mappings
	mappings := make([]*db.Mapping, len(hResults))
	for i, res := range hResults {
		fileName := strfmt.NewFileNameFromStr(res.FileName)
		mappings[i] = db.NewMapping(hashesMap[res.Sha256.HexStr()], directory, fileName.Name, fileName.Extension)
		mappings[i].SessionID = sessionID
	}
	err = ctx.SaveMappings(mappings)
	if err != nil {
		m.logger.Info().Msg("Failed to save Mappings.")
		return err
	}
	// Validate Archive name
	existingArchive, err := ctx.GetArchiveByName(archiveName)
	if err != nil {
		m.logger.Info().Msg("Failed to get Archive.")
		return err
	}
	if existingArchive != nil && !update {
		return fmt.Errorf("archive %q already exists, use --update to add contents to it", archiveName)
	}
	// Save Archive
	archive := db.NewArchive(archiveName)
	archive.SessionID = sessionID
	err = ctx.SaveArchives([]*db.Archive{archive})
	if err != nil {
		m.logger.Info().Msg("Failed to save Archive.")
		return err
	}
	// Reload Archive
	archive, err = ctx.GetArchiveByName(archiveName)
	if err != nil {
		m.logger.Info().Msg("Failed to reload Archive.")
		return err
	}
	// Save ArchiveContents
	archiveContents := make([]*db.ArchiveContent, len(hResults))
	for i, res := range hResults {
		fileName := strfmt.NewFileNameFromStr(res.FileName)
		archiveContents[i] = db.NewArchiveContent(archive.ID, res.Directory, fileName.Name, fileName.Extension, hashesMap[res.Sha256.HexStr()])
		archiveContents[i].SessionID = sessionID
	}
	err = ctx.SaveArchiveContents(archiveContents)
	if err != nil {
		m.logger.Info().Msg("Failed to save ArchiveContents.")
		return err
	}

	m.logger.Info().Msg("Saved archive metadata successfully.")
	return nil
}

// Save hashing results to metadata database along with their respective collections.
func (m *MetadataModule) saveHResults(ctx *db.DbContext, hResults []*core.FileMultiHash, ignore bool, collections []string) (err error) {
	sessionID, err := uuid.NewV7()
	if err != nil {
		m.logger.Info().Msg("Failed to generate SessionID.")
		return err
	}
	// save Session
	session := db.NewSession(sessionID, time.Now().UTC())
	err = ctx.SaveSessions([]*db.Session{session})
	if err != nil {
		m.logger.Info().Msg("Failed to save Sessions.")
		return err
	}
	// save Hash
	hashes := make([]*db.Hash, len(hResults))
	for i, res := range hResults {
		hashes[i] = db.NewHash(res, ignore)
		hashes[i].SessionID = sessionID
	}
	err = ctx.SaveHashes(hashes)
	if err != nil {
		m.logger.Info().Msg("Failed to save Hashes.")
		return err
	}
	// save Mapping
	sha256s := make([]string, len(hResults))
	for i, res := range hResults {
		sha256s[i] = res.Sha256.HexStr()
	}
	hashes, err = ctx.GetHashesBySha256s(sha256s)
	if err != nil {
		m.logger.Info().Msg("Failed to reload Hashes.")
		return err
	}
	hashesMap := map[string]db.Bytes32{}
	for _, hash := range hashes {
		hashesMap[hash.Sha256] = hash.ID
	}
	mappings := make([]*db.Mapping, len(hResults))
	for i, res := range hResults {
		fileName := strfmt.NewFileNameFromStr(res.FileName)
		mappings[i] = db.NewMapping(hashesMap[res.Sha256.HexStr()], res.Directory, fileName.Name, fileName.Extension)
		mappings[i].SessionID = sessionID
	}
	err = ctx.SaveMappings(mappings)
	if err != nil {
		m.logger.Info().Msg("Failed to save Mappings.")
		return err
	}
	if !slicext.IsEmpty(collections) {
		// save Set
		sets := make([]*db.Set, len(collections))
		for i, name := range collections {
			sets[i] = db.NewSet(name)
			sets[i].SessionID = sessionID
		}
		err = ctx.SaveSets(sets)
		if err != nil {
			m.logger.Info().Msg("Failed to save Sets.")
			return err
		}
		// save SetHash
		sets, err = ctx.GetSetsByNames(collections)
		if err != nil {
			m.logger.Info().Msg("Failed to reload Sets.")
			return err
		}
		setHashes := make([]*db.SetHash, len(sets)*len(hashes))
		for i, set := range sets {
			hashesLen := len(hashes)
			for j, hash := range hashes {
				setHashes[i*hashesLen+j] = db.NewSetHash(set.ID, hash.ID)
				setHashes[i*hashesLen+j].SessionID = sessionID
			}
		}
		err = ctx.SaveSetHashes(setHashes)
		if err != nil {
			m.logger.Info().Msg("Failed to save SetHashes.")
			return err
		}
	}

	m.logger.Info().Msg("Saved metadata successfully.")
	return err
}

// Return database path to store Metadata module's ouputs inside Unifiler workspace.
func MetadataWorkspaceDatabase(workspaceDir string) string {
	return filepath.Join(workspaceDir, "metadata.db")
}

// Define Cobra Command for Metadata module.
func MetadataCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "metadata",
		Short: "Centralized file metadata database.",
	}
	rootCmd.PersistentFlags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")

	indexCmd := &cobra.Command{
		Use:   "index <input>",
		Short: "Index input's content.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd, args)
			m := NewMetadataModule(c, "archive")
			input := ""
			if len(flags.Inputs) > 0 {
				input = flags.Inputs[0]
			}
			m.logError(m.Index(flags.WorkspaceDir, input, flags.Name, flags.Update))
		},
	}
	indexCmd.Flags().StringArrayP("inputs", "i", []string{}, "Directory to archive.")
	indexCmd.Flags().StringP("name", "n", "", "Name for the archive in database (Defaults to directory name)")
	indexCmd.Flags().Bool("update", false, "Allow update current archive if exists.")
	rootCmd.AddCommand(indexCmd)

	refineCmd := &cobra.Command{
		Use:   "refine <input>...",
		Short: "Refine inputs against metadata database.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd, args)
			m := NewMetadataModule(c, "scan")
			m.logError(m.Refine(flags.WorkspaceDir, flags.Inputs, flags.Collections, flags.OnlyObsoleted, flags.Invert, flags.Erase))
		},
	}
	refineCmd.Flags().StringSliceP("collections", "c", []string{}, "Names of collections of known files, comma-separated list supported.")
	refineCmd.Flags().Bool("erase", false, "Force delete matched files instead of moving.")
	refineCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files/Directories to refine.")
	refineCmd.Flags().Bool("invert", false, "Take action on non-matched files instead of matched ones.")
	refineCmd.Flags().BoolP("obsoleted", "o", false, "Only match obsoleted files.")
	rootCmd.AddCommand(refineCmd)

	scanCmd := &cobra.Command{
		Use:   "scan <input>...",
		Short: "Scan inputs for file metadata.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd, args)
			m := NewMetadataModule(c, "scan")
			m.logError(m.Scan(flags.WorkspaceDir, flags.Inputs, flags.Collections, flags.Deleted))
		},
	}
	scanCmd.Flags().StringSliceP("collections", "c", []string{}, "Names of collections of known files, comma-separated list supported. If a collection existed, files will be appended to that collection.")
	scanCmd.Flags().Bool("delete", false, "Mark the inputs as obsoleted.")
	scanCmd.Flags().StringArrayP("inputs", "i", []string{}, "Files/Directories to hash.")
	rootCmd.AddCommand(scanCmd)

	rootCmd.AddCommand(metadataQueryCmd())

	return rootCmd
}

func metadataQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Query metadata database.",
	}

	hashCmd := &cobra.Command{
		Use:   "hash",
		Short: "Query hash information.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd, nil)
			m := NewMetadataModule(c, "query_hash")
			m.logError(m.QueryHash(flags.WorkspaceDir, flags.Collections, flags.Hashes, flags.OnlyObsoleted))
		},
	}
	hashCmd.Flags().StringSliceP("collections", "c", []string{}, "Names of collections of known files, comma-separated list supported.")
	hashCmd.Flags().StringSliceP("hashes", "v", []string{}, "SHA-256s of known files, comma-separated list supported.")
	hashCmd.Flags().BoolP("obsoleted", "o", false, "Only match obsoleted files.")
	queryCmd.AddCommand(hashCmd)

	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Query session information.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd, nil)
			m := NewMetadataModule(c, "query_session")
			m.logError(m.QuerySession(flags.WorkspaceDir, flags.ID))
		},
	}
	sessionCmd.Flags().StringP("id", "i", "", "Session ID.")
	queryCmd.AddCommand(sessionCmd)

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Query collection information.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd, nil)
			m := NewMetadataModule(c, "query_set")
			m.logError(m.QuerySet(flags.WorkspaceDir, flags.Name))
		},
	}
	setCmd.Flags().StringP("name", "n", "", "Collection name.")
	queryCmd.AddCommand(setCmd)

	return queryCmd
}

// Struct MetadataFlags contains all flags used by Metadata module.
type MetadataFlags struct {
	Collections   []string
	Deleted       bool
	Erase         bool
	Hashes        []string
	ID            string
	Inputs        []string
	Invert        bool
	Name          string
	OnlyObsoleted bool
	Update        bool
	WorkspaceDir  string
}

// Extract all flags from a Cobra Command.
func ParseMetadataFlags(cmd *cobra.Command, args []string) *MetadataFlags {
	collections, _ := cmd.Flags().GetStringSlice("collections")
	deleted, _ := cmd.Flags().GetBool("deleted")
	erase, _ := cmd.Flags().GetBool("erase")
	hashes, _ := cmd.Flags().GetStringSlice("hashes")
	id, _ := cmd.Flags().GetString("id")
	inputs, _ := cmd.Flags().GetStringArray("inputs")
	invert, _ := cmd.Flags().GetBool("invert")
	name, _ := cmd.Flags().GetString("name")
	obsoleted, _ := cmd.Flags().GetBool("obsoleted")
	update, _ := cmd.Flags().GetBool("update")
	workspaceDir, _ := cmd.Flags().GetString("workspace")
	inputs = append(args, inputs...)

	return &MetadataFlags{
		Collections:   collections,
		Deleted:       deleted,
		Erase:         erase,
		Hashes:        hashes,
		ID:            id,
		Inputs:        inputs,
		Invert:        invert,
		Name:          name,
		OnlyObsoleted: obsoleted,
		Update:        update,
		WorkspaceDir:  workspaceDir,
	}
}
