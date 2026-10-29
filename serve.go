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
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/engine"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/server"
	"github.com/tforceaio/tf-unifiler/task"
)

const serverShutdownTimeout = 17 * time.Second

// ServeModule starts the HTTP server and the task queue that backs it.
type ServeModule struct {
	ctrl    *engine.Controller
	cfg     *config.RootConfig
	httpSrv *server.Server
	queue   *task.Queue
	logger  zerolog.Logger
}

// Return new ServeModule.
func NewServeModule(c *engine.Controller) *ServeModule {
	return &ServeModule{
		ctrl:   c,
		cfg:    c.Config,
		logger: c.CommandLogger("server", "serve"),
	}
}

// Start the task queue and listen for RPC requests. Blocks until the server fails or an interrupt/terminate signal triggers a graceful shutdown.
func (m *ServeModule) Start() error {
	addr := fmt.Sprintf(":%d", m.cfg.Server.HTTP.Port)
	m.logger.Info().
		Str("addr", addr).
		Int("concurrency", m.cfg.Server.Concurrency).
		Msg("Start listening for RPC requests.")

	runner := func(ctx context.Context, args []string, out io.Writer) error {
		return engine.RunTask(m.ctrl, args, ctx, out)
	}
	m.queue = task.New(runner, m.cfg.Server.Concurrency, m.cfg.Server.Backlog, m.ctrl.Logger)
	m.queue.Start()

	m.httpSrv = server.New(addr, m.statusInfo(), m.queue)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- m.httpSrv.Start()
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		stop()
		m.logger.Info().Msg("Shutdown signal received, stopping HTTP server.")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()
		queueErr := m.queue.Shutdown(shutdownCtx)
		if err := m.httpSrv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return queueErr
	}
}

// Gather the static identity reported by the server's status endpoint.
func (m *ServeModule) statusInfo() server.StatusInfo {
	pwd, _ := os.Getwd()
	pwd, _ = filesys.GetAbsPath(pwd)
	exec, _ := os.Executable()
	exec, _ = filesys.GetAbsPath(exec)
	return server.StatusInfo{
		Version:    engine.Version(),
		WorkingDir: pwd,
		ConfigDir:  m.cfg.ConfigDir,
		Executable: exec,
		Portable:   m.cfg.IsPortable,
	}
}

// Decorator to log error occurred when calling handlers.
func (m *ServeModule) logError(err error) error {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
	return err
}

// Define Cobra Command for Serve module.
func ServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := engine.InitAppWithBanner(context.TODO())
			defer c.Close()
			m := NewServeModule(c)
			return m.logError(m.Start())
		},
	}
	return serveCmd
}
