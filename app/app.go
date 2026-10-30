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

package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/engine"
)

// Main is the program entry point.
func Main() {
	isDeferred, err := isDeferred(buildRootCmd(), os.Args[1:])
	if err == nil && isDeferred {
		if ferr := runQueued(stripQueueFlag(os.Args[1:])); ferr != nil {
			fmt.Fprintln(os.Stderr, ferr)
			os.Exit(1)
		}
		return
	}

	rootCmd := buildRootCmd() // rootCmd cannot share the instance with isDeferred() because isDeferred() mutate it during execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Assemble the program's root command.
func buildRootCmd() *cobra.Command {
	rootCmd := engine.BuildRootCmd()
	rootCmd.AddCommand(ServerCmd())
	rootCmd.PersistentFlags().Bool(queueFlagName, false, "Queue this command to the running TFunifiler server instead of executing immediately.")
	return rootCmd
}
