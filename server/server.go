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

package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server is a thin wrapper around http.Server with pre-registered routes.
type Server struct {
	engine *gin.Engine
	http   *http.Server
}

// Entrypoint for creating a new Server listening on addr.
func New(addr string) *Server {
	s := &Server{}
	s.engine = s.router()
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.engine,

		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s
}

// Return the http.Handler built for this Server.
func (s *Server) Handler() http.Handler {
	return s.engine
}

// Start listening. Blocks until the server is stopped or fails.
func (s *Server) Start() error {
	err := s.http.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Gracefully stop serving and wait for in-flight requests to finish, bounded by ctx.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// Build the route table served by this Server.
func (s *Server) router() *gin.Engine {
	// gin.New() instead of Default: no built-in logger, project logs via zerolog.
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}
