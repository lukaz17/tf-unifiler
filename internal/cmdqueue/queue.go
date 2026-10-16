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

package cmdqueue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrEmptyArgs returned when submitted command contains no arguments.
	ErrEmptyArgs = errors.New("args must contain at least one element")
	// ErrClosed returned when the queue no longer accepts new jobs.
	ErrClosed = errors.New("command queue has been closed")
)

// Queue stores submitted jobs in memory and executes them sequentially in
// submission order using a single worker goroutine.
type Queue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	runner Runner
	jobs   []*Job
	byID   map[string]*Job
	closed bool
	done   chan struct{}
}

// Entrypoint for creating a new Queue. The worker is not started yet, spawn
// it via Run in a goroutine.
func New(runner Runner) *Queue {
	q := &Queue{
		runner: runner,
		jobs:   []*Job{},
		byID:   map[string]*Job{},
		done:   make(chan struct{}),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue validates and registers a new job. Returns ErrEmptyArgs when args
// is empty, a runner error when the command route is invalid, and ErrClosed
// once the queue has been stopped.
func (q *Queue) Enqueue(args []string) (*Job, error) {
	if len(args) == 0 {
		return nil, ErrEmptyArgs
	}
	if err := q.runner.Find(args); err != nil {
		return nil, err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil, ErrClosed
	}
	job := &Job{
		ID:          uuid.NewString(),
		Args:        args,
		Status:      StatusQueued,
		SubmittedAt: time.Now().UTC(),
	}
	q.jobs = append(q.jobs, job)
	q.byID[job.ID] = job
	q.cond.Signal()
	return job, nil
}

// Return a snapshot copy of the job with matching ID.
func (q *Queue) Job(id string) (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	job, ok := q.byID[id]
	if !ok {
		return Job{}, false
	}
	return *job, true
}

// Return a snapshot copy of all jobs, newest first.
func (q *Queue) Jobs() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	jobs := make([]Job, 0, len(q.jobs))
	for i := len(q.jobs) - 1; i >= 0; i-- {
		jobs = append(jobs, *q.jobs[i])
	}
	return jobs
}

// Cancel a job which has not started yet. Returns false when the job does
// not exist or is already running/finished.
func (q *Queue) Cancel(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	job, ok := q.byID[id]
	if !ok || job.Status != StatusQueued {
		return false
	}
	job.Status = StatusCancelled
	job.FinishedAt = time.Now().UTC()
	return true
}

// Run consumes and executes queued jobs sequentially until Stop() is called
// and all remaining work is settled. Intended to be wrapped in a goroutine.
func (q *Queue) Run() {
	defer close(q.done)
	for {
		q.mu.Lock()
		var job *Job
		for job == nil && !q.closed {
			job = q.nextQueued()
			if job == nil {
				q.cond.Wait()
			}
		}
		if job == nil {
			// Queue is closed, and no queued jobs are remaining.
			q.mu.Unlock()
			return
		}
		job.Status = StatusRunning
		job.StartedAt = time.Now().UTC()
		q.mu.Unlock()

		err := q.execute(job.Args)

		q.mu.Lock()
		job.FinishedAt = time.Now().UTC()
		if err != nil {
			job.Status = StatusFailed
			job.Error = err.Error()
		} else {
			job.Status = StatusDone
		}
		q.mu.Unlock()
	}
}

// Stop marks queue as closed. Jobs which have not started yet are marked as
// cancelled, the currently running job is allowed to complete.
func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	for _, job := range q.jobs {
		if job.Status == StatusQueued {
			job.Status = StatusCancelled
			job.FinishedAt = time.Now().UTC()
		}
	}
	q.cond.Broadcast()
}

// Wait until the worker exited, or timeout elapsed. Call Stop() before this
// method, otherwise the worker will never exit.
func (q *Queue) Wait(timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-q.done:
	case <-timer.C:
	}
}

// Return first queued (not yet executed) job in submission order.
// Caller must hold q.mu.
func (q *Queue) nextQueued() *Job {
	for _, job := range q.jobs {
		if job.Status == StatusQueued {
			return job
		}
	}
	return nil
}

// Execute the runner with panic recovery.
func (q *Queue) execute(args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred during command execution: %v", r)
		}
	}()
	return q.runner.Execute(args)
}
