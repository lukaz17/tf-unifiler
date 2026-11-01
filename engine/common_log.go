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
	"io"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// commandError is the error inside a command body, as opposed to Cobra rejecting the arguments before any body ran.
type commandError struct {
	err    error
	logged bool
}

// Return the inner error's message.
func (e *commandError) Error() string {
	return e.err.Error()
}

// Return the inner error.
func (e *commandError) Unwrap() error {
	return e.err
}

// Log the err using given logger instance.
// Then wrap it into a commandError for further handling with PrintError.
func LogCommandError(logger zerolog.Logger, err error) error {
	if err == nil {
		return nil
	}
	e := logger.Err(err) // nil when the log level hides errors.
	logged := e.Enabled()
	e.Msg("Unexpected error has occurred.")
	return &commandError{err: err, logged: logged}
}

// Print err returned by executing cmd to w, unless the failing command already logged it.
// Otherwise print it similar to Cobra's format.
func PrintError(w io.Writer, cmd *cobra.Command, err error) {
	var ce *commandError
	switch {
	case err == nil:
	case errors.As(err, &ce):
		if !ce.logged {
			fmt.Fprintln(w, cmd.ErrPrefix(), err)
		}
	default:
		fmt.Fprintln(w, cmd.ErrPrefix(), err)
		if cmd.CalledAs() == "" {
			// Cobra rejected an unknown command before running it, and prints a hint instead of usage.
			fmt.Fprintf(w, "Run '%s --help' for usage.\n", cmd.CommandPath())
			return
		}
		fmt.Fprintln(w, cmd.UsageString())
	}
}
