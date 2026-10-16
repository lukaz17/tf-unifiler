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

package engine

import (
	"errors"
	"io"

	"github.com/tforceaio/tf-unifiler/internal/cmdqueue"
	"github.com/tforceaio/tf-unifiler/server"
)

// execViaQueue toggles suppression of interactive terminal features while
// commands are being executed through the internal command queue. The queue
// worker executes jobs sequentially, so plain reads are sufficient.
var execViaQueue = false

// queuedRunner implements cmdqueue.Runner by validating and executing
// command argument slices through Cobra within the current process.
type queuedRunner struct{}

// Validate command route by resolving it against the Cobra command tree.
func (queuedRunner) Find(args []string) error {
	if len(args) > 0 && args[0] == "serve" {
		return errors.New("command \"serve\" cannot be queued")
	}
	rootCmd := newRootCmd()
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	_, _, err := rootCmd.Find(args)
	if err != nil {
		return err
	}
	return nil
}

// Execute the command slice in-process by invoking the matching Cobra
// command without spawning a new process. Interactive terminal features are
// suppressed for the duration of the execution.
func (queuedRunner) Execute(args []string) error {
	execViaQueue = true
	defer func() { execViaQueue = false }()
	rootCmd := newRootCmd()
	rootCmd.SetArgs(args)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	return rootCmd.Execute()
}

// queueDispatcher adapts cmdqueue.Queue to the server.Dispatcher interface.
type queueDispatcher struct {
	queue *cmdqueue.Queue
}

// Enqueue command args and return the identifier of the created job.
func (d *queueDispatcher) Enqueue(args []string) (string, error) {
	job, err := d.queue.Enqueue(args)
	if err != nil {
		return "", err
	}
	return job.ID, nil
}

// Return a job view for the job with matching ID.
func (d *queueDispatcher) Job(id string) (server.JobView, bool) {
	job, ok := d.queue.Job(id)
	if !ok {
		return server.JobView{}, false
	}
	return asJobView(job), true
}

// Return job views of all queued and historical jobs.
func (d *queueDispatcher) Jobs() []server.JobView {
	jobs := d.queue.Jobs()
	views := make([]server.JobView, 0, len(jobs))
	for _, job := range jobs {
		views = append(views, asJobView(job))
	}
	return views
}

// Cancel a job which has not started yet.
func (d *queueDispatcher) Cancel(id string) bool {
	return d.queue.Cancel(id)
}

// Convert queued command job into its externally visible representation.
func asJobView(job cmdqueue.Job) server.JobView {
	return server.JobView{
		ID:          job.ID,
		Args:        job.Args,
		Status:      string(job.Status),
		SubmittedAt: job.SubmittedAt,
		StartedAt:   job.StartedAt,
		FinishedAt:  job.FinishedAt,
		Error:       job.Error,
	}
}
