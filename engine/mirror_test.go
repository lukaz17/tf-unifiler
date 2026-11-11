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

package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tforceaio/tf-unifiler/config"
)

func TestMirrorExportClosesInvalidChecksumFile(t *testing.T) {
	workspace := t.TempDir()
	checksumFile := filepath.Join(t.TempDir(), "photos.sha256")
	if err := os.WriteFile(checksumFile, []byte("zz *photo.jpg\n"), 0644); err != nil {
		t.Fatalf("creating checksum file: %v", err)
	}
	base := &Controller{Config: &config.RootConfig{}}
	args := []string{"mirror", "export", "-w", workspace, "-i", checksumFile}

	var out bytes.Buffer
	err := RunTask(base, args, context.Background(), &out)

	if err == nil {
		t.Fatalf("RunTask(%v) succeeded with an invalid hash, want an error", args)
	}
	// Windows refuses to delete a file that is still open.
	if err := os.Remove(checksumFile); err != nil {
		t.Errorf("deleting checksum file after export: %v", err)
	}
}

func TestMirrorExportFailsOnUnresolvableOutput(t *testing.T) {
	// An export that ignores the failure links into the current folder.
	t.Chdir(t.TempDir())
	workspace := t.TempDir()
	content := []byte("photo")
	hash := sha256.Sum256(content)
	hashHex := hex.EncodeToString(hash[:])
	if err := os.MkdirAll(MirrorWorkspaceRoot(workspace), 0755); err != nil {
		t.Fatalf("creating store: %v", err)
	}
	if err := os.WriteFile(filepath.Join(MirrorWorkspaceRoot(workspace), hashHex), content, 0644); err != nil {
		t.Fatalf("creating stored file: %v", err)
	}
	checksumFile := filepath.Join(t.TempDir(), "photos.sha256")
	if err := os.WriteFile(checksumFile, fmt.Appendf(nil, "%s *album/photo.jpg\n", hashHex), 0644); err != nil {
		t.Fatalf("creating checksum file: %v", err)
	}
	base := &Controller{Config: &config.RootConfig{}}
	// Windows cannot resolve a path containing NUL.
	args := []string{"mirror", "export", "-w", workspace, "-i", checksumFile, "-o", "out\x00dir"}

	var out bytes.Buffer
	err := RunTask(base, args, context.Background(), &out)

	if err == nil {
		t.Errorf("RunTask(%q) succeeded with an output folder that cannot be resolved, want an error", args)
	}
}

func TestParseMirrorFlagsKeepsCommaInInput(t *testing.T) {
	cmd, _, err := MirrorCmd().Find([]string{"scan"})
	if err != nil {
		t.Fatalf("finding scan: %v", err)
	}
	if err := cmd.ParseFlags([]string{"-i", "a,b.txt"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}
	flags := ParseMirrorFlags(cmd, nil)
	if len(flags.Inputs) != 1 || flags.Inputs[0] != "a,b.txt" {
		t.Errorf("inputs = %q, want [\"a,b.txt\"]", flags.Inputs)
	}
}
