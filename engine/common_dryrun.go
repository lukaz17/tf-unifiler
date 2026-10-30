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

	"github.com/spf13/cobra"
)

var (
	// errStopDryRun aborts a dry run once Cobra has parsed and validated the command, before any command body runs.
	errStopDryRun = errors.New("stop dry run")
	// errNoCommandToRun is returned when Cobra handles args itself instead of running a command.
	errNoCommandToRun = errors.New("arguments run no command (help, version or a command group)")
)

// Run args through Cobra's parsing and validation on root without running any command body, and return the command that would run.
// Return a nil command when Cobra handles args itself (help, version, or a command without a body).
// Mutates root to inject a PersistentPreRunE; do not reuse it in a later call.
func DryRun(root *cobra.Command, args []string) (*cobra.Command, error) {
	var target *cobra.Command
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// Cobra checks these only after PersistentPreRunE, which this hook stops.
		if err := cmd.ValidateRequiredFlags(); err != nil {
			return err
		}
		if err := cmd.ValidateFlagGroups(); err != nil {
			return err
		}
		target = cmd
		return errStopDryRun
	}
	root.SilenceErrors = true
	root.SilenceUsage = true
	// A nil slice would make Cobra read os.Args instead.
	root.SetArgs(append([]string{}, args...))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	err := root.Execute()
	if errors.Is(err, errStopDryRun) {
		return target, nil
	}
	return nil, err
}

// Check that args run a command defined by the engine, by dry-running them on the engine's command tree.
// Cobra's own help and completion commands are rejected: they run no module, and RedactArgs cannot resolve them to mask their args.
func ValidateArgs(args []string) error {
	root := BuildRootCmd()
	defined := commandSet(root) // taken before Execute adds Cobra's own commands to root.
	cmd, err := DryRun(root, args)
	if err != nil {
		return err
	}
	if cmd == nil {
		return errNoCommandToRun
	}
	if !defined[cmd] {
		return fmt.Errorf("command %q is not supported", cmd.CommandPath())
	}
	return nil
}

// Return every command of the tree rooted at root, root included.
func commandSet(root *cobra.Command) map[*cobra.Command]bool {
	set := map[*cobra.Command]bool{}
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		set[c] = true
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
	return set
}
