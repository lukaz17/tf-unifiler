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

// Package cmdqueue provides an in-memory FIFO command queue which executes
// command argument slices sequentially within the host process.
package cmdqueue

import (
	"time"
)

// Runner validates and executes command argument slices inside the host
// process.
type Runner interface {
	// Find validates that the command route is supported, without executing.
	Find(args []string) error
	// Execute runs the command for real.
	Execute(args []string) error
}

// Status represents the current execution state of a queued job.
type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusDone      Status = "done"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Job is a single queued command execution. Job entries are mutated only by
// the worker goroutine while snapshots are handed out as struct copies.
type Job struct {
	ID          string
	Args        []string
	Status      Status
	SubmittedAt time.Time
	StartedAt   time.Time
	FinishedAt  time.Time
	Error       string
}
