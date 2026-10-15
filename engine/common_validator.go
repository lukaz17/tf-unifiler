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
	"errors"
	"fmt"
	"slices"

	"github.com/tforceaio/tf-unifiler/db"
	"github.com/tforceaio/tf-unifiler/filesys"
)

// Check for valid archive format.
func validateArchiveFormat(format string) error {
	if format == "" {
		return errors.New("format is required")
	}
	formats := []string{"7z", "rar"}
	if slices.Contains(formats, format) {
		return nil
	}
	return fmt.Errorf("unknown format: %s", format)
}

// Check a anchive name is available for create or update.
func validateArchiveName(ctx *db.DbContext, archiveName string, update bool) error {
	existingArchive, err := ctx.GetArchiveByName(archiveName)
	if err != nil {
		return fmt.Errorf("failed to get archive: %w", err)
	}
	if existingArchive != nil && !update {
		return fmt.Errorf("archive %q already exists, use --update to add contents to it", archiveName)
	}
	return nil
}

// Check for non empty input and its existence on disk.
func validateInput(input, label string) error {
	if input == "" {
		return fmt.Errorf("%s is not set", label)
	}
	if !filesys.IsFileExist(input) {
		return fmt.Errorf("%s is not found", label)
	}
	return nil
}

// Check that at least one input is provided.
func validateInputs(inputs []string) error {
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	return nil
}

// Check that exactly one input is provided.
func validateSingleInput(inputs []string) error {
	if len(inputs) == 0 {
		return errors.New("input is not set")
	}
	if len(inputs) > 1 {
		return errors.New("only 1 input file is allowed")
	}
	return nil
}

// Check that a required string parameter is not empty.
func validateRequiredString(input, label string) error {
	if input == "" {
		return fmt.Errorf("%s is not set", label)
	}
	return nil
}

// Check that workspace is set and exists as a directory on disk.
func validateWorkspace(ws string) error {
	if ws == "" {
		return errors.New("workspace is not set")
	}
	if !filesys.IsDirectoryExist(ws) {
		return errors.New("workspace is not found")
	}
	return nil
}
