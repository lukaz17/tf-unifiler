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
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/engine"
	"github.com/tforceaio/tf-unifiler/server"
)

// queueFlagName is the name of the global bool flag that determine execution is deferred or immediate.
const queueFlagName = "queue"

// Check args for the --queue flag to determine whether execution is deferred or immediate.
// Mutates the root command through engine.DryRun; do not reuse it in a later call.
func isDeferred(root *cobra.Command, args []string) (bool, error) {
	cmd, err := engine.DryRun(root, args)
	if err != nil || cmd == nil {
		return false, err
	}
	queued, _ := cmd.Flags().GetBool(queueFlagName)
	return queued, nil
}

// Forward args to the running TFunifiler server to be enqueued.
// It discovers the address from config, confirms the listener is TFunifiler, then enqueues the accepted task.
func runQueued(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("nothing to enqueue")
	}

	cfg, err := config.InitKoanf(true)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}
	host := server.DialHost(cfg.Server.HTTP.Hostname)
	addr := net.JoinHostPort(host, strconv.Itoa(cfg.Server.HTTP.Port))

	ctx := context.Background()
	client := server.NewClient("http://"+addr, cfg.Server.APIKey)

	st, err := client.Status(ctx)
	if errors.Is(err, server.ErrUnauthorized) {
		return fmt.Errorf("not forwarding to the server on %s: %w", addr, err)
	}
	if err != nil {
		// A server on another machine cannot be started from here.
		if !cfg.Server.Autostart || !server.IsLoopback(host) {
			return fmt.Errorf("no TFunifiler server reachable on %s; is 'unifiler server' running? (%w)", addr, err)
		}
		lock := server.NewLock(cfg.ConfigDir)
		probe := func(ctx context.Context) error {
			_, perr := client.Status(ctx)
			return perr
		}
		if serr := ensureServer(ctx, lock, probe, spawnServer, serverStartupTimeout, serverPollInterval); serr != nil {
			return fmt.Errorf("auto-starting TFunifiler server on %s: %w", addr, serr)
		}
		if st, err = client.Status(ctx); err != nil {
			return fmt.Errorf("no TFunifiler server reachable on %s after auto-start (%w)", addr, err)
		}
	}
	if err := verifyServer(st, cfg.Server.Name, engine.Version()); err != nil {
		return fmt.Errorf("not forwarding to the server on %s: %w", addr, err)
	}

	t, err := client.Enqueue(ctx, args)
	if err != nil {
		return err
	}

	fmt.Printf("Enqueued task %s (%s)\n", t.ID, t.Status)
	return nil
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

// Validate status document that belong to a matching TFunifiler server.
func verifyServer(st server.StatusResponse, wantService, wantVersion string) error {
	if err := verifyService(st, wantService); err != nil {
		return err
	}
	if st.Version != wantVersion {
		return fmt.Errorf("server version %q does not match client %q; restart 'unifiler server' with this build", st.Version, wantVersion)
	}
	return nil
}

// Validate status document that belong to a TFunifiler server of any version.
func verifyService(st server.StatusResponse, wantService string) error {
	if st.Service != wantService {
		return fmt.Errorf("service %q does not match expected %q; it is not TFunifiler", st.Service, wantService)
	}
	return nil
}
