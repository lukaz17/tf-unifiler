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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const defaultOutputLimit = 64 * 1024 // bytes of captured output retained per task.

var (
	// ErrClosed is returned when enqueueing after the queue has shut down.
	ErrClosed = errors.New("task queue is closed")
	// ErrEmptyArgs is returned when enqueueing a task with no arguments.
	ErrEmptyArgs = errors.New("task arguments cannot be empty")
	// ErrInvalidArgs is returned when enqueueing a task whose arguments the queue's Validator rejects.
	ErrInvalidArgs = errors.New("invalid task arguments")
	// ErrNotFound is returned when no task matches the given id.
	ErrNotFound = errors.New("task not found")
	// ErrQueueFull is returned when the pending backlog is at capacity.
	ErrQueueFull = errors.New("task queue is full")
)

// Queue is an in-memory FIFO task queue processed by a fixed pool of workers.
type Queue struct {
	runner    Runner
	validator Validator
	redactor  Redactor
	workers   int
	outLimit  int
	logger    zerolog.Logger

	mu        sync.Mutex
	entries   map[string]*entry
	order     []string
	accepting bool
	pending   chan string

	wg sync.WaitGroup
}

// Return new queue that runs tasks via runner, rejects their args via validator and masks them via redactor.
func New(runner Runner, validator Validator, redactor Redactor, concurrency, backlog int, logger zerolog.Logger) *Queue {
	if concurrency < 1 {
		concurrency = 1
	}
	if backlog < 1 {
		backlog = 1
	}
	return &Queue{
		runner:    runner,
		validator: validator,
		redactor:  redactor,
		workers:   concurrency,
		outLimit:  defaultOutputLimit,
		logger:    logger,
		entries:   make(map[string]*entry),
		pending:   make(chan string, backlog),
		accepting: true,
	}
}

// Start the worker pool. Call once.
func (q *Queue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

// Stop accepting new tasks, cancel pending and running tasks, and wait for the workers to drain, bounded by ctx.
func (q *Queue) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	if q.accepting {
		q.accepting = false
		close(q.pending)
	}
	for _, e := range q.entries {
		e.mu.Lock()
		switch e.status {
		case StatusQueued:
			e.status = StatusCanceled
			e.finishedAt = time.Now().UTC()
		case StatusRunning:
			if e.cancel != nil {
				e.cancel()
			}
		}
		e.mu.Unlock()
	}
	q.mu.Unlock()

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Add a task to the back of the queue and return its snapshot.
func (q *Queue) Enqueue(args []string) (Task, error) {
	if len(args) == 0 {
		return Task{}, ErrEmptyArgs
	}
	if err := q.validateArgs(args); err != nil {
		return Task{}, fmt.Errorf("%w: %w", ErrInvalidArgs, err)
	}
	redacted := q.redactArgs(args)

	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.accepting {
		return Task{}, ErrClosed
	}

	e := &entry{
		id:         uuid.NewString(),
		args:       append([]string(nil), args...),
		redacted:   redacted,
		status:     StatusQueued,
		out:        &boundedBuffer{limit: q.outLimit},
		enqueuedAt: time.Now().UTC(),
	}

	// Non-blocking send so a full backlog is reported rather than blocking the caller.
	// The accepting check and the send share q.mu, so no send can race with Shutdown closing the channel.
	select {
	case q.pending <- e.id:
		q.entries[e.id] = e
		q.order = append(q.order, e.id)
		q.logger.Debug().Str("id", e.id).Strs("args", e.redacted).Msg("Task enqueued.")
		return e.snapshot(), nil
	default:
		return Task{}, ErrQueueFull
	}
}

// Return a snapshot of the task with the given id.
func (q *Queue) Get(id string) (Task, bool) {
	e := q.lookup(id)
	if e == nil {
		return Task{}, false
	}
	return e.snapshot(), true
}

// Return snapshots of all tasks in enqueue order.
func (q *Queue) List() []Task {
	q.mu.Lock()
	entries := make([]*entry, 0, len(q.order))
	for _, id := range q.order {
		if e := q.entries[id]; e != nil {
			entries = append(entries, e)
		}
	}
	q.mu.Unlock()

	tasks := make([]Task, 0, len(entries))
	for _, e := range entries {
		tasks = append(tasks, e.snapshot())
	}
	return tasks
}

// Mark a queued task canceled or signal a running task to stop (best-effort).
func (q *Queue) Cancel(id string) error {
	e := q.lookup(id)
	if e == nil {
		return ErrNotFound
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	switch e.status {
	case StatusQueued:
		e.status = StatusCanceled
		e.finishedAt = time.Now().UTC()
	case StatusRunning:
		e.status = StatusCanceled
		if e.cancel != nil {
			e.cancel()
		}
	}
	return nil
}

// Remove all terminal (succeeded/failed/canceled) tasks and return the number removed.
func (q *Queue) Clear() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	kept := make([]string, 0, len(q.order))
	removed := 0
	for _, id := range q.order {
		e := q.entries[id]
		if e == nil {
			continue
		}
		if e.isTerminal() {
			delete(q.entries, id)
			removed++
			continue
		}
		kept = append(kept, id)
	}
	q.order = kept
	return removed
}

// Return the number of tasks not yet terminal (queued or running) and the total number of tasks currently tracked.
func (q *Queue) Stats() (pending, total int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	total = len(q.order)
	for _, id := range q.order {
		e := q.entries[id]
		if e == nil {
			continue
		}
		if !e.isTerminal() {
			pending++
		}
	}
	return pending, total
}

// Return a copy of args with secrets masked by the queue's Redactor, or a plain copy when it has none.
func (q *Queue) redactArgs(args []string) []string {
	if q.redactor == nil {
		return append([]string(nil), args...)
	}
	return q.redactor(args)
}

// Return the entry with the given id, or nil when no task matches.
func (q *Queue) lookup(id string) *entry {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.entries[id]
}

// Execute the task with the given id via the runner and record its final status,
// skipping one canceled before it starts; a Cancel landing mid-run wins over the runner's outcome.
func (q *Queue) run(id string) {
	e := q.lookup(id)
	if e == nil {
		return
	}

	e.mu.Lock()
	if e.status == StatusCanceled {
		e.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.status = StatusRunning
	e.startedAt = time.Now().UTC()
	out := e.out
	args := append([]string(nil), e.args...)
	redacted := e.redacted
	e.mu.Unlock()

	q.logger.Info().Str("id", id).Strs("args", redacted).Msg("Task started.")
	runErr := q.safeRun(ctx, args, out)

	e.mu.Lock()
	e.finishedAt = time.Now().UTC()
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}
	switch {
	case e.status == StatusCanceled:
		if runErr != nil && e.err == "" {
			e.err = runErr.Error()
		}
	case runErr != nil:
		e.status = StatusFailed
		e.err = runErr.Error()
	default:
		e.status = StatusSucceeded
	}
	status := e.status
	duration := e.finishedAt.Sub(e.startedAt)
	e.mu.Unlock()

	q.logger.Info().Str("id", id).Str("status", string(status)).Dur("duration", duration).Msg("Task finished.")
}

// Isolate a task's panic so it cannot bring down the worker.
func (q *Queue) safeRun(ctx context.Context, args []string, out *boundedBuffer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("task panicked: %v", r)
		}
	}()
	return q.runner(ctx, args, out)
}

// Return the error of the queue's Validator for args, or nil when it has none.
func (q *Queue) validateArgs(args []string) error {
	if q.validator == nil {
		return nil
	}
	return q.validator(args)
}

// Consume ids from the pending channel and run each task until Shutdown closes the channel.
func (q *Queue) worker() {
	defer q.wg.Done()
	for id := range q.pending {
		q.run(id)
	}
}
