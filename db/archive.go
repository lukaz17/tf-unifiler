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
	"github.com/google/uuid"
)

type Archive struct {
	ID   uuid.UUID `gorm:"column:id;primaryKey"`
	Name string    `gorm:"column:name;uniqueIndex"`

	SessionID uuid.UUID `gorm:"column:session_id"`
}

func NewArchive(name string) *Archive {
	return &Archive{
		Name: name,
	}
}

func (ctx *DbContext) GetArchiveByName(name string) (*Archive, error) {
	return ctx.findArchiveByName(name)
}

func (ctx *DbContext) GetArchivesByNames(names []string) ([]*Archive, error) {
	return ctx.findArchivesByNames(names)
}

func (ctx *DbContext) SaveArchives(archives []*Archive) error {
	names := make([]string, len(archives))
	for i, archive := range archives {
		names[i] = archive.Name
	}
	changedArchives, err := ctx.findArchivesByNames(names)
	if err != nil {
		return err
	}
	changedArchivesMap := map[string]uuid.UUID{}
	for _, archive := range changedArchives {
		changedArchivesMap[archive.Name] = archive.ID
	}
	newArchives := []*Archive{}
	for _, archive := range archives {
		if _, ok := changedArchivesMap[archive.Name]; ok {
			continue
		}
		newArchives = append(newArchives, archive)
		changedArchivesMap[archive.Name] = archive.ID
	}
	return ctx.writeArchives(newArchives, []*Archive{})
}

func (ctx *DbContext) findArchiveByName(name string) (*Archive, error) {
	var doc *Archive
	result := ctx.db.Model(&Archive{}).
		Where("name = ?", name).
		First(&doc)
	if ctx.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (ctx *DbContext) findArchivesByNames(names []string) ([]*Archive, error) {
	var docs []*Archive
	result := ctx.db.Model(&Archive{}).
		Where("0 = ? OR name IN ?", len(names), names).
		Order("name ASC").
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) writeArchives(newArchives []*Archive, _ []*Archive) error {
	tx := ctx.db.Begin()
	for _, archive := range newArchives {
		if archive.ID == uuid.Nil {
			var err error
			archive.ID, err = uuid.NewRandom()
			if err != nil {
				return err
			}
		}
		result := tx.Create(archive)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}
