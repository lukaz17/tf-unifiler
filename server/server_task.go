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
	"errors"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tforceaio/tf-unifiler/task"
)

// Enqueue a new task.
func (s *Server) handleEnqueue(c *gin.Context) {
	// A browser sends a cross-origin JSON body only after a CORS preflight, which this server never grants; a form can send the same body as text/plain without one.
	if mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type")); err != nil || mediaType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "request body must be application/json"})
		return
	}

	var req enqueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	t, err := s.tasks.Enqueue(req.Args)
	if err != nil {
		switch {
		case errors.Is(err, task.ErrEmptyArgs), errors.Is(err, task.ErrInvalidArgs):
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
