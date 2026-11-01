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

	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiarch"
	"github.com/tforce-io/tf-golib/opx"
)

func TestOpen(t *testing.T) {
	ctx := getTestDB("DbContext", "Open")
	if ctx == nil {
		t.Error("cannot open database")
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
