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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler/config"
)

func TestFilePackWritesSummaryToTaskOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	output := filepath.Join(dir, "output.7z")
	// The archiver does not exist, so pack fails right after printing its summary and never touches the inputs.
	base := &Controller{Config: &config.RootConfig{Path: &config.PathConfig{X7zPath: filepath.Join(dir, "missing-7z")}}}
	args := []string{"file", "pack", input, "-o", output}

	var out bytes.Buffer
	_ = RunTask(base, args, context.Background(), &out)

	for _, want := range []string{"\nInputs:\n - " + input + "\n", "\nOutput:\n - " + output + "\n"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("task output does not contain %q, so the summary was printed elsewhere; got:\n%s", want, out.String())
		}
	}
}

func TestFilePackWritesArchiverOutputToTaskOutput(t *testing.T) {
	t.Setenv(fakeArchiverEnv, "1")
	dir := t.TempDir()
	base := &Controller{Config: &config.RootConfig{Path: &config.PathConfig{X7zPath: os.Args[0]}}}
	args := []string{"file", "pack", filepath.Join(dir, "input.txt"), "-o", filepath.Join(dir, "output.7z")}

	var out bytes.Buffer
	err := RunTask(base, args, context.Background(), &out)

	if err != nil {
		t.Fatalf("RunTask(%v): %v", args, err)
	}
	if !strings.Contains(out.String(), fakeArchiverOutput) {
		t.Errorf("task output does not contain %q, so the archiver wrote elsewhere; got:\n%s", fakeArchiverOutput, out.String())
	}
}
