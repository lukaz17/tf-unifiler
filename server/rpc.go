// Copyright (C) 2026 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// JobView is the externally visible representation of a queued job.
type JobView struct {
	ID          string    `json:"id"`
	Args        []string  `json:"args"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submittedAt"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  time.Time `json:"finishedAt"`
	Error       string    `json:"error,omitempty"`
}

// Dispatcher defines queue operations which the RPC layer delegates to.
type Dispatcher interface {
	// Enqueue command args, returns the identifier of the created job.
	Enqueue(args []string) (string, error)
	// Return the job with matching ID.
	Job(id string) (JobView, bool)
	// Return all queued and historical jobs.
	Jobs() []JobView
	// Cancel a job which has not started yet.
	Cancel(id string) bool
}

// WithDispatcher registers the command queue RPC endpoints.
func WithDispatcher(d Dispatcher) Option {
	return func(s *Server) {
		s.dispatcher = d
	}
}

// EnqueueJobRequest contains a command line split into argument tokens.
type EnqueueJobRequest struct {
	Args []string `json:"args"`
}

// Enqueue a command args slice for internal execution.
func (s *Server) handleEnqueue(c *gin.Context) {
	req := &EnqueueJobRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(req.Args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "args must contain at least one element"})
		return
	}
	id, err := s.dispatcher.Enqueue(req.Args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"id": id, "status": "queued"})
}

// List all queued and historical jobs.
func (s *Server) handleListJobs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"jobs": s.dispatcher.Jobs()})
}

// Return the job with matching ID.
func (s *Server) handleGetJob(c *gin.Context) {
	view, ok := s.dispatcher.Job(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, view)
}

// Cancel a job which has not started yet.
func (s *Server) handleCancelJob(c *gin.Context) {
	id := c.Param("id")
	if _, ok := s.dispatcher.Job(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	if !s.dispatcher.Cancel(id) {
		c.JSON(http.StatusConflict, gin.H{"error": "job is already running or finished"})
		return
	}
	c.Status(http.StatusNoContent)
}
