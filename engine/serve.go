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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/server"
)

const serverShutdownTimeout = 17 * time.Second

// ServeModule handles serving the HTTP server.
type ServeModule struct {
	cfg     *config.RootConfig
	httpSrv *server.Server
	logger  zerolog.Logger
}

// Return new ServeModule.
func NewServeModule(c *Controller) *ServeModule {
	return &ServeModule{
		cfg:    c.Config,
		logger: c.CommandLogger("server", "serve"),
	}
}

// Start listen for HTTP requests. Blocks until the server fails or an interrupt/terminate signal triggers a graceful shutdown.
func (m *ServeModule) Start() error {
	addr := fmt.Sprintf(":%d", m.cfg.Server.HTTP.Port)
	m.logger.Info().
		Str("addr", addr).
		Msg("Start listening for HTTP requests.")
	m.httpSrv = server.New(addr)

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
		return m.httpSrv.Shutdown(shutdownCtx)
	}
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
