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

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/server"
)

// queueFlagName is the name of the global --queue bool flag that forwards a command to the running server instead of executing it immediately.
const queueFlagName = "queue"

// errStopQueuePlanning aborts the isDeferred dry run once Cobra has parsed and validated the command, before any command body runs.
var errStopQueuePlanning = errors.New("stop queue planning")

// Check args for the --queue flag to decide whether execution is deferred or immediate.
// Mutates the root command to inject a PersistentPreRunE that dry-runs Cobra; do not reuse it in a later call.
func isDeferred(root *cobra.Command, args []string) (queued bool, err error) {
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		queued, _ = cmd.Flags().GetBool(queueFlagName)
		return errStopQueuePlanning
	}
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs(args)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	err = root.Execute()
	if errors.Is(err, errStopQueuePlanning) {
		return queued, nil
	}
	return false, err
}

// Return a new args slice with all instances of the --queue flag removed.
func stripQueueFlag(args []string) []string {
	queueFlag := "--" + queueFlagName
	rest := make([]string, 0, len(args))
	for i, a := range args {
		switch {
		case a == "--":
			return append(rest, args[i:]...)
		case a == queueFlag || strings.HasPrefix(a, queueFlag+"="):
			// Client-only flag: drop it.
		default:
			rest = append(rest, a)
		}
	}
	return rest
}

// Forward args to the running TFunifiler server to be enqueued.
// It discovers the port from config, confirms the listener is TFunifiler, then enqueues the accepted task.
func runQueued(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("nothing to enqueue")
	}

	cfg, err := config.InitKoanf(true)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}
	port := cfg.Server.HTTP.Port

	ctx := context.Background()
	client := server.NewClient(fmt.Sprintf("http://127.0.0.1:%d", port))

	st, err := client.Status(ctx)
	if err != nil {
		return fmt.Errorf("no TFunifiler server reachable on 127.0.0.1:%d; is 'unifiler serve' running? (%w)", port, err)
	}
	if st.Service != server.ServiceName {
		return fmt.Errorf("a server is listening on 127.0.0.1:%d but it is not TFunifiler; not forwarding", port)
	}

	t, err := client.Enqueue(ctx, args)
	if err != nil {
		return err
	}

	fmt.Printf("Enqueued task %s (%s)\n", t.ID, t.Status)
	return nil
}
