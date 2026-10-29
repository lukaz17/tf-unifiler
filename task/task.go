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

package task

import (
	"context"
	"io"
	"sync"
	"time"
)

// Status is the lifecycle state of a task.
type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Runner executes a task's argument vector, writing human-readable output to out.
// It must honor ctx where it can, though cancellation of an already-running task is best-effort.
type Runner func(ctx context.Context, args []string, out io.Writer) error

// Task is an immutable snapshot of a task's state, safe to hand to callers.
type Task struct {
	ID         string     `json:"id"`
	Args       []string   `json:"args"`
	Status     Status     `json:"status"`
	Err        string     `json:"error,omitempty"`
	Output     string     `json:"output,omitempty"`
	EnqueuedAt time.Time  `json:"enqueuedAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}

// entry is the live, mutable state of a task. Its fields are guarded by mu; the
// live pointer never escapes the queue (callers only ever receive snapshots).
type entry struct {
	mu         sync.Mutex
	id         string
	args       []string
	status     Status
	err        string
	out        *boundedBuffer
	cancel     context.CancelFunc
	enqueuedAt time.Time
	startedAt  time.Time
	finishedAt time.Time
}

// Return an immutable copy of the entry's current state.
func (e *entry) snapshot() Task {
	e.mu.Lock()
	defer e.mu.Unlock()

	t := Task{
		ID:         e.id,
		Args:       append([]string(nil), e.args...),
		Status:     e.status,
		Err:        e.err,
		Output:     e.out.String(),
		EnqueuedAt: e.enqueuedAt,
	}
	if !e.startedAt.IsZero() {
		started := e.startedAt
		t.StartedAt = &started
	}
	if !e.finishedAt.IsZero() {
		finished := e.finishedAt
		t.FinishedAt = &finished
	}
	return t
}

func (e *entry) isTerminal() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status == StatusSucceeded || e.status == StatusFailed || e.status == StatusCanceled
}

// boundedBuffer is a concurrency-safe byte sink capped at limit bytes.
// Writes past the cap are dropped (but reported as written so the logger never errors) and marked as truncated. A zero limit means unbounded.
type boundedBuffer struct {
	mu        sync.Mutex
	buf       []byte
	limit     int
	truncated bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.limit <= 0 {
		b.buf = append(b.buf, p...)
		return len(p), nil
	}
	avail := b.limit - len(b.buf)
	if avail <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > avail {
		b.buf = append(b.buf, p[:avail]...)
		b.truncated = true
	} else {
		b.buf = append(b.buf, p...)
	}
	return len(p), nil
}

func (b *boundedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.truncated {
		return string(b.buf) + "\n... output truncated ...\n"
	}
	return string(b.buf)
}
