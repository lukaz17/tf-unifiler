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

package db

import (
	"crypto/sha256"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/tforce-io/tf-golib/opx/slicext"
)

type ArchiveContent struct {
	ArchiveID uuid.UUID `gorm:"column:archive_id;primaryKey"`
	PathHash  Bytes32   `gorm:"column:path_hash;primaryKey"`
	Directory string    `gorm:"column:directory"`
	Name      string    `gorm:"column:name"`
	Extension string    `gorm:"column:extension"`
	HashID    uuid.UUID `gorm:"column:hash_id"`

	SessionID uuid.UUID `gorm:"column:session_id"`
}

func NewArchiveContent(archiveID uuid.UUID, directory, name, extension string, hashID uuid.UUID) *ArchiveContent {
	return &ArchiveContent{
		ArchiveID: archiveID,
		PathHash:  archiveContentPathHash(directory, name, extension),
		Directory: directory,
		Name:      name,
		Extension: extension,
		HashID:    hashID,
	}
}

func (ctx *DbContext) GetArchiveContentsByArchiveIDs(archiveIDs uuid.UUIDs) ([]*ArchiveContent, error) {
	return ctx.findArchiveContentsByArchiveIDs(archiveIDs)
}

func (ctx *DbContext) GetArchiveContentsByHashIDs(hashIDs uuid.UUIDs) ([]*ArchiveContent, error) {
	return ctx.findArchiveContentsByHashIDs(hashIDs)
}

func (ctx *DbContext) SaveArchiveContents(archiveContents []*ArchiveContent) error {
	archiveIDs := make([]uuid.UUID, len(archiveContents))
	for i, ac := range archiveContents {
		archiveIDs[i] = ac.ArchiveID
	}
	changedArchiveContents, err := ctx.findArchiveContentsByArchiveIDs(archiveIDs)
	if err != nil {
		return err
	}
	newArchiveContents := []*ArchiveContent{}
	for _, ac := range archiveContents {
		existed := slicext.ContainsFunc(changedArchiveContents, ac, areEqualArchiveContents)
		if existed {
			continue
		}
		newArchiveContents = append(newArchiveContents, ac)
		changedArchiveContents = append(changedArchiveContents, ac)
	}
	return ctx.writeArchiveContents(newArchiveContents, []*ArchiveContent{})
}

func (ctx *DbContext) findArchiveContentsByArchiveIDs(archiveIDs uuid.UUIDs) ([]*ArchiveContent, error) {
	var docs []*ArchiveContent
	result := ctx.db.Model(&ArchiveContent{}).
		Where("archive_id IN ?", archiveIDs).
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) findArchiveContentsByHashIDs(hashIDs uuid.UUIDs) ([]*ArchiveContent, error) {
	var docs []*ArchiveContent
	result := ctx.db.Model(&ArchiveContent{}).
		Where("hash_id IN ?", hashIDs).
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) writeArchiveContents(newArchiveContents []*ArchiveContent, _ []*ArchiveContent) error {
	tx := ctx.db.Begin()
	for _, ac := range newArchiveContents {
		result := tx.Create(ac)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}

func areEqualArchiveContents(x, y *ArchiveContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return x.ArchiveID == y.ArchiveID && x.PathHash == y.PathHash
}

func archiveContentPathHash(directory, name, extension string) Bytes32 {
	path := filepath.Join(directory, name+extension)
	return Bytes32(sha256.Sum256([]byte(path)))
}
