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
	"github.com/tforceaio/tf-unifiler/task"
)

// TaskService is the subset of the task queue the HTTP layer depends on.
type TaskService interface {
	Enqueue(args []string) (task.Task, error)
	Get(id string) (task.Task, bool)
	List() []task.Task
	Cancel(id string) error
	Clear() int
	Stats() (pending, total int)
}

// Server is a thin wrapper around http.Server with pre-registered routes.
type Server struct {
	engine *gin.Engine
	http   *http.Server
	info   StatusInfo
	tasks  TaskService
}

// Entrypoint for creating a new Server listening on addr, backed by tasks.
func New(addr, apiKey string, info StatusInfo, tasks TaskService) *Server {
	s := &Server{info: info, tasks: tasks}
	s.engine = s.router(apiKey)
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
func (s *Server) router(apiKey string) *gin.Engine {
	// gin.New() instead of Default: no built-in logger, project logs via zerolog.
	r := gin.New()
	r.Use(gin.Recovery())
	if apiKey != "" {
		r.Use(requireAPIKey(apiKey))
	}
	r.GET("/status", s.handleStatus)
	r.POST("/tasks", s.handleEnqueue)
	r.GET("/tasks", s.handleList)
	r.GET("/tasks/:id", s.handleGet)
	r.POST("/tasks/:id/cancel", s.handleCancel)
	r.DELETE("/tasks", s.handleClear)
	return r
}

// Report the server's identity and live task counts.
func (s *Server) handleStatus(c *gin.Context) {
	pending, total := s.tasks.Stats()
	c.JSON(http.StatusOK, StatusResponse{
		Service:      s.info.Service,
		Version:      s.info.Version,
		WorkingDir:   s.info.WorkingDir,
		ConfigDir:    s.info.ConfigDir,
		Executable:   s.info.Executable,
		Portable:     s.info.Portable,
		PendingTasks: pending,
		TotalTasks:   total,
	})
}

// StatusInfo is the static server identity reported by the status endpoint.
type StatusInfo struct {
	Service    string
	Version    string
	WorkingDir string
	ConfigDir  string
	Executable string
	Portable   bool
}

// StatusResponse is the body returned by the status endpoint, shared by the client.
type StatusResponse struct {
	Service      string `json:"service"`
	Version      string `json:"version"`
	WorkingDir   string `json:"workingDir"`
	ConfigDir    string `json:"configDir"`
	Executable   string `json:"executable"`
	Portable     bool   `json:"portable"`
	PendingTasks int    `json:"pendingTasks"`
	TotalTasks   int    `json:"totalTasks"`
}
