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
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/internal/cmdqueue"
	"github.com/tforceaio/tf-unifiler/server"
)

// ServeModule handles serving the HTTP server with internal command queue.
type ServeModule struct {
	cfg     *config.RootConfig
	httpSrv *server.Server
	queue   *cmdqueue.Queue
	logger  zerolog.Logger
}

// Return new ServeModule.
func NewServeModule(c *Controller) *ServeModule {
	return &ServeModule{
		cfg:    c.Config,
		logger: c.CommandLogger("server", "serve"),
	}
}

// Start listen for HTTP requests and consume queued commands. Blocks until
// the server is stopped.
func (m *ServeModule) Start() error {
	addr := fmt.Sprintf(":%d", m.cfg.Server.HTTP.Port)
	m.logger.Info().
		Str("addr", addr).
		Msg("Start listening for HTTP requests.")
	m.queue = cmdqueue.New(queuedRunner{})
	m.httpSrv = server.New(addr, server.WithDispatcher(&queueDispatcher{queue: m.queue}))
	go m.queue.Run()
	defer func() {
		m.queue.Stop()
		m.queue.Wait(10 * time.Second)
		m.httpSrv.Close()
	}()
	return m.httpSrv.Start()
}

// Decorator to log error occurred when calling handlers.
func (m *ServeModule) logError(err error) error {
	logProgramError(m.logger, err)
	return err
}

// Define Cobra Command for Serve module.
func ServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := InitApp()
			defer c.Close()
			m := NewServeModule(c)
			return m.logError(m.Start())
		},
	}
	return serveCmd
}
