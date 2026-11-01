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

const queueableAnnotation = "unifiler_queueable" // command annotation allowing ValidateArgs to accept the command and its subcommands as tasks.

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

// Mark cmd and its subcommands as queueable so ValidateArgs accepts them as tasks.
func MarkCommandQueueable(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[queueableAnnotation] = "true"
}

// Check that args run a queueable command, by dry-running them on the engine's command tree.
// A command is rejected unless it or an ancestor is marked by MarkCommandQueueable, so a new command stays local until it opts in.
// Cobra's own help and completion commands are never marked: they run no module, and RedactArgs cannot resolve them to mask their args.
func ValidateArgs(args []string) error {
	cmd, err := DryRun(BuildRootCmd(), args)
	if err != nil {
		return err
	}
	if cmd == nil {
		return errNoCommandToRun
	}
	if !isQueueable(cmd) {
		return fmt.Errorf("command %q cannot run on the server", cmd.CommandPath())
	}
	return nil
}

// Report whether cmd or one of its ancestors is marked by MarkCommandQueueable.
func isQueueable(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if _, ok := c.Annotations[queueableAnnotation]; ok {
			return true
		}
	}
	return false
}
