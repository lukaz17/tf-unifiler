// Copyright (C) 2025 T-Force I/O
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
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforce-io/tf-golib/opx/slicext"
	"github.com/tforce-io/tf-golib/strfmt"
	"github.com/tforceaio/tf-unifiler-go/core"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/db"
	"github.com/tforceaio/tf-unifiler-go/filesys"
)

// MetadataModule handles user requests related file hashes.
type MetadataModule struct {
	logger zerolog.Logger
}

// Return new MetadataModule.
func NewMetadataModule(c *Controller, cmdName string) *MetadataModule {
	return &MetadataModule{
		logger: c.CommandLogger("metadata", cmdName),
	}
}

// Compute hashes of inputs (files/folders) and refining their contents.
// All files in collections are used by default for matching, onlyObsoleted will use obsoleted files only.
// Invert will match non-existed files in database instead.
// Erase will delete the file directly instead of moving them.
func (m *MetadataModule) Refine(workspaceDir string, inputs, collections []string, onlyObsoleted, invert, erase bool) error {
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
	}
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	m.logger.Info().
		Strs("collections", collections).
		Bool("erase", erase).
		Strs("files", inputs).
		Bool("invert", invert).
		Bool("onlyObsoleted", onlyObsoleted).
		Str("workspace", workspaceDir).
		Msg("Start refining file system.")

	contents, err := filesys.List(inputs, true)
	if err != nil {
		return err
	}
	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
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
		sha256 := hex.EncodeToString(fhResults[2].Hash)
		m.logger.Info().
			Str("md5", hex.EncodeToString(fhResults[0].Hash)).
			Str("path", c.RelativePath).
			Str("sha1", hex.EncodeToString(fhResults[1].Hash)).
			Str("sha256", sha256).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
		metadatas, err := ctx.GetHashesInSets(collections, []string{sha256}, onlyObsoleted)
		if err != nil {
			return err
		}
		noMetadata := len(metadatas) == 0
		if invert == noMetadata {
			newFile := strfmt.NewPathFromStr(c.AbsolutePath)
			intDir := opx.Ternary(invert, ".extra", ".backup")
			newFile.Parents = append(newFile.Parents, intDir)
			if erase {
				err = os.Remove(c.AbsolutePath)
			} else {
				err = filesys.CreateDirectoryRecursive(newFile.ParentPath())
				if err != nil {
					return err
				}
				err = os.Rename(c.AbsolutePath, newFile.FullPath())
			}
			if err != nil {
				return err
			}
			if erase {
				m.logger.Info().
					Str("path", c.RelativePath).
					Msg("Deleted file.")
			} else {
				m.logger.Info().
					Str("src", c.RelativePath).
					Str("dest", newFile.FullPath()).
					Msg("Moved file.")
			}
		}
	}

	return nil
}

// Scan and compute hashes using common algorithms (MD5, SHA-1, SHA-256, SHA-512) for inputs (files/folders)
// and add them to collection.
// Mark them as obseleted if delete is true.
func (m *MetadataModule) Scan(workspaceDir string, inputs, collections []string, delete bool) error {
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
	}
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
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

	contents, err := filesys.List(inputs, true)
	if err != nil {
		return err
	}

	hResults := []*core.FileMultiHash{}
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
			Strs("algos", algos).
			Str("path", c.RelativePath).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
		fileMultiHash := &core.FileMultiHash{
			Md5:      fhResults[0].Hash,
			Sha1:     fhResults[1].Hash,
			Sha256:   fhResults[2].Hash,
			Sha512:   fhResults[3].Hash,
			Size:     uint32(fhResults[0].Size),
			FileName: c.Name,
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
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
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
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
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
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesys.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
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
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
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
	hashesMap := map[string]uuid.UUID{}
	for _, hash := range hashes {
		hashesMap[hash.Sha256] = hash.ID
	}
	mappings := make([]*db.Mapping, len(hResults))
	for i, res := range hResults {
		fileName := strfmt.NewFileNameFromStr(res.FileName)
		mappings[i] = db.NewMapping(hashesMap[res.Sha256.HexStr()], fileName.Name, fileName.Extension)
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
	return filepath.Join(workspaceDir, ".unifiler", "metadata.db")
}

// Define Cobra Command for Metadata module.
func MetadataCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "metadata",
		Short: "File metadata management and mass file update using metadata.",
	}

	refineCmd := &cobra.Command{
		Use:   "refine <input>...",
		Short: "Refine file system contents by metadata.",
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
	refineCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
	rootCmd.AddCommand(refineCmd)

	scanCmd := &cobra.Command{
		Use:   "scan <input>...",
		Short: "Compute hashes for files using common algorithms (MD5, SHA-1, SHA-256, SHA-512) and persist them.",
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
	scanCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
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
	hashCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
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
	sessionCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
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
	setCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
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
		WorkspaceDir:  workspaceDir,
	}
}
