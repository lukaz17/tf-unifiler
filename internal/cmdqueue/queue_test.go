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
	"sync"
	"testing"
	"time"
)

// fakeRunner implements Runner for tests. Actions are protected with mutex
// because Execute runs within the worker goroutine.
type fakeRunner struct {
	mu       sync.Mutex
	findErr  error
	execErr  error
	panics   bool
	executed [][]string
	delay    time.Duration
	wait     <-chan struct{}
}

func (f *fakeRunner) Find(args []string) error {
	return f.findErr
}

func (f *fakeRunner) Execute(args []string) error {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	if f.wait != nil {
		<-f.wait
	}
	if f.panics {
		panic("boom")
	}
	f.mu.Lock()
	f.executed = append(f.executed, append([]string(nil), args...))
	f.mu.Unlock()
	return f.execErr
}

func (f *fakeRunner) executedCalls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([][]string(nil), f.executed...)
}

func waitStatus(t *testing.T, q *Queue, id string, want Status) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if job, ok := q.Job(id); ok && job.Status == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	job, _ := q.Job(id)
	t.Fatalf("job %s did not reach status %q in time, last status %q", id, want, job.Status)
}

func TestQueueEnqueue(t *testing.T) {
	q := New(&fakeRunner{})
	defer q.Stop()

	job, err := q.Enqueue([]string{"file", "pack"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected job ID to be assigned")
	}
	if job.Status != StatusQueued {
		t.Errorf("expected status %q, got %q", StatusQueued, job.Status)
	}
}

func TestQueueEnqueueEmpty(t *testing.T) {
	q := New(&fakeRunner{})
	if _, err := q.Enqueue([]string{}); !errors.Is(err, ErrEmptyArgs) {
		t.Errorf("expected ErrEmptyArgs, got %v", err)
	}
}

func TestQueueEnqueueInvalidRoute(t *testing.T) {
	q := New(&fakeRunner{findErr: errors.New("unknown command")})
	if _, err := q.Enqueue([]string{"bogus"}); err == nil {
		t.Fatal("expected runner error to bubble up")
	}
}

func TestQueueSequentialExecution(t *testing.T) {
	runner := &fakeRunner{delay: 10 * time.Millisecond}
	q := New(runner)
	go q.Run()
	defer q.Stop()

	ids := []string{}
	for _, args := range [][]string{{"file", "pack"}, {"file", "hash"}, {"file", "rename"}} {
		job, err := q.Enqueue(args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids = append(ids, job.ID)
	}

	for _, id := range ids {
		waitStatus(t, q, id, StatusDone)
	}

	calls := runner.executedCalls()
	if len(calls) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(calls))
	}
	for i, want := range [][]string{{"file", "pack"}, {"file", "hash"}, {"file", "rename"}} {
		if calls[i][1] != want[1] {
			t.Errorf("expected execution %d to be %q, got %q", i, want[1], calls[i][1])
		}
	}
}

func TestQueueFailedJob(t *testing.T) {
	runner := &fakeRunner{execErr: errors.New("command failed")}
	q := New(runner)
	go q.Run()
	defer q.Stop()

	job, err := q.Enqueue([]string{"file", "pack"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitStatus(t, q, job.ID, StatusFailed)
	snapshot, _ := q.Job(job.ID)
	if snapshot.Error != "command failed" {
		t.Errorf("expected error message in job, got %q", snapshot.Error)
	}
}

func TestQueuePanicJob(t *testing.T) {
	runner := &fakeRunner{panics: true}
	q := New(runner)
	go q.Run()
	defer q.Stop()

	job, err := q.Enqueue([]string{"file", "pack"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitStatus(t, q, job.ID, StatusFailed)
	snapshot, _ := q.Job(job.ID)
	if snapshot.Error == "" {
		t.Error("expected panic message to be stored in job error")
	}
}

func TestQueueCancel(t *testing.T) {
	runner := &fakeRunner{delay: 10 * time.Millisecond}
	q := New(runner)

	cancelled, err := q.Enqueue([]string{"file", "rename"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.Cancel(cancelled.ID) {
		t.Fatal("expected cancel of queued job to succeed")
	}
	if q.Cancel(cancelled.ID) {
		t.Error("expected second cancel to fail")
	}
	if q.Cancel("missing") {
		t.Error("expected cancel of unknown job to fail")
	}

	go q.Run()
	defer q.Stop()

	valid, err := q.Enqueue([]string{"file", "pack"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitStatus(t, q, valid.ID, StatusDone)

	snapshot, _ := q.Job(cancelled.ID)
	if snapshot.Status != StatusCancelled {
		t.Errorf("expected job status %q, got %q", StatusCancelled, snapshot.Status)
	}
	calls := runner.executedCalls()
	if len(calls) != 1 || calls[0][1] != "pack" {
		t.Errorf("expected only the valid job to execute, got %v", calls)
	}
}

func TestQueueStopCancelsQueued(t *testing.T) {
	release := make(chan struct{})
	runner := &fakeRunner{wait: release}
	q := New(runner)
	go q.Run()

	first, err := q.Enqueue([]string{"file", "hash"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Wait until worker starts the first job.
	waitStatus(t, q, first.ID, StatusRunning)

	second, err := q.Enqueue([]string{"file", "pack"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q.Stop()

	if job, ok := q.Job(second.ID); !ok || job.Status != StatusCancelled {
		t.Errorf("expected job 2 to be cancelled, got %v", job.Status)
	}

	close(release)
	waitStatus(t, q, first.ID, StatusDone)

	if calls := runner.executedCalls(); len(calls) != 1 {
		t.Errorf("expected only the running job to execute, got %d", len(calls))
	}
}

func TestQueueEnqueueAfterStop(t *testing.T) {
	q := New(&fakeRunner{})
	go q.Run()
	q.Stop()

	if _, err := q.Enqueue([]string{"file", "pack"}); !errors.Is(err, ErrClosed) {
		t.Errorf("expected ErrClosed, got %v", err)
	}
}
