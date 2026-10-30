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
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/tforceaio/tf-unifiler/server"
)

const (
	// serverPollInterval is how long the client waits between probes while an auto-started server comes up.
	serverPollInterval = 250 * time.Millisecond

	// serverStartupTimeout bounds how long the client waits for an auto-started server to become reachable.
	serverStartupTimeout = 15 * time.Second
)

// serverProbe reports whether the server is reachable; a nil error means reachable.
type serverProbe func(ctx context.Context) error

// serverSpawn starts a detached server process.
type serverSpawn func() error

// Make a server reachable. It waits when one is already starting otherwise spawns one, then polls until the server answers or the timeout elapses.
func ensureServer(ctx context.Context, lock *server.Lock, probe serverProbe, spawn serverSpawn, timeout, interval time.Duration) error {
	if _, alive := lock.Owner(); !alive {
		if err := spawn(); err != nil {
			return fmt.Errorf("starting server: %w", err)
		}
	}
	return pollUntilReachable(ctx, probe, timeout, interval)
}

// Probe until the server is reachable, or the deadline passes, or ctx is done.
func pollUntilReachable(ctx context.Context, probe serverProbe, timeout, interval time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		if err := probe(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("server did not become reachable within %s", timeout)
		case <-time.After(interval):
		}
	}
}

// Launch "unifiler server start" as a detached background process.
func spawnServer() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating executable: %w", err)
	}
	cmd := exec.Command(exe, "server", "start")
	cmd.SysProcAttr = detachSysProcAttr()

	devnull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("opening null device: %w", err)
	}
	defer devnull.Close()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = devnull, devnull, devnull

	return cmd.Start() // detached; the child keeps running after this process exits
}
