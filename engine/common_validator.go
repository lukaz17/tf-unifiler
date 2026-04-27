// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package engine

import (
	"errors"
	"fmt"

	"github.com/tforceaio/tf-unifiler/filesys"
)

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

// Check for non empty inputs.
func validateInputs(inputs []string) error {
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	return nil
}

// Check that exactly 1 input is provided.
func validateSingleInput(inputs []string) error {
	if len(inputs) == 0 {
		return errors.New("input is not set")
	}
	if len(inputs) > 1 {
		return errors.New("only 1 input file is allowed")
	}
	return nil
}

// Check for not empty workspace and its existence on disk.
func validateWorkspace(ws string) error {
	if ws == "" {
		return errors.New("workspace is not set")
	}
	if !filesys.IsDirectoryExist(ws) {
		return errors.New("workspace is not found")
	}
	return nil
}
