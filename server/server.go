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
	"errors"
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
}

// Server is a thin wrapper around http.Server with pre-registered routes.
type Server struct {
	engine *gin.Engine
	http   *http.Server
	tasks  TaskService
}

// Entrypoint for creating a new Server listening on addr, backed by tasks.
func New(addr string, tasks TaskService) *Server {
	s := &Server{tasks: tasks}
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
	r.POST("/tasks", s.handleEnqueue)
	r.GET("/tasks", s.handleList)
	r.GET("/tasks/:id", s.handleGet)
	r.POST("/tasks/:id/cancel", s.handleCancel)
	r.DELETE("/tasks", s.handleClear)
	return r
}

// Enqueue a new task.
func (s *Server) handleEnqueue(c *gin.Context) {
	var req enqueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	t, err := s.tasks.Enqueue(req.Args)
	if err != nil {
		switch {
		case errors.Is(err, task.ErrEmptyArgs):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, task.ErrQueueFull):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		case errors.Is(err, task.ErrClosed):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusAccepted, t)
}

// List all tasks.
func (s *Server) handleList(c *gin.Context) {
	c.JSON(http.StatusOK, s.tasks.List())
}

// Get a single task by id.
func (s *Server) handleGet(c *gin.Context) {
	t, ok := s.tasks.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// Cancel a task by id.
func (s *Server) handleCancel(c *gin.Context) {
	id := c.Param("id")
	if err := s.tasks.Cancel(id); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	t, _ := s.tasks.Get(id)
	c.JSON(http.StatusOK, t)
}

// Clear all finished tasks.
func (s *Server) handleClear(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"cleared": s.tasks.Clear()})
}

type enqueueRequest struct {
	Args []string `json:"args"`
}
