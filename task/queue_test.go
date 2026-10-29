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
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func silentLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

// Poll cond until it is true or the deadline elapses.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}

func TestSequentialOrderAtConcurrencyOne(t *testing.T) {
	var mu sync.Mutex
	order := []string{}
	release := make(chan struct{})

	q := New(func(_ context.Context, args []string, _ io.Writer) error {
		<-release
		mu.Lock()
		order = append(order, args[0])
		mu.Unlock()
		return nil
	}, 1, 8, silentLogger())
	q.Start()

	for _, name := range []string{"a", "b", "c"} {
		if _, err := q.Enqueue([]string{name}); err != nil {
			t.Fatalf("enqueue %s: %v", name, err)
		}
	}
	close(release)

	waitFor(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(order) == 3
	})
	mu.Lock()
	defer mu.Unlock()
	if order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Fatalf("expected FIFO order [a b c], got %v", order)
	}
}

func TestConcurrencyRunsInParallel(t *testing.T) {
	var running int32
	var maxSeen int32
	proceed := make(chan struct{})

	q := New(func(context.Context, []string, io.Writer) error {
		cur := atomic.AddInt32(&running, 1)
		for {
			old := atomic.LoadInt32(&maxSeen)
			if cur <= old || atomic.CompareAndSwapInt32(&maxSeen, old, cur) {
				break
			}
		}
		<-proceed
		atomic.AddInt32(&running, -1)
		return nil
	}, 3, 8, silentLogger())
	q.Start()

	for i := 0; i < 3; i++ {
		if _, err := q.Enqueue([]string{"job"}); err != nil {
			t.Fatalf("enqueue: %v", err)
		}
	}
	waitFor(t, func() bool { return atomic.LoadInt32(&maxSeen) == 3 })
	close(proceed)
}

func TestShutdownRejectsEnqueue(t *testing.T) {
	q := New(func(context.Context, []string, io.Writer) error { return nil }, 1, 4, silentLogger())
	q.Start()
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if _, err := q.Enqueue([]string{"x"}); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestEnqueueRejectsEmptyArgs(t *testing.T) {
	q := New(func(context.Context, []string, io.Writer) error { return nil }, 1, 4, silentLogger())
	if _, err := q.Enqueue(nil); !errors.Is(err, ErrEmptyArgs) {
		t.Fatalf("expected ErrEmptyArgs, got %v", err)
	}
}

func TestBacklogFull(t *testing.T) {
	block := make(chan struct{})
	q := New(func(context.Context, []string, io.Writer) error {
		<-block
		return nil
	}, 1, 1, silentLogger())
	q.Start()

	// First is picked up by the worker; second fills the backlog of 1; third overflows.
	if _, err := q.Enqueue([]string{"1"}); err != nil {
		t.Fatalf("enqueue 1: %v", err)
	}
	waitFor(t, func() bool { return len(q.pending) == 0 })
	if _, err := q.Enqueue([]string{"2"}); err != nil {
		t.Fatalf("enqueue 2: %v", err)
	}
	_, err := q.Enqueue([]string{"3"})
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	close(block)
}

func TestCancelQueuedTask(t *testing.T) {
	block := make(chan struct{})
	var ran int32

	// One worker, first task blocks so the second stays queued.
	q := New(func(_ context.Context, args []string, _ io.Writer) error {
		if args[0] == "blocker" {
			<-block
			return nil
		}
		atomic.AddInt32(&ran, 1)
		return nil
	}, 1, 8, silentLogger())
	q.Start()

	if _, err := q.Enqueue([]string{"blocker"}); err != nil {
		t.Fatalf("enqueue blocker: %v", err)
	}
	queued, err := q.Enqueue([]string{"victim"})
	if err != nil {
		t.Fatalf("enqueue victim: %v", err)
	}

	if err := q.Cancel(queued.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	close(block)

	waitFor(t, func() bool {
		got, _ := q.Get(queued.ID)
		return got.Status == StatusCanceled
	})
	if atomic.LoadInt32(&ran) != 0 {
		t.Fatal("canceled queued task should not have run")
	}
}

func TestCancelRunningTaskSignalsContext(t *testing.T) {
	started := make(chan struct{})
	var observedCancel int32

	q := New(func(ctx context.Context, _ []string, _ io.Writer) error {
		close(started)
		<-ctx.Done() // cooperative: the runner honors the context here.
		atomic.StoreInt32(&observedCancel, 1)
		return ctx.Err()
	}, 1, 8, silentLogger())
	q.Start()

	tsk, err := q.Enqueue([]string{"long"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	<-started
	if err := q.Cancel(tsk.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	waitFor(t, func() bool { return atomic.LoadInt32(&observedCancel) == 1 })
	waitFor(t, func() bool {
		got, _ := q.Get(tsk.ID)
		return got.Status == StatusCanceled
	})
}

func TestCancelUnknownTask(t *testing.T) {
	q := New(func(context.Context, []string, io.Writer) error { return nil }, 1, 4, silentLogger())
	if err := q.Cancel("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestClearRemovesTerminalOnly(t *testing.T) {
	q := New(func(context.Context, []string, io.Writer) error { return nil }, 1, 8, silentLogger())
	q.Start()

	tsk, _ := q.Enqueue([]string{"done"})
	waitFor(t, func() bool {
		got, _ := q.Get(tsk.ID)
		return got.Status == StatusSucceeded
	})
	if n := q.Clear(); n != 1 {
		t.Fatalf("expected 1 cleared, got %d", n)
	}
	if _, ok := q.Get(tsk.ID); ok {
		t.Fatal("cleared task should be gone")
	}
}

func TestStatsCountsPendingAndTotal(t *testing.T) {
	block := make(chan struct{})
	// One worker: the first task runs and blocks, the rest stay queued.
	q := New(func(_ context.Context, args []string, _ io.Writer) error {
		if args[0] == "blocker" {
			<-block
		}
		return nil
	}, 1, 8, silentLogger())
	q.Start()

	if _, err := q.Enqueue([]string{"blocker"}); err != nil {
		t.Fatalf("enqueue blocker: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := q.Enqueue([]string{"waiting"}); err != nil {
			t.Fatalf("enqueue waiting: %v", err)
		}
	}

	// 3 enqueued, none terminal yet: 1 running + 2 queued.
	waitFor(t, func() bool {
		pending, total := q.Stats()
		return pending == 3 && total == 3
	})

	close(block)

	// Once all drain to terminal, nothing is pending; total stays until cleared.
	waitFor(t, func() bool {
		pending, _ := q.Stats()
		return pending == 0
	})
	if _, total := q.Stats(); total != 3 {
		t.Fatalf("expected total 3 after completion, got %d", total)
	}
}

func TestFailedTaskRecordsError(t *testing.T) {
	q := New(func(context.Context, []string, io.Writer) error {
		return errors.New("boom")
	}, 1, 4, silentLogger())
	q.Start()

	tsk, err := q.Enqueue([]string{"x"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	waitFor(t, func() bool {
		got, _ := q.Get(tsk.ID)
		return got.Status == StatusFailed
	})
	got, _ := q.Get(tsk.ID)
	if got.Err != "boom" {
		t.Fatalf("expected error 'boom', got %q", got.Err)
	}
}

func TestOutputIsCaptured(t *testing.T) {
	q := New(func(_ context.Context, _ []string, out io.Writer) error {
		_, _ = io.WriteString(out, "hello output")
		return nil
	}, 1, 4, silentLogger())
	q.Start()

	tsk, err := q.Enqueue([]string{"x"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	waitFor(t, func() bool {
		got, _ := q.Get(tsk.ID)
		return got.Status == StatusSucceeded
	})
	got, _ := q.Get(tsk.ID)
	if got.Output != "hello output" {
		t.Fatalf("expected captured output, got %q", got.Output)
	}
}

func TestPanicInRunnerIsIsolated(t *testing.T) {
	q := New(func(context.Context, []string, io.Writer) error {
		panic("kaboom")
	}, 1, 4, silentLogger())
	q.Start()

	tsk, err := q.Enqueue([]string{"x"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	waitFor(t, func() bool {
		got, _ := q.Get(tsk.ID)
		return got.Status == StatusFailed
	})
	// The worker must still process further tasks.
	tsk2, err := q.Enqueue([]string{"y"})
	if err != nil {
		t.Fatalf("enqueue after panic: %v", err)
	}
	waitFor(t, func() bool {
		got, _ := q.Get(tsk2.ID)
		return got.Status == StatusFailed
	})
}
