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

package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiarch"
	"github.com/tforce-io/tf-golib/opx"
	"gorm.io/gorm"
)

func TestOpen(t *testing.T) {
	ctx := getTestDB("DbContext", "Open")
	if ctx == nil {
		t.Error("cannot open database")
	}
}

func TestConnectClosesDatabaseWhenMigrationFails(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "unifiler.db")
	// A view named like a table makes the migration fail after the database is opened.
	blocker, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		t.Fatalf("creating database: %v", err)
	}
	if err := blocker.Exec("CREATE VIEW archives AS SELECT 1 AS id").Error; err != nil {
		t.Fatalf("creating view: %v", err)
	}
	blockerDB, err := blocker.DB()
	if err != nil {
		t.Fatalf("getting connection: %v", err)
	}
	if err := blockerDB.Close(); err != nil {
		t.Fatalf("closing database: %v", err)
	}

	_, err = Connect(dbFile, zerolog.Nop())

	if err == nil {
		t.Fatalf("Connect succeeded with a view in place of a table, want an error")
	}
	// Windows refuses to delete a file that is still open.
	if err := os.Remove(dbFile); err != nil {
		t.Errorf("deleting database after failed Connect: %v", err)
	}
}

func getTestDB(entity, function string) *DbContext {
	hasher := sha256.New()
	featSign := fmt.Sprintf("%s/%s/v%d", entity, function, SchemaVersion)
	hasher.Write([]byte(featSign))
	hashBuf := hasher.Sum(nil)
	hash := hex.EncodeToString(hashBuf[:])
	fileName := fmt.Sprintf("unifiler_%s.db", hash)
	tmpDir := opx.Ternary(
		multiarch.IsWindows(),
		os.Getenv("TEMP"),
		"/tmp",
	)
	dbFile := filepath.Join(tmpDir, fileName)
	ctx, err := Connect(dbFile, zerolog.Nop())
	if err != nil {
		panic(err)
	}
	return ctx
}
