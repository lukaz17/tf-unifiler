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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiarch"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/db"
)

const (
	// fakeArchiverEnv makes the test binary act as an archiver that only prints fakeArchiverOutput, so pack runs without 7z.
	fakeArchiverEnv    = "UNIFILER_FAKE_ARCHIVER"
	fakeArchiverOutput = "fake archiver output"
	// fakeVideoToolsEnv makes the test binary act as mediainfo, printing fakeMediaInfoReport of a 10 s video,
	// and as ffmpeg, writing its -ss value into its output file, so extract-frames runs without them.
	fakeVideoToolsEnv   = "UNIFILER_FAKE_VIDEO_TOOLS"
	fakeMediaInfoReport = `{"media":{"track":[{"@type":"General","Duration":"10.000"},{"@type":"Video"}]}}`
)

func TestMain(m *testing.M) {
	if os.Getenv(fakeArchiverEnv) == "1" {
		fmt.Println(fakeArchiverOutput)
		os.Exit(0)
	}
	if os.Getenv(fakeVideoToolsEnv) == "1" {
		os.Exit(runFakeVideoTool(os.Args[1:]))
	}
	os.Exit(m.Run())
}

func getTestDB(entity, function string) *db.DbContext {
	hasher := sha256.New()
	featSign := fmt.Sprintf("%s/%s/v%d", entity, function, db.SchemaVersion)
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
	ctx, err := db.Connect(dbFile, zerolog.Nop())
	if err != nil {
		panic(err)
	}
	return ctx
}

// Act as ffmpeg when args start with -ss, as extract-frames always seeks, otherwise as mediainfo.
func runFakeVideoTool(args []string) int {
	if len(args) < 2 || args[0] != "-ss" {
		fmt.Print(fakeMediaInfoReport)
		return 0
	}
	// ffmpeg's output file is its last argument.
	if err := os.WriteFile(args[len(args)-1], []byte(args[1]), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
