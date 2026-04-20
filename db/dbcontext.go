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
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/tforceaio/tf-unifiler-go/filesys"
	"gorm.io/gorm"
)

var SchemaVersion = 2

// DbContext encapsulate all actions related to reading from and writing to database.
type DbContext struct {
	db  *gorm.DB
	uri string
}

// Return new DbContext if the connection is successful.
// Target database will be migrated to match database models.
func Connect(uri string) (*DbContext, error) {
	parentDir := filepath.Dir(uri)
	if !filesys.IsDirectoryExist(parentDir) {
		err := filesys.CreateDirectoryRecursive(parentDir)
		if err != nil {
			return nil, err
		}
	}
	db, err := gorm.Open(sqlite.Open(uri), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	c := &DbContext{db, uri}
	c.Migrate()
	return c, nil
}

// Disconnect from database. Currently used for Gorm.
func (c *DbContext) Disconnect() {
}

// Migrate database schema to match database models.
func (c *DbContext) Migrate() error {
	return c.db.AutoMigrate(
		&Hash{},
		&Mapping{},
		&Session{},
		&Set{},
		&SetHash{},
	)
}

// Count number of records in a single table that satisfy provided condition.
func (c *DbContext) Count(model interface{}, query, args interface{}) (int64, error) {
	var count int64
	if query == nil {
		result := c.db.Model(model).Count(&count)
		return count, result.Error
	}
	result := c.db.Model(model).
		Where(query, args).
		Count(&count)
	return count, result.Error
}

// Truncate all tables.
func (c *DbContext) Reset() {
	c.Truncate(&Hash{})
	c.Truncate(&Mapping{})
	c.Truncate(&Session{})
	c.Truncate(&Set{})
	c.Truncate(&SetHash{})
}

// Truncate specified table.
func (c *DbContext) Truncate(model interface{}) {
	c.db.Where("1 = 1").Delete(model)
}

// Determine if the error is record not found. Specifically for Gorm.
func (c *DbContext) isEmptyResultError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "record not found"
}
